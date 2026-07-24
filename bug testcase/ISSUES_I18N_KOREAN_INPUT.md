# Issue — Non-ASCII / Korean (Hangul) Input Blocked by Slug Handling

> **Status:** open · **Found:** 2026-07-24 · **For:** backend team
> **How found:** bulk-translating master-data category/scope names to Korean via curl (usergod token, prod `https://k-forum-api.yubicom.co.id`).
> **Summary:** several master-data endpoints **cannot accept a pure-Hangul `name`** (or any non-Latin script). The `name` column itself stores UTF-8 fine — the failure is entirely in **slug** handling. Two distinct root causes below.

Legend: `web` = `/api/v1/web/*` (backoffice).

---

## TL;DR

| Module | Endpoint | Korean name? | Root cause |
|---|---|---|---|
| Q&A categories | `PUT /web/qna/categories/:id` | ❌ | B — slug regenerated from name |
| Community categories | `PATCH /web/communities/categories/:id` | ❌ | B — slug regenerated from name |
| Regions | `POST /web/regions` | ❌ | A — slug validated ASCII-only |
| News categories | `POST /web/news/categories` | ❌ (create) | A — slug validated ASCII-only |
| News categories | `PUT /web/news/categories/:id` | ✅ (update) | keeps client slug |
| News scopes | `POST`/`PUT /web/news/scopes` | ✅ | accepts Unicode slug |
| Directory categories | `PUT /web/directory/categories/:id` | ✅ | no slug field |
| Event categories | `PUT /web/events/categories/:id` | ✅ | no slug field |

**Impact:** Q&A category, Community category, and Region names cannot be created/renamed in Korean until backend is fixed. The other four masters are already fully translated (Directory 20/20, News categories 30/30, News scopes 3/3, Event categories 11/11).

---

## Cause A — slug validated ASCII-only

The endpoint rejects a slug containing Hangul characters that the client sends.

**A1. `POST /api/v1/web/news/categories`**
```bash
curl -X POST "$BASE/api/v1/web/news/categories" -H "Authorization: Bearer $T" \
  -H "Content-Type: application/json" \
  -d '{"name":"한국어 테스트","slug":"한국어-테스트","is_active":true,"sort_order":999}'
# → 422 {"error_code":"ERR_UNPROCESSABLE_ENTITY","errors":"Invalid news category slug."}
```

**A2. `POST /api/v1/web/regions`**
```bash
curl -X POST "$BASE/api/v1/web/regions" -H "Authorization: Bearer $T" \
  -H "Content-Type: application/json" \
  -d '{"name":"한국어 지역 테스트","slug":"한국어-지역-테스트","description":"test"}'
# → 422 {"error_code":"ERR_UNPROCESSABLE_ENTITY","errors":"DOMAIN_REGION_SLUG_INVALID"}
```

---

## Cause B — slug regenerated from `name` (client slug ignored)

On update the server derives the slug from `name` with an ASCII-only slugify. A Hangul name → empty slug → "slug required". The `slug` sent by the client is discarded.

**Proof (Q&A):** send an ASCII name plus a valid slug, and the response shows the slug was re-derived from the name:
```bash
curl -X PUT "$BASE/api/v1/web/qna/categories/<id>" -H "Authorization: Bearer $T" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","slug":"imigrasi-dan-visa"}'
# → 200, but response data.slug == "test"  (derived from name, NOT the "imigrasi-dan-visa" we sent)
```

So a Hangul name fails:
```bash
curl -X PUT "$BASE/api/v1/web/qna/categories/<id>" -H "Authorization: Bearer $T" \
  -H "Content-Type: application/json" \
  -d '{"name":"이민 및 비자","slug":"imigrasi-dan-visa"}'
# → 422 {"errors":"DOMAIN_QNA_CATEGORY_SLUG_REQUIRED"}
```

**B1. `PUT /api/v1/web/qna/categories/:id`** → `422 DOMAIN_QNA_CATEGORY_SLUG_REQUIRED`
**B2. `PATCH /api/v1/web/communities/categories/:id`** → `422 DOMAIN_COMMUNITY_CATEGORY_SLUG_REQUIRED`

> Side note: `PATCH /api/v1/web/qna/categories/:id` → **404** (only `PUT` exists for category update).

---

## Why the other modules already work (keep them working)

- **`PUT /web/directory/categories/:id`** and **`PUT /web/events/categories/:id`** — no `slug` field at all → Hangul `name` accepted.
- **`PUT /web/news/categories/:id`** — **keeps** the client-sent `slug` (does NOT regenerate from name). Updating with a Hangul `name` while re-sending the existing ASCII slug succeeds. (This is why category *update* works but *create* — Cause A1 — does not.)
- **`POST`/`PUT /web/news/scopes`** — accepts a Hangul slug directly. This is the reference behaviour the others should match.

---

## Requested fix

- **Cause A** (`news/categories` POST, `regions` POST): allow Unicode letters/numbers in slug validation (e.g. regex class `\p{L}\p{N}` with the `u` flag), **or** accept the client-provided slug as-is. `news/scopes` already does this — align the others to it.
- **Cause B** (`qna/categories`, `communities/categories` update): stop regenerating the slug from `name` — respect the client-sent slug; **or** make the server slugify Unicode-aware; **or** fall back to a non-empty slug when slugify yields empty.
- **Ideal (both):** add a `translations[]` field (as `news/categories` & `news/scopes` already have: `[{language, name}]`) so the base name/slug can stay stable while the Korean name is stored per-locale — no destructive rename needed.

---

## Frontend status (already done this session)

The backoffice slug generators were made Unicode-aware so the client no longer produces an empty slug:
`app/pages/qa/index.vue`, `app/pages/news/categories.vue`, `app/pages/news/scopes.vue`, `app/pages/regions/index.vue` (`/[^\p{L}\p{N}\s-]+/gu`). This is necessary but **not sufficient** — Cause A/B are enforced server-side and still block end-to-end until the backend is fixed.
