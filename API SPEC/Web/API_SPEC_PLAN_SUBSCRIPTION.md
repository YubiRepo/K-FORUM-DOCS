# KAI Subscription API Spec — Backoffice (Admin Panel)

Dokumentasi API endpoint untuk subscription management di admin panel (Superadmin & Admin Region).

---

## Informasi Umum

- **Base URL Prefix**: `/api/v1/admin/subscription`
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Accept-Language: <lang_code>` (contoh: `ko`, `id`, `en`. Default: `ko`)
  - `Authorization: Bearer <access_token>` (untuk endpoint yang membutuhkan autentikasi)

- **Authentication**: ALL endpoint require Bearer token + Admin role verification

---

## ROLE-BASED ACCESS

```
Endpoint                                Superadmin  Admin Region
─────────────────────────────────────────────────────────────────
POST /subscription/request/:id/approve      ✓           ✗
POST /subscription/request/:id/reject       ✓           ✗
GET /subscription/requests                  ✓ (all)     ✓ (own region)
POST /subscription/downgrade                ✓           ✗
POST /subscription/refund                   ✓           ✗
GET /subscription/history                   ✓ (all)     ✓ (own region)
GET /subscription/analytics                 ✓           ✗
GET /plans                                  ✓           ✗
POST /plans                                 ✓           ✗
PUT /plans/:id                              ✓           ✗
GET /benefits/master                        ✓           ✗
POST /benefits/master                       ✗           ✗ (usergod only)
GET /benefits/plan/:plan_id                 ✓           ✗
PUT /benefits/plan/:plan_id                 ✓           ✗
```

---

## 1. GET /subscription/requests

**Description**: Ambil daftar subscription requests (pending/completed/rejected)

**Authentication**: Required (Superadmin / Admin Region)

**Method**: GET

**URL**: `/api/v1/admin/subscription/requests`

**Query Parameters**:
- `status`: "pending" | "completed" | "rejected" | "failed" (optional, default: all)
- `user_id`: uuid (optional, filter by user)
- `plan`: "standard" | "pro" (optional)
- `payment_method`: "manual" | "stripe" | "midtrans" (optional)
- `created_from`: "YYYY-MM-DD" (optional)
- `created_to`: "YYYY-MM-DD" (optional)
- `limit`: 20 (default), max 100
- `offset`: 0 (default)
- `sort`: "newest" | "oldest" (default: newest)

**Note**: Admin Region hanya bisa lihat requests dari user dalam region mereka.

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "req_001",
      "user": {
        "id": "user_123",
        "name": "John Doe",
        "email": "john@example.com",
        "region": "Jakarta"
      },
      "current_plan": "standard",
      "requested_plan": "pro",
      "amount": 80000,
      "status": "pending",
      "payment_method": "manual",
      "payment_provider_transaction_id": null,
      "manual_proof_url": "https://...",
      "manual_proof_type": "image",
      "verified_by_admin": false,
      "rejection_reason": null,
      "created_at": "2026-05-24T10:00:00.000Z",
      "approved_at": null
    },
    {
      "id": "req_002",
      "user": {
        "id": "user_456",
        "name": "Jane Smith",
        "email": "jane@example.com",
        "region": "Jakarta"
      },
      "current_plan": "standard",
      "requested_plan": "pro",
      "amount": 80000,
      "status": "pending",
      "payment_method": "stripe",
      "payment_provider_transaction_id": "txn_stripe_xxx",
      "manual_proof_url": null,
      "manual_proof_type": null,
      "verified_by_admin": false,
      "rejection_reason": null,
      "created_at": "2026-05-24T12:00:00.000Z",
      "approved_at": null
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 45
  }
}
```

---

## 2. GET /subscription/requests/:request_id

**Description**: Ambil detail single subscription request

**Authentication**: Required (Superadmin / Admin Region)

**Method**: GET

**URL**: `/api/v1/admin/subscription/requests/:request_id`

**Response (200 OK)**:
```json
{
  "data": {
    "id": "req_001",
    "user": {
      "id": "user_123",
      "name": "John Doe",
      "email": "john@example.com",
      "phone": "08123456789",
      "region": "Jakarta",
      "created_at": "2026-01-15T00:00:00.000Z"
    },
    "current_plan": "standard",
    "requested_plan": "pro",
    "amount": 80000,
    "status": "pending",
    "payment_method": "manual",
    "payment_provider_transaction_id": null,
    "manual_proof_url": "https://...",
    "manual_proof_type": "image",
    "verified_by_admin": false,
    "admin_notes": null,
    "rejection_reason": null,
    "created_at": "2026-05-24T10:00:00.000Z",
    "approved_at": null,
    "approved_by": null
  }
}
```

---

## 3. POST /subscription/requests/:request_id/approve

