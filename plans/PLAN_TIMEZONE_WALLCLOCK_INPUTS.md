# Plan + Report: Timezone Correctness — Wall-Clock Input Points

> **Supersedes** `PLAN_TIMEZONE.md` Fase 7a (Event) for the anchor choice — see
> [Relationship to PLAN_TIMEZONE.md](#relationship-to-plan_timezonemd) below.
> Fase 7b (DND) in that document remains the intended design and is
> reproduced here as-is (not yet implemented).

## Context

Investigation triggered by a report that "all time input is stored in
GMT+7." Live query against the dev Postgres instance disproved this for
instant timestamps: `SHOW timezone` = `UTC`, and
`accounting_categories.created_at` stored as `2026-06-18 01:57:43+00` —
exactly matching the `+0700` value originally reported (01:57 + 7h =
08:57), proving the discrepancy was a **display artifact** of whatever DB
client was used to view the row, not a storage bug. `created_at`/`updated_at`
and all other `TIMESTAMPTZ` columns are confirmed correct and are **out of
scope** — no code changes needed there.

The real problem is narrower: 12 `DATE` columns across 8 tables, each
combined with a separate wall-clock time value (or used as an "is this
active today" boundary) with no explicit timezone attached. Each was
classified along two axes — **input** (how the value is interpreted when
written) and **output** (how it's re-interpreted when read/compared) — into
one of three categories:

- **ABSOLUTE** — a recorded fact with no "now" comparison (`birth_date`,
  `refund_date`, `effective_date`, `transaction_date`,
  `subscriptions.current_period_start/end`). No ambiguity, no fix needed.
- **WALLCLOCK-OWNER** — belongs to one specific entity (an event, a
  merchant, a schedule entry). Anchor: an **explicit timezone field stored
  on that entity** — never a device inferred silently, never the global
  `default_timezone` setting as primary source. A device's current zone may
  *prefill* the field in the UI, but the field itself is always sent
  explicitly and confirmable. Fallback for old data with no field set:
  `default_timezone` → hardcoded `Asia/Jakarta`.
- **WALLCLOCK-PLATFORM** — no individual owner (Ads "is this campaign
  active today"). Anchor: `default_timezone` directly, no per-row field.

**Exception**: DND (notification quiet hours) is not a location case — it's
"what time is it right now where this device physically is," so the
recipient's own device timezone is the *correct* source, not a proxy for
something else. No explicit field needed there; see Fase 7b below.

## Scope

| Module | Category | Status |
|---|---|---|
| Event (`event_date`+`event_time`) | WALLCLOCK-OWNER | **DONE** (this doc) — see below |
| DND (notification quiet hours) | WALLCLOCK-OWNER (device, not field) | **DONE** (2026-07-17) — see [Fase 7b](#fase-7b-dnd--done-2026-07-17) |
| Community schedule occurrence (`community_schedule_entries`) | WALLCLOCK-OWNER | **DONE** (2026-07-17) — see [Fase 7c](#fase-7c-community-schedule-occurrence--done-2026-07-17) |
| Ads (`start_date`/`end_date` "active today") | WALLCLOCK-PLATFORM | **DONE** (2026-07-20) — see [Fase 7d](#fase-7d-ads--done-2026-07-20) |
| Directory business hours (JSONB, per-merchant) | WALLCLOCK-OWNER | **NOT STARTED** |
| `created_at`/`updated_at`/other instants | ABSOLUTE | Out of scope, verified correct |
| `birth_date`, `refund_date`, `effective_date`, `transaction_date`, subscription periods | ABSOLUTE | Out of scope, verified correct |

---

## Event — DONE (2026-07-17)

**Design decision** (explicit choice made over chat, deviating from
`PLAN_TIMEZONE.md` Fase 7a's original draft — see next section): the
organizer **must manually select** an IANA timezone when creating/editing an
event; it is **not** silently inferred from their device. The mobile client
may prefill the dropdown from the device's detected zone for convenience,
but the field is mandatory in the request body regardless, and in edit mode
prefills from the event's own already-stored value, not the device again.

### Backend (`k-forum-api`) — implemented, tested, migrated on dev DB

- **Migration** `20260717040308_add_event_timezone_and_utc_instants`:
  `events.timezone VARCHAR(64) NOT NULL DEFAULT 'Asia/Jakarta'`,
  `events.starts_at_utc TIMESTAMPTZ`, `events.ends_at_utc TIMESTAMPTZ` +
  index on `starts_at_utc`. Backfilled existing rows by combining
  `event_date`+`event_time` as `Asia/Jakarta` (the assumption already baked
  into the app before this change — not a regression). Applied to dev DB;
  verified via direct query that `event_time="19:04"` → `starts_at_utc`
  shows `12:04:00+00` (= 19:04 WIB), confirming the backfill math.
- **Domain** (`internal/domain/event/entity/event.go`): `Event` gained
  `Timezone string`, `StartsAtUTC time.Time`, `EndsAtUTC *time.Time`.
  `applyFields()` validates `timezone` via `time.LoadLocation` (new domain
  errors `CodeEventTimezoneRequired`/`CodeEventTimezoneInvalid`, added to
  `locales/{id,en,ko}.json`) and computes `StartsAtUTC`/`EndsAtUTC` from
  `event_date`+`event_time`+the validated location — a new private helper
  `combineDateAndTimeInLocation`, generalized from the old
  Jakarta-only `combineDateAndTime`. All 4 mutation entry points
  (`NewEvent`, `NewAdminEvent`, `Update`, `AdminUpdate`) now take a
  `timezone string` parameter.
- **DTOs** (`internal/app/dto/event_dto.go`): `EventInput.Timezone`
  (`binding:"required"`), `EventListItem.Timezone`,
  `EventDetailResponse.Timezone` — all four usecases that build these
  (create, admin-create, update, admin-update) and all read paths (list,
  detail, my-events, saved, schedule, featured, upcoming, search) updated to
  thread the field through, including two call sites in
  `postgres_event_query.go` (`ListMySavedEvents`, `ListMySchedule`) that had
  their own inline `rows.Scan` duplicating the shared `scanEventListRow`
  helper — both updated to keep column count in sync with the new SELECT.
- **Removed**: `jakartaLocation` package var and `combineDateAndTime()` in
  `internal/app/usecase/event/helpers.go` — no longer needed once every
  event carries its own validated timezone.
- **`schedule_event.go`** (reminder job) and **`get_calendar_export.go`**
  (ICS/Google/Outlook export) now read `ev.StartsAtUTC`/`ev.EndsAtUTC`
  directly instead of recombining date+time+hardcoded-zone at read time —
  purely instant arithmetic from here on, no more zone-guessing at runtime.
  Calendar export still renders wall-clock digits in the event's own zone
  (via `ev.StartsAtUTC.In(loc)`) so exported links show the organizer's
  intended local time, not raw UTC.
- **Bug fixed in the same pass**: `ListUpcomingEvents`
  (`postgres_event_query.go`) was comparing `event_date >= CURRENT_DATE` —
  a timezone-naive `DATE` against Postgres's session `TimeZone` GUC — never
  wired to any event-specific zone at all, unlike the rest of the Event
  module. Changed to `starts_at_utc >= NOW()`, a pure instant comparison
  with no zone ambiguity left.

### Mobile app (`k_forum`) — done, verified

- `lib/features/event/domain/entities/create_event_data.dart`: `timezone`
  is now a required constructor field, always emitted in `toJson()`.
  `fromEvent()` (used to prefill the edit form) reads `e.timezone ?? 'Asia/Jakarta'`.
- `lib/features/event/presentation/screens/event_form_screen.dart`: new
  mandatory timezone dropdown (`_buildTimezoneDropdown`), matching the
  existing category-dropdown visual pattern, with a required validator and
  the same `' *'` label convention used elsewhere in the form. New events
  prefill from the device's detected zone; edit mode prefills from the
  event's own stored value; both resolve through `_resolveTimezone()`,
  which falls back to `Asia/Jakarta` if the candidate is null or not one of
  the static dropdown options (Asia/Jakarta, Makassar, Jayapura, Singapore,
  Kuala Lumpur, Bangkok, Manila, Tokyo, Seoul, Sydney, London, UTC).
- **No new package needed** — `flutter_timezone` (`^5.1.0`) was already a
  dependency (added in an earlier, unrelated commit `72205dc` for FCM device
  registration) and already wrapped by `DeviceInfoService.getTimezone()`
  (`lib/core/services/device_info_service.dart`); the form reuses that
  existing service rather than introducing a second detection path.
- `lib/l10n/app_{en,id,ko}.arb` gained `eventFormTimezone`; the validator
  reuses the already-existing `eventFieldRequired` string (same one other
  required fields in this form already use).
- `test/features/event/event_mutation_result_test.dart` updated for the new
  required constructor field (the only other call site besides the form).
- **Verified independently** (not just taking the implementing agent's word
  for it): `flutter analyze lib/features/event/` → "No issues found!";
  reviewed the actual `git diff` for every changed file line-by-line.
- **Known limitation, accepted**: events created before this change have
  `timezone = NULL` at the DB row level pre-migration-backfill (backfilled
  to `Asia/Jakarta` by the migration, see above) — edit mode for such an
  event will show `Asia/Jakarta` without flagging "this was assumed, please
  confirm." Acceptable per the same backfill assumption used everywhere
  else in this change; revisit only if it causes real complaints.

### API Spec

`K-FORUM-DOCS/API SPEC/Mobile/API_SPEC_EVENT_MOBILE.md` and
`K-FORUM-DOCS/API SPEC/Web/API_SPEC_EVENT_BACKOFFICE.md` updated — `timezone`
added to every create/edit request example and validation list, every
list/detail/schedule response example, a new DO/DON'T bullet flagging this
as a breaking change dated 2026-07-17, and both error codes added to the
error-scenario tables. Verified against the current Go DTOs directly (not
just taken on faith) before writing.

**Pre-existing doc/code drift noticed while updating these files (flagged,
not fixed — unrelated to this change)**:
- Mobile `#17 Get My Schedule` spec still shows a flat response object;
  the actual `dto.EventScheduleResponse` nests event fields under an
  `"event"` key (see `EventScheduleResponse`'s own doc comment in
  `event_dto.go`, and [[project_event_module_bugs]] round 4 for why).
- Web `#2 List All Events` / `#3 Get Pending Approvals` examples omit
  `event_time` even though the embedded `EventListItem` always returns it.
- Mobile `#13 Get Calendar Export Links` response example shows
  `title`/`event_date`/`event_time` fields that don't exist on the real
  `CalendarExportResponse` DTO (which only has 3 URL fields).

### Verification

- `go build ./...` / `go vet ./...` clean.
- Full `go test ./internal/interfaces/http/handler/mobile/...` (85s) and
  `.../web/...` (43s) — all green, no regressions, including the pre-existing
  `TestMobileEvent_ScheduleEvent_ReminderJobUsesCorrectTimezone` and
  `TestMobileEvent_GetCalendarExport_CorrectTimezone` from the earlier
  (2026-07-07) WIB-hardcode bugfix round.
- New tests added: `TestMobileEvent_CreateEvent_InvalidTimezone` (bad IANA
  string → 422) and `TestMobileEvent_CreateEvent_TimezoneAnchorsStartsAtUTC`
  (event created with `Asia/Seoul` + `"10:00"` → asserts `starts_at_utc` =
  `2026-12-01 01:00:00+00`, i.e. correctly UTC+9 — not the UTC+7 a stale
  Jakarta assumption would have produced). Both pass.
- Migration applied to dev DB directly (`go run ./cmd/migrate up` with
  `DB_HOST=localhost DB_PORT=5434` to bypass the docker-network hostname in
  `.env`) and spot-checked via live query.

### Relationship to PLAN_TIMEZONE.md

`PLAN_TIMEZONE.md`'s Fase 7a drafted a different anchor for Event: the
*organizer's device timezone*, captured automatically at submit time (with
`Asia/Jakarta` fallback), no explicit field. That draft was superseded
before implementation began, for a concrete reason surfaced during design
review: device timezone is a proxy for "where is the organizer's phone
right now," not "where does the event happen" — an organizer traveling
abroad while creating an event back home would have mislabeled the event's
zone under the device-inferred design. The mandatory-explicit-field design
implemented here avoids that failure mode at the cost of one extra form
field. Fase 7a's migration shape (additive `starts_at_utc`/`ends_at_utc`,
backfill via `Asia/Jakarta`, read paths switched to the new columns) was
otherwise directly reused — only the *source* of the anchor changed.

---

## Fase 7b — DND — DONE (2026-07-17)

**Principle**: the DND window (`"22:00–08:00"`) is a recurring wall-clock
rule, not an instant — the stored fields stay plain `"HH:MM"` strings,
**not** converted to UTC. Only the zone used to answer "what time is it
right now" became dynamic per-recipient instead of hardcoded Jakarta.

### What changed

- `internal/domain/notification/service/module_preference_gate.go`:
  removed the exported `NowInJakarta()` (nothing calls it anymore).
  `DefaultModulePreferenceGate` now takes a second constructor arg
  `notifRepo.DeviceRegistrationRepository` (legal per CLAUDE.md — domain
  services may depend on domain repository interfaces) and exposes
  `NowInUserTimezone(ctx, userID, defaultTimezone string) time.Time`:
  looks up the user's most-recently-active device
  (`FindActiveByUserID`, already ordered `last_seen_at DESC`) → its
  `Timezone` if set → falls back to `defaultTimezone` (passed in by the
  caller, since domain services cannot depend on
  `port.SystemSettingsProvider`) → falls back to `Asia/Jakarta`.
- `resolveTimezoneLocation(devices, defaultTimezone)` is the shared private
  helper both `NowInUserTimezone` and the bulk path use.
- **`ModulePreferenceGate` interface signature changed** (both
  implementations + the free function updated): `FilterChannels(ctx,
  userID, channels, now time.Time)` → `FilterChannels(ctx, userID,
  channels, defaultTimezone string)` — the gate now resolves `now` itself
  internally instead of receiving a pre-computed instant. Same for
  `IsAllowedBulk(..., channel, defaultTimezone string)` and the free
  function `FilterAllowedUsers(..., defaultTimezone string)`.
- **`IsAllowedBulk` restructured to per-user** (previously computed
  `now := NowInJakarta()` once for the whole batch): now bulk-fetches
  devices via the existing `DeviceRegistrationRepository.FindActiveByUserIDs`
  — **but only when `channel == push`** (DND never blocks any other
  channel, so in_app/email/sms fanout calls skip the device fetch
  entirely — verified by a test asserting the fake device repo's bulk
  method is called exactly once for `push` and zero times for `in_app`).
  `postgres_device_repository.go`'s `FindActiveByUserIDs` query gained
  `ORDER BY user_id, last_seen_at DESC NULLS LAST, created_at DESC` (it had
  no ordering at all before) so `devicesByUser[uid][0]` is deterministically
  the most-recently-active device per user — required for this to be
  correct, not just for single-user lookups.
- **`FilterChannels` also skips the device fetch** when the requested
  channel list doesn't include push at all (DND is a no-op for those calls
  regardless of timezone).
- **Where `defaultTimezone` comes from**: `Dispatcher`
  (`internal/app/service/notification/dispatcher.go`) and
  `BroadcastFanoutRelay` (`internal/interfaces/mq/relay/broadcast_fanout_relay.go`)
  each gained a `sysSettings port.SystemSettingsProvider` field. A new
  exported helper `ResolveDefaultTimezone(ctx, provider)` lives in
  `internal/app/service/notification` (same package as `dispatcher.go`,
  called unqualified there; imported as `notifdispatch` in the relay,
  which already had that import for `DeactivateInvalidTokens`) and reads
  `default_timezone` via `GetAll`
  (that setting is `is_public=false`, unlike `default_language` which uses
  `GetPublic`) with `""` fallback on any error — the domain gate is the
  single place that then falls further back to `Asia/Jakarta`, so there's
  only one hardcoded-fallback definition in the whole chain.
- **`cmd/worker/main.go` gained its first-ever Redis connection.** The
  worker process previously had zero Redis usage; `SystemSettingsProvider`
  needs the same Redis-cached decorator pattern `cmd/app/main.go` uses
  (`cache.NewCachedSystemSettingsProvider`, 5 min TTL). Operationally safe —
  `docker-compose.dev.yml`'s `worker` service already `depends_on: redis:
  condition: service_healthy` and shares `.env`'s `REDIS_ADDR`, so this was
  anticipated infra, just never actually connected to before.

### Verification

- `go build ./...` / `go vet ./...` clean across the whole repo.
- New domain-level tests (`module_preference_gate_test.go`, no HTTP/DB
  harness needed — this package has zero dependency on
  `testhelper.TestServer`, confirmed no RabbitMQ/dispatcher coverage exists
  at the handler-test level either, matching the pre-existing gap noted in
  [[project_event_module_bugs]] round 3):
  - `TestNowInUserTimezone_FollowsMostRecentDevice` — device `Asia/Seoul`
    → offset is UTC+9, not frozen WIB.
  - `TestNowInUserTimezone_FallbackChain` — all 4 links of the fallback
    chain individually (no device + no default → Jakarta; no device +
    valid default → default; device with `Timezone: nil` → default;
    invalid default string → Jakarta, no panic).
  - `TestIsAllowedBulk_PerUserTimezoneNotFrozenToJakarta` — a DND window
    computed to cover "now in Jakarta" does NOT block a Seoul-device user
    whose actual local time falls outside that window — this is the
    concrete regression the old single-`now`-for-the-batch design would
    have failed.
  - `TestIsAllowedBulk_DeviceFetchIsSingleBatchQuery` — asserts the fake
    device repo's bulk method is called exactly once across 3 users, and
    zero times for a non-push channel.
  - `TestFilterChannels_DNDUsesRecipientDeviceTimezone` — 4 sub-cases
    (device-zone DND active, no-device falls back to defaultTimezone,
    non-push channel list skips DND entirely, `all_enabled=false`
    overrides everything).
  - Full existing `TestIsAllowedBulk_BroadcastRespectsGlobalDisableAndDND`
    suite still passes unchanged in behavior (empty `defaultTimezone` + no
    device in the fake repo → same Asia/Jakarta-equivalent behavior as
    before).
- Full repo test suite (`go test ./...`, DB-backed packages included) green,
  no regressions — in particular `internal/infrastructure/persistence`
  (covers the `ORDER BY` addition) and full mobile+web handler suites.

### API Spec

`K-FORUM-DOCS/API SPEC/Mobile/API_SPEC_NOTIFICATION_PREFERENCES.md` — added
a note explaining DND has no explicit timezone field (device-based,
fallback chain), contrasted with Event's mandatory explicit field.
**Pre-existing drift flagged, not fixed**: `API_SPEC_FCM.md` in the same
directory documents a `fcm_tokens` table with only
`fcm_token`/`device_id`/`platform` — the real table is
`device_registrations`, which has had a `timezone` column (among others)
since 2026-07-10; that spec file needs its own separate revision pass.

## Fase 7c — Community schedule occurrence — DONE (2026-07-17)

**Confirmed live, tested, routed code** — not dead/experimental (verified
before implementing, since an earlier research pass had wrongly suggested
the recurrence engine "seems not fully implemented"): 11 mobile routes +
web moderation routes wired in `router.go`, 11 passing handler tests in
`community_schedule_handler_test.go` before this change even started.

**Bug — more subtle than Event's, and NOT the same shape**: unlike Event
(`event_date` DATE + `event_time` VARCHAR, no zone info at all — combined
server-side), community schedule's `start_at` is a single client-supplied
`TIMESTAMPTZ`. The client's JSON *does* carry a real offset
(`"2026-07-20T09:00:00+07:00"`) at the moment of the create request — but
that offset is **discarded on every subsequent read**: Postgres normalizes
`TIMESTAMPTZ` storage to UTC internally, and pgx always scans it back with
`Location() == UTC`, regardless of the original offset. Every place that
called `entry.StartAt.Location()` to reconstruct occurrence times or
weekdays was therefore silently wrong for any entry whose local time
crosses a UTC midnight boundary (e.g. 03:00 WIB = 20:00 UTC the *previous*
day) — both the reconstructed hour AND, for `WEEKLY`/`MONTHLY` recurrence
with no explicit `BYDAY`, the very weekday/day-of-month the series was
anchored to.

**Getting the fix right required more care than a naive
`s/entry.StartAt.Location()/entry.TimezoneLocation()/`** — an early draft
of this fix (caught before landing) made exactly that swap while still
reading `entry.StartAt.Hour()` directly, which reads the *UTC* hour and
would have relabeled it as if it were already the local hour, introducing
a fresh 7-hour error strictly worse than the original bug. The correct
fix converts the instant into the entry's zone first
(`entry.StartAt.In(loc)`) and *then* reads wall-clock components from that.

### What changed

- **Migration** `20260717080434_add_community_schedule_entry_timezone`:
  `community_schedule_entries.timezone VARCHAR(64) NOT NULL DEFAULT
  'Asia/Jakarta'` (backfill, same rationale as Event/DND — the assumption
  already baked into the app before this fix, not a regression).
- **Domain** (`internal/domain/community/entity/community_schedule_entry.go`):
  `Timezone string` field, validated in `applyFields` via
  `time.LoadLocation` (new domain errors
  `CodeScheduleEntryTimezoneRequired`/`CodeScheduleEntryTimezoneInvalid`).
  New method `TimezoneLocation()` resolves it with an Asia/Jakarta fallback
  (only reachable for pre-backfill rows or missing tzdata) — this is the
  one place that owns the fallback, so no other file duplicates it.
- **`internal/domain/community/service/schedule_service.go`**:
  `OccurrenceStartAt()` now does `localStartAt := entry.StartAt.In(loc)`
  before reading `.Hour()/.Minute()/.Second()`, where
  `loc := entry.TimezoneLocation()` — this is the correction described
  above. `IsOccurrencePast()`'s all-day end-of-day boundary also switched
  to `entry.TimezoneLocation()`.
- **`internal/app/usecase/community/occurrence_expansion.go`**:
  `expandOccurrenceDates()` gained a `loc *time.Location` parameter; a new
  `truncateDateInLocation(t, loc)` reads calendar Y/M/D from `t.In(loc)`
  (the entry's own zone) instead of blindly trusting whatever Location `t`
  happens to carry — this is what fixes the RRULE weekday/day-of-month
  anchor bug for early-morning entries. `from`/`to` (the query window
  bounds) are untouched — they're plain calendar dates parsed with no
  zone ambiguity of their own, nothing to fix there.
- **Three call sites** needed threading `Timezone` into a manually-built
  `communityentity.CommunityScheduleEntry{}` (read-model → domain-shaped
  struct, only fields needed for the calculation) and passing the
  resolved location into `expandOccurrenceDates`: `list_occurrences.go`,
  `community_activity_preview.go` (the "upcoming_schedule" preview feature
  from [[project_community_activity_preview]], 2026-07-16 — shares this
  exact expansion logic and was equally affected). `get_occurrence_detail.go`
  needed no changes — it operates on the full entity from `entryRepo.FindByID`
  (which now carries `Timezone` automatically), not a manually-built one.
- **DTO/usecase**: `CommunityCreateScheduleEntryInput.Timezone string`
  (mandatory), `CommunityUpdateScheduleEntryInput.Timezone *string`
  (optional-with-fallback, same pattern as every other editable field in
  `edit_schedule_entry.go`). `CommunityScheduleEntryItem`/
  `CommunityScheduleEntryReadItem` (port) both gained `Timezone` — consumed
  by the create/edit response and the `GetScheduleEntryDetail` read path
  (`get_schedule_entry_detail.go`), needed so an edit form can prefill the
  same timezone the entry already has. `CommunityOccurrenceItem` (the
  occurrence list/detail response) deliberately did **not** get a
  `Timezone` field — `StartAt`/`EndAt` there are already correct absolute
  instants once the bug above is fixed, and Go's `time.Time` JSON
  marshaling already includes the correct offset; adding a redundant raw
  string field wasn't justified.
- **Persistence**: `postgres_community_announcement_schedule_repository.go`
  (Save/Update/FindByID + scan) and `postgres_community_announcement_schedule_query.go`
  (`ListScheduleEntriesInWindow`, `ListActiveScheduleEntriesInWindowForCommunities`,
  `GetScheduleEntryDetail` — all three share one SELECT column-list fragment
  and one scan helper, updated once for all three). `CommunityScheduleAdminReadItem`
  (backoffice admin list) deliberately **not** touched — admin listing just
  displays `start_at` raw, doesn't compute occurrences, out of scope.

### Verification

- New tests, both with an explicit **premise self-check** (`require.Equal`
  asserting the UTC-vs-local weekday/day actually differs before asserting
  the fix), so the test can't pass vacuously if the bug scenario turns out
  not to reproduce as constructed:
  - `internal/domain/community/service/schedule_service_test.go`:
    `TestOccurrenceStartAt_CorrectlyReconstructsLocalHourAfterUTCRoundTrip`
    (03:00 WIB entry, simulated post-round-trip UTC label, both the
    same-day and a later recurring occurrence reconstruct to 03:00 WIB —
    not 03:00 UTC), `TestOccurrenceStartAt_FallsBackToJakartaWhenTimezoneEmpty`,
    `TestIsOccurrencePast_AllDayUsesEntryTimezoneEndOfDay` (Seoul UTC+9
    all-day boundary).
  - `internal/app/usecase/community/occurrence_expansion_test.go`:
    `TestExpandOccurrenceDates_WeeklyDefaultUsesCreatorLocalWeekday` (a
    Monday-WIB entry whose UTC instant falls on Sunday — every generated
    occurrence must still land on Monday), `TestExpandOccurrenceDates_OneOff_UsesLocalCalendarDate`.
- Existing 11 `community_schedule_handler_test.go` tests + activity-preview
  handler tests still pass unchanged (test payloads updated to include the
  new mandatory `timezone` field).
- Full repo test suite green — run both isolated per-package and as a full
  sequential (`-p 1`) `go test ./...` pass, since an initial parallel full-suite
  run produced widespread unrelated failures that turned out to be Docker/
  testcontainers resource contention from running many packages' dedicated
  Postgres containers simultaneously, not a real regression (confirmed by
  re-running the same "failing" tests in isolation — all passed).
- Migration applied to dev DB, confirmed via live query.

### API Spec

`K-FORUM-DOCS/API SPEC/Mobile/API_SPEC_COMMUNITY_MOBILE.md` — added a note
to the `upcoming_schedule` preview field table.

**Addendum (2026-07-17, later same day — triggered by the user asking
"does mobile/backoffice spec need adjusting")**: the note above originally
claimed the actual create/edit/RSVP/occurrence endpoints have "no spec
documentation in API SPEC at all" — that overstated the gap. A real,
route-matching spec **does** exist for all 11 mobile routes + all 3 web
moderation routes, just filed under `K-FORUM-DOCS/Modules/Community/`
instead of the top-level `API SPEC/Mobile/`/`API SPEC/Web/` folders:
- `Modules/Community/API_SPEC_COMMUNITY_ANNOUNCEMENT_SCHEDULE_MOBILE.md` (B1–B11)
- `Modules/Community/API_SPEC_COMMUNITY_ANNOUNCEMENT_SCHEDULE_BACKOFFICE.md` (B1–B3)
- `Modules/Community/COMMUNITY_ANNOUNCEMENT_SCHEDULE_RULES.md` and
  `COMMUNITY_ANNOUNCEMENT_SCHEDULE_DB_SCHEMA.md` (business-rule/schema
  source-of-truth docs, distinct from the client-facing API contract docs)

All four were stale (zero mention of `timezone` despite it existing in the
real migration/entity) and have now been updated: Schedule Entry Object
example gains `"timezone"`, B4 (create) documents it as mandatory, B5
(edit) documents it as **optional with fallback** — a real asymmetry
found in the DTOs (`CommunityCreateScheduleEntryInput.Timezone string
binding:"required"` vs `CommunityUpdateScheduleEntryInput.Timezone
*string`, no `required` tag) that the docs needed to reflect accurately
rather than copying Event's "mandatory on both create and edit" wording.
Both new domain error codes added to the backoffice/mobile error tables.
RULES.md gained a `timezone` row in the entity table plus a note in
§Recurrence & Occurrence explaining the bug this fixed. DB_SCHEMA.md
gained the DDL column, the Go struct mirror field, a migration-reference
entry pointing at the real migration file, and a note in its own
§Catatan Occurrence & Recurrence.

**A genuine code gap, not just a doc gap, found in the same pass**: the
backoffice moderation response (`CommunityScheduleAdminItem`, used by
both `ListScheduleWeb` and `GetScheduleDetailWeb`) had **no `Timezone`
field at all** — an admin moderating a schedule entry couldn't see what
timezone it was in. User chose to fix this in code rather than just
document the gap: added `Timezone` to `CommunityScheduleAdminItem` (DTO),
`CommunityScheduleAdminReadItem` (port), the shared SELECT/scan in
`ListScheduleForAdmin`/`GetScheduleDetailForAdmin`
(`postgres_community_announcement_schedule_query.go`), and both usecases
(`list_schedule_web.go`, `get_schedule_detail_web.go`). Existing
`TestWebCommunitySchedule_ListAndGetDetail_Success` extended with an
assertion that the detail response's `timezone` field is
`"Asia/Jakarta"` (the column default, since the test's fixture inserts
the row without an explicit value) — locks in the new field. Full repo
test suite re-verified green (isolated per-package + sequential `-p 1`
full run) after this addendum.

## Fase 7d — Ads — DONE (2026-07-20)

**Confirmed public, untargeted ad serving** — Ads has no per-advertiser or
per-region scoping (unlike Event/Directory, which are tied to a specific
venue/store). Every ad is shown to the entire user base identically, so
"is this campaign active today" has exactly one relevant wallclock for the
whole platform — `default_timezone` — not a per-row field. This confirms the
WALLCLOCK-PLATFORM classification from the original scoping pass.

**Bug**: `postgres_ads_query.go` answered "what is today" 3 mutually
inconsistent ways in the same file: `GetHomeAds` bound
`time.Now().Truncate(24h)` (Go-process-local, effectively UTC on this infra)
as a SQL parameter; `ListActive`'s main list used the same value but
string-interpolated via `fmt.Sprintf` into a `'YYYY-MM-DD'::date` literal;
`ListActive`'s own slider sub-query (same method, same response) used a bare
`CURRENT_DATE` (Postgres session TZ). Around UTC midnight (07:00 WIB), a
single `/mobile/ads` response could show the main list and the slider
disagreeing on which ads count as expired.

### What changed

- **`internal/app/usecase/ads/helpers.go`**: new `resolveAdsToday(ctx,
  provider port.SystemSettingsProvider) string` — reads `default_timezone`
  via `provider.GetAll()` (same map-lookup + `AsString()` shape as
  `ResolveDefaultTimezone` in the Notification dispatcher, not reused
  directly since that lives in `app/service/notification`, an
  application-layer package persistence/other-module usecases shouldn't
  import), resolves it through `sysvo.NewPlatformTimezone`, falls back to
  `sysvo.DefaultPlatformTimezone()` (`Asia/Jakarta`) on any missing/invalid
  value, and returns `.Now().Format("2006-01-02")` — a plain civil-date
  string, not a `time.Time`, so it can only ever be bound as a SQL parameter,
  never re-interpreted through some other Location by accident.
- **`internal/app/port/ads_query_model.go`**: `AdQueryReadModel.GetHomeAds`
  gained a `today string` parameter; `AdActiveListReadQuery` gained a
  `Today string` field. Both usecases (`get_home_ads.go`,
  `list_active_ads.go`) now call `resolveAdsToday` once per request and pass
  the same value through — this is what guarantees the main list and the
  slider in one `ListActive` response can no longer disagree.
- **`internal/infrastructure/persistence/postgres_ads_query.go`**: all 3
  techniques replaced with the same pattern — bind the caller-supplied
  `today` string as a parameter and cast with `$N::date` in the query text
  (a static cast on a placeholder, not a literal interpolation). `GetHomeAds`
  no longer calls `time.Now()` at all; `ListActive`'s `fmt.Sprintf` date
  literal and its slider's `CURRENT_DATE` are both gone.
- **DI**: `ads.Dependencies` gained `SystemSettings port.SystemSettingsProvider`,
  threaded into `NewGetHomeAdsUseCase`/`NewListActiveAdsUseCase`. Wired from
  the already-existing `sysSettingsProvider` in `cmd/app/main.go` and
  `sysSettingsRepo` in `internal/testhelper/testserver.go` (both already
  satisfied the interface for other modules — no new provider construction
  needed).

### Verification

- New `internal/app/usecase/ads/helpers_test.go`:
  `TestResolveAdsToday_FollowsDefaultTimezoneNotFixedZone` (Tokyo setting →
  result matches an independently-computed `PlatformTimezone("Asia/Tokyo").Now()`,
  proving the mechanism actually resolves through the configured zone) and
  `TestResolveAdsToday_FallbackChain` (nil provider, provider error, missing
  setting key, invalid timezone string — all 4 fall back to Asia/Jakarta).
  Wall-clock itself can't be faked here (no clock injection point in this
  helper), so these tests validate the resolution mechanism rather than a
  specific frozen instant — same limitation/approach as the DND fix's
  `TestNowInUserTimezone_FallbackChain`.
- `go build ./...` / `go vet ./...` clean.
- Full `go test ./internal/interfaces/http/handler/mobile/...` (62s) and
  `.../web/...` (19s) green, no regressions — existing
  `TestMobileAds_GetHomeAds_Success` / `TestMobileAds_ListActiveAds_Success`
  continue to pass against the new signature.

## Not yet scoped in detail (flagged, not designed)

- **Directory business hours** — `internal/app/usecase/directory/hours_helper.go`'s
  `isOpenNow` calls `time.Now()` with **zero** timezone conversion at all
  (not even Jakarta) — a plain bug, not merely an inconsistency. Decision:
  explicit per-merchant `timezone` field (business hours are tied to a
  physical store address, same reasoning as Event), fallback
  `default_timezone` for merchants who haven't set one.

## Out of scope (verified safe, do not touch)

- `created_at`/`updated_at` and every other `TIMESTAMPTZ` instant column —
  confirmed UTC end-to-end via live query, no container/DSN/session
  timezone override found anywhere in `k-forum-api`.
- `birth_date`, `refund_date`, `effective_date`, `transaction_date` —
  recorded facts, never compared against "now."
- `subscriptions.current_period_start/end` — enforcement happens via a
  one-off scheduled job fired at a precise instant computed at
  approval time; the `DATE` columns are display-only projections of that
  instant, not something re-parsed later.
