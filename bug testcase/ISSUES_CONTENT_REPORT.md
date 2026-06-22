# Content Report — Endpoint Verification & Issues

> **Module:** Content Reports (`/api/v1/web/reports/content`)
> **Store:** [`app/stores/contentReportStore.ts`](../../app/stores/contentReportStore.ts) · **Spec:** [`API_SPEC_CONTENT_REPORT_BACKOFFICE.md`](./API_SPEC_CONTENT_REPORT_BACKOFFICE.md)
> **Verified:** 2026-06-22 (fresh `usergod` token, live API)

## Endpoint status

| # | Endpoint | Method | Status | Note |
|---|---|---|---|---|
| B1 | `/reports/content` | GET | ✅ 200 | List/queue. Pagination in `meta` (`current_page`/`per_page`/`total`). |
| B2 | `/reports/content/{id}` | GET | ✅ 200 | Detail. |
| B3 | `/reports/content/target?reportable_type=&reportable_id=` | GET | 🔴 **500** | "Other reports on same target" — **backend bug**. |
| B4 | `/reports/content/{id}/claim` | POST | ✅ 200 | Claim / start review. *(409 `DOMAIN_CONTENT_REPORT_ALREADY_RESOLVED` if already resolved — correct.)* |
| B5 | `/reports/content/{id}/resolve` | POST | ✅ 200 | Body: `{resolution, action?, note?, cascade_target?}`. `resolution`: `action_taken`\|`dismissed`; `action`: `remove_content`\|`ban_author`\|`suspend_community`\|`none`. |
| B6 | `/reports/content/bulk-resolve` | POST | ✅ 200 | Body: `{report_ids[], resolution, note?}` → `{resolved, skipped}`. |
| B7 | `/reports/content/reporters/{user_id}/trust` | GET | ✅ 200 | Reporter trust (`total_reports`, `resolved`, `dismissed`, `dismiss_ratio`, `flag`). |
| B8 | `/reports/content/stats?period=30d&community_id=` | GET | ✅ 200 | Counts by status/reason + `avg_resolution_hours`. |

## 🔴 Backend issue
- **B3 `GET /reports/content/target` → 500** on every request. All other read fields work; only this aggregate fails. Blocks the "other reports on this target" panel.

## 🐛 Frontend bugs (FIXED 2026-06-22)
These were wrong paths in the store, not backend faults:
| Function | Was (404) | Now |
|---|---|---|
| `claimReview` | `POST …/{id}/reviews` | **`POST …/{id}/claim`** |
| `fetchReporterTrust` | `GET …/reporters/{id}` | **`GET …/reporters/{id}/trust`** |

## Notes
- Resolve works without a prior `claim` (claim is optional, not a required gate).
- `reportable_type` seen: `news_post`, `directory_listing`.