**Description**: Approve subscription request (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: POST

**URL**: `/api/v1/admin/subscription/requests/:request_id/approve`

**Request Body**:
```json
{
  "admin_notes": "Verified via bank statement" | null
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "request_id": "req_001",
    "status": "completed",
    "user_id": "user_123",
    "old_plan": "standard",
    "new_plan": "pro",
    "subscription_id": "sub_pro_xxx",
    "approved_at": "2026-05-24T10:05:00.000Z",
    "approved_by": "superadmin_user_id"
  },
  "message": "Request approved. User subscription activated.",
  "events": [
    {
      "type": "email_sent",
      "recipient": "john@example.com",
      "template": "subscription_activated"
    },
    {
      "type": "notification_sent",
      "user_id": "user_123",
      "message": "Your Pro plan is now active"
    }
  ]
}
```

**Response (409 Conflict - Already completed/rejected)**:
```json
{
  "message": "Cannot approve non-pending request",
  "data": {
    "current_status": "completed"
  }
}
```

---

## 4. POST /subscription/requests/:request_id/reject

**Description**: Reject subscription request (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: POST

**URL**: `/api/v1/admin/subscription/requests/:request_id/reject`

**Request Body**:
```json
{
  "reason": "Amount doesn't match" | "Invalid proof" | "Duplicate request" | "Other",
  "rejection_details": "Bank transfer amount is Rp 70,000 but should be Rp 80,000" | null
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "request_id": "req_001",
    "status": "rejected",
    "reason": "Amount doesn't match",
    "rejection_details": "Bank transfer amount is Rp 70,000 but should be Rp 80,000",
    "rejected_at": "2026-05-24T10:05:00.000Z",
    "rejected_by": "superadmin_user_id"
  },
  "message": "Request rejected.",
  "events": [
    {
      "type": "email_sent",
      "recipient": "john@example.com",
      "template": "subscription_rejected",
      "variables": {
        "reason": "Amount doesn't match"
      }
    }
  ]
}
```

---

## 5. POST /subscription/downgrade

**Description**: Downgrade user subscription manually (admin-initiated refund, etc)

**Authentication**: Required (Superadmin only)

**Method**: POST

**URL**: `/api/v1/admin/subscription/downgrade`

**Request Body**:
```json
{
  "user_id": "user_123",
  "new_plan": "standard",
  "reason": "User refund request" | "Admin action" | "Promo adjustment" | "Other",
  "effective_date": "2026-05-24" | null
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "user_id": "user_123",
    "old_plan": "pro",
    "new_plan": "standard",
    "action": "downgrade",
    "initiated_by": "superadmin",
    "reason": "User refund request",
    "effective_date": "2026-05-24",
    "downgraded_at": "2026-05-24T10:05:00.000Z"
  },
  "message": "User downgraded from Pro to Standard",
  "events": [
    {
      "type": "email_sent",
      "recipient": "john@example.com",
      "template": "subscription_downgraded"
    },
    {
      "type": "subscription_history_created",
      "action": "downgrade",
      "initiated_by": "admin"
    }
  ]
}
```

---

## 6. POST /subscription/refund

**Description**: Process manual refund (cancel subscription, return to previous plan)

**Authentication**: Required (Superadmin only)

**Method**: POST

**URL**: `/api/v1/admin/subscription/refund`

**Request Body**:
```json
{
  "user_id": "user_123",
  "refund_amount": 129000,
  "reason": "User request" | "Payment error" | "Service issue" | "Other",
  "notes": "Refund processed to bank account BCA xxxxxx" | null
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "refund_id": "refund_xxx",
    "user_id": "user_123",
    "subscription_id_cancelled": "sub_pro_xxx",
    "refund_amount": 129000,
    "refund_date": "2026-05-24",
    "reason": "User request",
    "notes": "Refund processed to bank account BCA xxxxxx",
    "old_plan": "pro",
    "new_plan": "standard",
    "refund_status": "processed"
  },
  "message": "Refund processed. User downgraded to Standard.",
  "events": [
    {
      "type": "email_sent",
      "recipient": "john@example.com",
      "template": "refund_processed",
      "variables": {
        "amount": "Rp 129,000",
        "account": "BCA xxxxxx"
      }
    }
  ]
}
```

---

## 7. GET /subscription/history

**Description**: Ambil subscription history (all users untuk superadmin, own region untuk admin)

**Authentication**: Required (Superadmin / Admin Region)

**Method**: GET

**URL**: `/api/v1/admin/subscription/history`

**Query Parameters**:
- `user_id`: uuid (optional)
- `action`: "signup" | "upgrade" | "downgrade" | "renewal" | "expired" | "cancelled" (optional)
- `created_from`: "YYYY-MM-DD" (optional)
- `created_to`: "YYYY-MM-DD" (optional)
- `limit`: 20 (default), max 100
- `offset`: 0 (default)

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "hist_001",
      "user": {
        "id": "user_123",
        "name": "John Doe",
        "email": "john@example.com"
      },
      "old_plan": null,
      "new_plan": "standard",
      "action": "signup",
      "initiated_by": "system",
      "created_at": "2026-01-15T00:00:00.000Z"
    },
    {
      "id": "hist_002",
      "user": {
        "id": "user_123",
        "name": "John Doe",
        "email": "john@example.com"
      },
      "old_plan": "standard",
      "new_plan": "pro",
      "action": "upgrade",
      "initiated_by": "user",
      "created_at": "2026-05-10T10:00:00.000Z"
    }
  ],
  "pagination": {
    "limit": 20,
    "offset": 0,
    "total": 256
  }
}
```

---

## 8. GET /subscription/analytics

**Description**: Ambil subscription analytics dashboard (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: GET

**URL**: `/api/v1/admin/subscription/analytics`

**Query Parameters**:
- `date_from`: "YYYY-MM-DD" (optional, default: 30 days ago)
- `date_to`: "YYYY-MM-DD" (optional, default: today)

**Response (200 OK)**:
```json
{
  "data": {
    "kpis": {
      "total_users": 5234,
      "pro_users": 380,
      "standard_users": 4854,
      "mrr": 49020000,
      "conversion_rate": 7.3,
      "churn_rate": 8.2,
      "avg_revenue_per_user": 9358
    },
    "breakdown_by_plan": {
      "standard": {
        "users": 4854,
        "revenue": 0,
        "growth_percentage": null,
        "note": "Default plan, no revenue"
      },
      "pro": {
        "users": 380,
        "revenue": 49020000,
        "growth_percentage": 5.2,
        "new_users_this_month": 20,
        "cancelled_this_month": 5
      }
    },
    "churn_analysis": {
      "total_churned": 5,
      "top_reasons": [
        {
          "reason": "Features not needed",
          "count": 3,
          "percentage": 60
        },
        {
          "reason": "Too expensive",
          "count": 2,
          "percentage": 40
        }
      ]
    },
    "payment_analysis": {
      "total_requests": 25,
      "approval_rate": 84,
      "pending": 3,
      "failed_or_rejected": 4
    },
    "date_from": "2026-04-24",
    "date_to": "2026-05-24"
  }
}
```

---

## 9. GET /plans

**Description**: Ambil daftar semua plans (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: GET

**URL**: `/api/v1/admin/subscription/plans`

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "plan_standard",
      "name": "Standard",
      "price": 49000,
      "currency": "IDR",
      "duration_days": 30,
      "description": "Akses fitur utama: baca, join komunitas, tanya QnA",
      "is_default": true,
      "status": "active",
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-05-24T00:00:00.000Z"
    },
    {
      "id": "plan_pro",
      "name": "Pro",
      "price": 129000,
      "currency": "IDR",
      "duration_days": 30,
      "description": "Standard + posting news (approval), buat community, buat company/merchants, buat event",
      "is_default": false,
      "status": "active",
      "created_at": "2026-01-01T00:00:00.000Z",
      "updated_at": "2026-05-24T00:00:00.000Z"
    }
  ]
}
```

