# Filter Issues — Consolidated (all modules)

**Last verified:** 2026-07-24
**Method:** live calls to production API `https://k-forum-api.yubicom.co.id/api/v1`
with demo account `test_member`. Each list/browse param was probed three ways:
baseline (no filter) vs a **valid** value vs a **garbage** value (e.g.
`00000000-0000-0000-0000-000000000000`). A param is judged **IGNORED** when a
garbage value still returns the full baseline set, and **WORKS** when a garbage
value returns `0`/error and valid values narrow the set correctly.

> This file supersedes the filter-specific notes in
> `NEWS_SCOPE_FILTER_ISSUE.md` and the filter sections of the per-module issue
> docs. Keep all filter findings here so there is a single source of truth.

---

## TL;DR — what's broken

Backend **silently ignores** these filter params (app sends them correctly):

| Module | Endpoint | Ignored param(s) |
|---|---|---|
| Directory | `GET /mobile/directory/merchants` | `category_id`, `is_open_now` |
| Community | `GET /mobile/communities` | `category_id` |
| Events | `GET /mobile/events` | `region_id`, `city`, `is_free`, `is_featured` |

Everything else tested **works** server-side. Notably, **News `scope_id` and
`category_id` now WORK** — the old client-side `_onlyScope` workaround is
obsolete (see News section).

---

## Full matrix

Legend: ✅ works · ❌ ignored (bug) · ⚠️ inconclusive (no test data)

### News — `GET /mobile/news/articles`
| Param | Result | Evidence |
|---|---|---|
| `scope_id` | ✅ | Indonesia vs Korea return different article sets; garbage → 0 |
| `category_id` | ✅ | Latest vs Nasional differ; garbage → 0 |
| `q` | ✅ | no-match → 0 |

List payload uses **flat** fields `scope_id` / `scope_name` / `category_name`
(no nested `scope` object).

### Directory — `GET /mobile/directory/merchants` (baseline: 3 merchants, all Food & Beverage)
| Param | Result | Evidence |
|---|---|---|
| `category_id` | ❌ | Retail → 3, garbage → 3 (should be 0) |
| `is_open_now` | ❌ | true → 3, false → 3 |
| `search` | ✅ | "Noodle" → 1, no-match → 0 |
| `type` | ✅ | `food_beverage` → 3, `retail` → 0 (**note:** app's category chips send `category_id`, not `type`) |
| `city` | ✅ | garbage → 0 |
| `rating_min` | ✅ | 5 → 0, 0 → 3 |
| `has_product` | ✅ | true → 0, false → 3 |
| `has_service` | ✅ | true → 0, false → 3 |

### Community — `GET /mobile/communities` (baseline: 8 communities)
| Param | Result | Evidence |
|---|---|---|
| `category_id` | ❌ | Bisnis&UMKM → 8, garbage → 8 (should be 1) |
| `q` | ✅ | "Catur" → 1, no-match → 0 |
| `region_id` | ✅ | garbage → 0 |
| `visibility` | ✅ | public → 7, private → 1, garbage → 0 |

### Events — `GET /mobile/events` (baseline: 17 events)
| Param | Result | Evidence |
|---|---|---|
| `category_id` | ✅ | Budaya Korea → 1, garbage → 0 |
| `search` | ✅ | no-match → 0 |
| `region_id` | ❌ | garbage → 17 |
| `city` | ❌ | garbage → 17 |
| `is_free` | ❌ | true → 17, false → 17 |
| `is_featured` | ❌ | true → 17, false → 17 |

### QnA
| Endpoint | Param | Result | Evidence |
|---|---|---|---|
| `GET /mobile/qna/questions/public` | `category_id` | ⚠️ | all 6 public questions have no category assigned; garbage → 0 suggests it is parsed. Assume works (same pattern as FAQ). |
| `GET /mobile/qna/questions/public` | `filter` | ✅ | unanswered → 4, answered → 6, garbage → falls back to all (6) |
| `GET /mobile/qna/faqs` | `category_id` | ✅ | Imigrasi&Visa → 2, garbage → 0 (baseline 24) |

### Region — `GET /mobile/regions`
| Param | Result | Evidence |
|---|---|---|
| `search` | ✅ | no-match → 0 (baseline 11) |

### Announcements — `GET /mobile/announcements`
| Param | Result | Evidence |
|---|---|---|
| `type` | ✅ | garbage → HTTP 422 validation error (param is validated) |

---

## Root cause & fix (backend)

For every ❌ row the app already sends the correct value (verified: id chips
send the category/region **UUID**, booleans send `true`/`false`). The handlers
simply don't apply these params to the query. Fix is on the API:

1. **Directory** `GET /mobile/directory/merchants` — apply `category_id`
   (match against `merchant.categories[].id`) and `is_open_now`.
2. **Community** `GET /mobile/communities` — apply `category_id`
   (match against `community.category.id`).
3. **Events** `GET /mobile/events` — apply `region_id`, `city`, `is_free`,
   `is_featured`.

`category_id` / `region_id` filtering already works on other endpoints
(Events category, Community region, QnA/FAQ category), so this is
inconsistent coverage, not a missing capability.

## Client-side notes

- **News:** backend now honours `scope_id`/`category_id`. The client
  `_onlyScope` / `_applyScope` guards in `news_screen.dart` and
  `news_all_screen.dart` are now redundant (kept harmless because the model
  maps `scope_id`→`scope`). Remove once we trust the backend filter.
- **Directory category:** as a stop-gap only, results could be filtered
  client-side by `merchant.categories[].id`, but this breaks pagination
  (server returns a full unfiltered page), so prefer the backend fix.
- **Test data is thin** (3 merchants / 8 communities), so even after the
  backend fix the filters won't visibly narrow much until more data exists.

## Reproduce

```bash
BASE="https://k-forum-api.yubicom.co.id/api/v1"
TOKEN=$(curl -s -X POST -H "Content-Type: application/json" \
  --data '{"identifier":"test_member","password":"Member1234","device_id":"x"}' \
  "$BASE/mobile/auth/login" | python3 -c 'import sys,json;print(json.load(sys.stdin)["data"]["token"])')

# Directory category ignored → both return the same 3
curl -s -H "Authorization: Bearer $TOKEN" "$BASE/mobile/directory/merchants?limit=50&category_id=00000000-0000-0000-0000-000000000000"
```

> Some WAF setups reject the default `python-urllib` User-Agent with 403 —
> send a browser-like `User-Agent` header (curl's default works).
