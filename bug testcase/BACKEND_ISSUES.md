# Backend Issues & API Anomalies

> **Status:** v1 · **Last updated:** 2026-06-20
> **Source:** discovered during backoffice integration + seeding (curl testing) this session.
> **For:** backend team. Prioritise the 🔴 Blocker items.

Legend: `web` = `/api/v1/web/*` (backoffice) · `mobile` = `/api/v1/mobile/*` (member app) · `auth` = `/api/v1/auth/*`.

---

## 🔴 Blocker — server error / broken endpoint

| Endpoint | Method | Symptom | Impact |
|---|---|---|---|
| `/api/v1/web/ads` | GET | **500** on every query variant | Ads list empty in UI even though `POST /web/ads` → `201`; analytics `total_ads_active: 0`. Likely list/aggregate query bug. |
| `/api/v1/web/regions/me` | GET | **500** | "My Managed Regions" (REGION_API_SPEC B13) unusable; frontend falls back to fetch-by-ID. |
| `/api/v1/web/auth/login` | POST | **502** (once, transient) | Empty token momentarily; retry succeeded. Worth checking for intermittent upstream timeout. |

---

## 🟠 Wrong behaviour — silently ignores input / wrong verb

| Endpoint | Method | Symptom | Correct usage |
|---|---|---|---|
| `/api/v1/web/qna/faqs` | POST | `status` field **ignored** → always created as `draft` | publish/archive only via `PATCH /qna/faqs/:id/status` |
| `/api/v1/web/qna/faqs/:id` | PATCH | returns `200` but `status` **ignored** | use `/status` sub-route |
| `/api/v1/web/qna/faqs/:id` | PUT | **404** (route doesn't exist, though spec implies it) | route should exist or spec corrected |
| `/api/v1/web/directory/categories/:id` | PUT | partial body → **422** (`name` required) | requires full object; partial update unsupported |
| `/api/v1/web/role-permission/permissions` | POST | duplicate key **not rejected** (`201` instead of `409`); some creates `201` but **don't persist** (flaky); list count inconsistent (22/23/0) | needs unique constraint + persistence fix |

---

## 🟡 Shape mismatch vs spec (frontend already adapted)

| Endpoint | Anomaly |
|---|---|
| `/api/v1/web/directory/merchants` | data in `data.items` + `data.summary` (`total_*` keys); pagination in `meta` (not `pagination`) |
| `/api/v1/web/directory/categories` | uses `order` (not `sort_order`); no `description` / `merchant_count` |
| `/api/v1/web/directory/analytics/overview` | **flat** shape, totals differ from spec (nested) |
| `/api/v1/web/directory/merchants/:id` | `stats.total_views` (not `view_count`); **no** `owner` object |
| `/api/v1/web/directory/companies/:id` | flat — **no** `owner` object, no `merchants[]` |
| all list endpoints (users, regions, …) | **`offset` ignored** — must use `page` + `limit`; total in `meta` |
| `/api/v1/web/auth/me` | **no `avatar` field** (only `/profile/me` & `/auth/session` carry it) |

---

## 🟣 Role & Permission scope

| Endpoint | Method | Symptom |
|---|---|---|
| `/api/v1/web/role-permission/role-permissions/bulk-assign` | POST | `422 DOMAIN_ROLE_ASSIGNMENT_SCOPE_MISMATCH` when assigning a **global** permission to a **region-scoped** role (`admin`). Scope of role and permission must match. |
| `/api/v1/web/role-permission/permissions` | POST | `category: "finance"` → **422** (valid categories: `content`, `moderation`, `member`, `admin`, `system`) |

---

## ⚪ Not bugs — business rules / context (documented to avoid false reports)

| Endpoint | Status | Reason |
|---|---|---|
| `/api/v1/web/directory/companies`·`/merchants` | POST → **404** | No backoffice create; companies/merchants are created owner-side via `/mobile/*` |
| `/api/v1/mobile/directory/companies`, `/mobile/events`, `/mobile/communities` | POST → **403** | `create_store` / `create_event` / `create_community` are **PRO-plan** benefits; Standard plan correctly denied |
| `/api/v1/web/subscription/requests/:id/approve` | POST → **422** | `pending` request needs payment-verification step first (state-machine transition). Worked around with `PATCH /web/users/:id/subscription`. |
| `/api/v1/web/profile/avatar` | POST → **404** | Seed account had no profile record yet (confirmed) — not an endpoint bug |
| `/api/v1/mobile/directory/merchants` | POST → **422** | Requires **both** flat `address` AND nested `location{}` |
| `/api/v1/mobile/events` | POST → **422** | `venue_name` (offline), `online_platform` (online) required; `event_type` enum must be `offline`\|`online`\|`hybrid` |

---

## ✅ Now live (previously 404 during the session)

- `GET /api/v1/web/regions/{id}/invitations` (B13) → `200`
- `GET /api/v1/web/regions/invitations/pending` (B15) → `200`

---

### Top priorities for backend
1. `GET /web/ads` → **500** (blocks the whole Ads/Promotions module).
2. `GET /web/regions/me` → **500** (blocks admin-region scoping).
3. `POST /role-permission/permissions` persistence + duplicate handling.