---

## 10. POST /plans

**Description**: Create new plan (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: POST

**URL**: `/api/v1/admin/subscription/plans`

**Request Body**:
```json
{
  "name": "Premium",
  "price": 299000,
  "currency": "IDR",
  "duration_days": 30,
  "description": "Advanced features with priority support",
  "is_default": false,
  "status": "active"
}
```

**Response (201 Created)**:
```json
{
  "data": {
    "id": "plan_premium",
    "name": "Premium",
    "price": 299000,
    "currency": "IDR",
    "duration_days": 30,
    "description": "Advanced features with priority support",
    "is_default": false,
    "status": "active",
    "created_at": "2026-05-24T10:00:00.000Z"
  }
}
```

---

## 11. PUT /plans/:plan_id

**Description**: Update plan details (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: PUT

**URL**: `/api/v1/admin/subscription/plans/:plan_id`

**Request Body**:
```json
{
  "price": 59000 | null,
  "description": "Updated description" | null,
  "status": "active" | "disabled" | null
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "id": "plan_standard",
    "name": "Standard",
    "price": 59000,
    "currency": "IDR",
    "duration_days": 30,
    "description": "Updated description",
    "is_default": true,
    "status": "active",
    "updated_at": "2026-05-24T10:00:00.000Z"
  }
}
```

---

## 12. GET /benefits/master

**Description**: Ambil benefits master (superadmin only, informational)

**Authentication**: Required (Superadmin only)

**Method**: GET

**URL**: `/api/v1/admin/subscription/benefits/master`

**Response (200 OK)**:
```json
{
  "data": [
    {
      "id": "benefit_post_news",
      "key": "post_news",
      "description": "Member can post news",
      "approval_required": true,
      "approval_role": "superadmin",
      "created_at": "2026-01-01T00:00:00.000Z"
    },
    {
      "id": "benefit_create_community",
      "key": "create_community",
      "description": "Member can create and manage community",
      "approval_required": false,
      "approval_role": null,
      "created_at": "2026-01-01T00:00:00.000Z"
    },
    {
      "id": "benefit_create_store",
      "key": "create_store",
      "description": "Member can create store/merchant listing",
      "approval_required": false,
      "approval_role": null,
      "created_at": "2026-01-01T00:00:00.000Z"
    },
    {
      "id": "benefit_create_event",
      "key": "create_event",
      "description": "Member can host event",
      "approval_required": false,
      "approval_role": null,
      "created_at": "2026-01-01T00:00:00.000Z"
    },
    {
      "id": "benefit_view_analytics",
      "key": "view_analytics",
      "description": "Member can view analytics dashboard",
      "approval_required": false,
      "approval_role": null,
      "created_at": "2026-01-01T00:00:00.000Z"
    }
  ]
}
```

---

## 13. GET /benefits/plan/:plan_id

**Description**: Ambil benefits assignment untuk specific plan (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: GET

**URL**: `/api/v1/admin/subscription/benefits/plan/:plan_id`

**Response (200 OK)**:
```json
{
  "data": {
    "plan_id": "plan_pro",
    "plan_name": "Pro",
    "benefits": [
      {
        "id": "pb_001",
        "benefit_key": "post_news",
        "benefit_description": "Member can post news",
        "enabled": true,
        "metadata": {
          "approval_required": true
        },
        "created_at": "2026-01-01T00:00:00.000Z",
        "updated_at": "2026-05-24T00:00:00.000Z"
      },
      {
        "id": "pb_002",
        "benefit_key": "create_community",
        "benefit_description": "Member can create and manage community",
        "enabled": true,
        "metadata": null,
        "created_at": "2026-01-01T00:00:00.000Z",
        "updated_at": "2026-05-24T00:00:00.000Z"
      },
      {
        "id": "pb_003",
        "benefit_key": "create_store",
        "benefit_description": "Member can create store/merchant listing",
        "enabled": true,
        "metadata": null,
        "created_at": "2026-01-01T00:00:00.000Z",
        "updated_at": "2026-05-24T00:00:00.000Z"
      }
    ]
  }
}
```

---

## 14. PUT /benefits/plan/:plan_id

**Description**: Update benefits assignment untuk specific plan (superadmin only)

**Authentication**: Required (Superadmin only)

**Method**: PUT

**URL**: `/api/v1/admin/subscription/benefits/plan/:plan_id`

**Request Body**:
```json
{
  "benefits": [
    {
      "benefit_key": "post_news",
      "enabled": true
    },
    {
      "benefit_key": "create_community",
      "enabled": true
    },
    {
      "benefit_key": "create_event",
      "enabled": false
    }
  ]
}
```

**Response (200 OK)**:
```json
{
  "data": {
    "plan_id": "plan_pro",
    "plan_name": "Pro",
    "benefits_updated": [
      {
        "benefit_key": "post_news",
        "enabled": true
      },
      {
        "benefit_key": "create_community",
        "enabled": true
      },
      {
        "benefit_key": "create_event",
        "enabled": false
      }
    ],
    "updated_at": "2026-05-24T10:00:00.000Z"
  },
  "message": "Benefits updated. Changes apply immediately."
}
```

---

## Error Response Standard

### Format A: Standard Message Error
```json
{
  "message": "Pesan error deskriptif"
}
```

### Format B: Validation Error (422)
```json
{
  "message": "Validation failed",
  "errors": {
    "field_name": ["Error message 1"],
    "another_field": ["Error message 2"]
  }
}
```

### Format C: Permission Error (403)
```json
{
  "message": "Insufficient permissions for this action",
  "required_role": "superadmin"
}
```

---

## Status Codes

- `200 OK` - Sukses
- `201 Created` - Resource created
- `400 Bad Request` - Bad request
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Permission denied
- `404 Not Found` - Resource not found
- `409 Conflict` - Conflict
- `422 Unprocessable Entity` - Validation error
- `500 Internal Server Error` - Server error

---

## Rate Limiting

- Limit: 200 requests per minute per admin
- Headers:
  - `X-RateLimit-Limit: 200`
  - `X-RateLimit-Remaining: 195`
  - `X-RateLimit-Reset: 1629856800`

---

END OF DOCUMENT
