# System Settings — DB Schema

**Status:** Draft v1
**Last Updated:** 2026-06-12
**Module:** System Settings
**Database:** PostgreSQL

Aturan bisnis lengkap lihat `SYSTEM_SETTINGS_RULES.md`.

---

## Daftar Isi

1. [Diagram Relasi](#diagram-relasi)
2. [system_settings](#1-system_settings)
3. [legal_documents](#2-legal_documents)
4. [legal_document_versions](#3-legal_document_versions)
5. [user_legal_acceptances](#4-user_legal_acceptances)
6. [Seed Data](#seed-data)
7. [Query Patterns](#query-patterns)
8. [Catatan Implementasi](#catatan-implementasi)

---

## Diagram Relasi

```
system_settings (key-value, standalone)

legal_documents (1) ──< legal_document_versions (N)
                                   │
                                   └──< user_legal_acceptances (N) >── users
```

---

## 1. SYSTEM_SETTINGS

Key-value store untuk semua setting scalar. Satu baris = satu setting key.
Di-seed lengkap via migration — aplikasi tidak pernah INSERT key baru saat runtime, hanya UPDATE value.

```sql
CREATE TABLE system_settings (
  id            UUID         PRIMARY KEY DEFAULT gen_random_uuid(),

  group_key     VARCHAR(50)  NOT NULL,
  -- 'general' | 'mobile_app' | 'security' | 'email' | 'storage'
  -- | 'payment' | 'moderation' | 'maintenance' | 'contact'

  setting_key   VARCHAR(100) NOT NULL,
  -- Unique key, e.g. 'app_name', 'min_version_android'

  value         JSONB        NOT NULL,
  -- Nilai disimpan sebagai JSONB agar satu kolom bisa menampung semua tipe.
  -- string  → "KAI App"
  -- number  → 5
  -- boolean → true
  -- array   → ["image/jpeg", "image/png"]

  value_type    VARCHAR(20)  NOT NULL,
  -- 'string' | 'number' | 'boolean' | 'array' | 'json'
  -- Dipakai untuk validasi server-side & rendering form di backoffice.

  is_public     BOOLEAN      NOT NULL DEFAULT false,
  -- true = diekspos via GET /api/v1/mobile/config (tanpa auth).
  -- Setting sensitif WAJIB false.

  is_sensitive  BOOLEAN      NOT NULL DEFAULT false,
  -- true = nilai dimask di response list backoffice & audit log (e.g. smtp_username).

  editable_by   VARCHAR(20)  NOT NULL DEFAULT 'superadmin',
  -- 'usergod'    = hanya usergod (setting teknis/infrastruktur)
  -- 'superadmin' = usergod + superadmin (setting operasional)
  -- Dicek di middleware selain permission 'manage_system_settings'.

  description   VARCHAR(500) NULL,
  -- Penjelasan singkat — tampil sebagai helper text di backoffice.

  updated_by    UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  updated_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  created_at    TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT uq_system_settings_key UNIQUE (setting_key),
  CONSTRAINT ck_system_settings_value_type
    CHECK (value_type IN ('string', 'number', 'boolean', 'array', 'json')),
  CONSTRAINT ck_system_settings_editable_by
    CHECK (editable_by IN ('usergod', 'superadmin'))
);

CREATE INDEX idx_system_settings_group
  ON system_settings (group_key);

CREATE INDEX idx_system_settings_public
  ON system_settings (is_public)
  WHERE is_public = true;
```

**Kenapa key-value, bukan tabel singleton berkolom banyak?**
- Tambah setting baru = INSERT seed, tanpa ALTER TABLE / migrasi schema.
- Metadata per setting (`is_public`, `is_sensitive`, `editable_by`) menempel di baris — permission & exposure granular per key.
- Trade-off: tidak ada type safety di level kolom — ditutup oleh `value_type` + validasi service layer.

---

## 2. LEGAL_DOCUMENTS

Master dokumen legal. Satu baris per jenis dokumen — dibuat via seed, tidak bertambah saat runtime.

```sql
CREATE TABLE legal_documents (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),

  doc_type     VARCHAR(50) NOT NULL UNIQUE,
  -- 'terms' | 'privacy' | 'community_guidelines'

  title        VARCHAR(200) NOT NULL,
  -- 'Syarat & Ketentuan', 'Kebijakan Privasi', 'Pedoman Komunitas'

  created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
  updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  CONSTRAINT ck_legal_documents_type
    CHECK (doc_type IN ('terms', 'privacy', 'community_guidelines'))
);
```

---

## 3. LEGAL_DOCUMENT_VERSIONS

Setiap versi konten dokumen. Versi `published` immutable — perubahan = versi baru.

```sql
CREATE TABLE legal_document_versions (
  id                    UUID         PRIMARY KEY DEFAULT gen_random_uuid(),
  document_id           UUID         NOT NULL REFERENCES legal_documents(id) ON DELETE CASCADE,

  version               VARCHAR(20)  NOT NULL,
  -- Format bebas semver-ish: '1.0.0', '2.1.0'

  content               TEXT         NOT NULL,
  -- Konten lengkap dalam Markdown.

  status                VARCHAR(20)  NOT NULL DEFAULT 'draft',
  -- 'draft'     = sedang disusun, bisa diedit bebas
  -- 'published' = aktif — HANYA SATU per document (enforced via partial unique index)
  -- 'archived'  = versi lama, read-only

  effective_date        DATE         NULL,
  -- Tanggal mulai berlaku. Wajib diisi saat publish.

  require_reacceptance  BOOLEAN      NOT NULL DEFAULT false,
  -- true = semua user diminta menyetujui ulang setelah versi ini published.

  published_at          TIMESTAMPTZ  NULL,
  published_by          UUID         NULL REFERENCES users(id) ON DELETE SET NULL,

  created_by            UUID         NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
  updated_by            UUID         NULL REFERENCES users(id) ON DELETE SET NULL,
  created_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
  updated_at            TIMESTAMPTZ  NOT NULL DEFAULT NOW(),

  CONSTRAINT uq_legal_versions_doc_version UNIQUE (document_id, version),
  CONSTRAINT ck_legal_versions_status
    CHECK (status IN ('draft', 'published', 'archived'))
);

-- Hanya satu versi 'published' per dokumen pada satu waktu.
CREATE UNIQUE INDEX idx_legal_versions_one_published
  ON legal_document_versions (document_id)
  WHERE status = 'published';

CREATE INDEX idx_legal_versions_document
  ON legal_document_versions (document_id, status, created_at DESC);
```

**Transisi status (di service layer, dalam satu transaksi):**

```
draft ──publish──> published ──(versi baru di-publish)──> archived
```

```sql
-- Pseudo-flow publish versi baru (1 transaksi):
BEGIN;
  UPDATE legal_document_versions
     SET status = 'archived', updated_at = NOW()
   WHERE document_id = :doc_id AND status = 'published';

  UPDATE legal_document_versions
     SET status = 'published', published_at = NOW(),
         published_by = :actor_id, effective_date = :effective_date
   WHERE id = :new_version_id AND status = 'draft';
COMMIT;
```

---

## 4. USER_LEGAL_ACCEPTANCES

Jejak persetujuan user terhadap versi dokumen. Append-only — tidak pernah di-update/delete.

```sql
CREATE TABLE user_legal_acceptances (
  id           UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id      UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  version_id   UUID        NOT NULL REFERENCES legal_document_versions(id) ON DELETE CASCADE,

  accepted_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

  source       VARCHAR(20) NOT NULL DEFAULT 'app',
  -- 'registration' = otomatis saat daftar (menyetujui versi published saat itu)
  -- 'app'          = dialog persetujuan ulang di mobile
  -- 'web'          = dari backoffice/web

  ip_address   INET        NULL,
  -- Opsional, untuk kebutuhan legal/audit.

  CONSTRAINT uq_user_legal_acceptance UNIQUE (user_id, version_id),
  CONSTRAINT ck_legal_acceptance_source
    CHECK (source IN ('registration', 'app', 'web'))
);

CREATE INDEX idx_legal_acceptances_user
  ON user_legal_acceptances (user_id);

CREATE INDEX idx_legal_acceptances_version
  ON user_legal_acceptances (version_id);
```

---

## Seed Data

### Legal Documents

```sql
INSERT INTO legal_documents (doc_type, title) VALUES
  ('terms',                'Syarat & Ketentuan'),
  ('privacy',              'Kebijakan Privasi'),
  ('community_guidelines', 'Pedoman Komunitas')
ON CONFLICT (doc_type) DO NOTHING;
```

### System Settings

```sql
INSERT INTO system_settings
  (group_key, setting_key, value, value_type, is_public, is_sensitive, editable_by, description)
VALUES
  -- ── general ──────────────────────────────────────────────────────────────
  ('general', 'app_name',                    '"KAI App"',          'string',  true,  false, 'superadmin', 'Nama platform'),
  ('general', 'tagline',                     '"Korea Asosiasi Indonesia"', 'string', true, false, 'superadmin', 'Tagline platform'),
  ('general', 'support_email',               '"support@kai.or.id"','string',  true,  false, 'superadmin', 'Email support'),
  ('general', 'platform_url',                '"https://kai.or.id"','string',  true,  false, 'superadmin', 'URL utama platform'),
  ('general', 'default_timezone',            '"Asia/Jakarta"',     'string',  false, false, 'superadmin', 'Timezone default'),
  ('general', 'default_language',            '"id"',               'string',  true,  false, 'superadmin', 'Bahasa default UI — harus ada di system_languages'),
  ('general', 'public_registration_enabled', 'true',               'boolean', false, false, 'superadmin', 'Izinkan registrasi publik'),
  ('general', 'email_verification_required', 'true',               'boolean', false, false, 'superadmin', 'Wajib verifikasi OTP email setelah daftar'),

  -- ── mobile_app ───────────────────────────────────────────────────────────
  ('mobile_app', 'min_version_android',    '"1.0.0"', 'string',  true,  false, 'superadmin', 'Versi minimum Android (semver)'),
  ('mobile_app', 'min_version_ios',        '"1.0.0"', 'string',  true,  false, 'superadmin', 'Versi minimum iOS (semver)'),
  ('mobile_app', 'latest_version_android', '"1.0.0"', 'string',  true,  false, 'superadmin', 'Versi terbaru di Play Store'),
  ('mobile_app', 'latest_version_ios',     '"1.0.0"', 'string',  true,  false, 'superadmin', 'Versi terbaru di App Store'),
  ('mobile_app', 'force_update_enabled',   'true',    'boolean', true,  false, 'superadmin', 'Blokir app di bawah versi minimum'),
  ('mobile_app', 'playstore_url',          '""',      'string',  true,  false, 'superadmin', 'Link Play Store'),
  ('mobile_app', 'appstore_url',           '""',      'string',  true,  false, 'superadmin', 'Link App Store'),
  ('mobile_app', 'update_message',         '"Versi baru tersedia. Silakan perbarui aplikasi Anda."', 'string', true, false, 'superadmin', 'Pesan dialog update'),

  -- ── security (usergod only) ──────────────────────────────────────────────
  ('security', 'max_login_attempts',           '5',     'number',  false, false, 'usergod', 'Gagal login sebelum lockout'),
  ('security', 'lockout_duration_minutes',     '30',    'number',  false, false, 'usergod', 'Durasi lockout'),
  ('security', 'otp_expiry_minutes',           '5',     'number',  false, false, 'usergod', 'Masa berlaku OTP'),
  ('security', 'otp_max_attempts',             '5',     'number',  false, false, 'usergod', 'Maks salah input OTP'),
  ('security', 'otp_resend_cooldown_seconds',  '60',    'number',  false, false, 'usergod', 'Cooldown kirim ulang OTP'),
  ('security', 'access_token_expiry_minutes',  '60',    'number',  false, false, 'usergod', 'Expiry JWT access token'),
  ('security', 'refresh_token_expiry_days',    '30',    'number',  false, false, 'usergod', 'Expiry refresh token'),
  ('security', 'seamless_token_expiry_seconds','60',    'number',  false, false, 'usergod', 'TTL one-time token seamless login'),
  ('security', 'password_min_length',          '8',     'number',  false, false, 'usergod', 'Panjang minimum password'),
  ('security', 'password_require_number',      'true',  'boolean', false, false, 'usergod', 'Password wajib angka'),
  ('security', 'password_require_uppercase',   'false', 'boolean', false, false, 'usergod', 'Password wajib huruf kapital'),
  ('security', 'password_require_symbol',      'false', 'boolean', false, false, 'usergod', 'Password wajib simbol'),

  -- ── email (usergod only; password SMTP di env) ───────────────────────────
  ('email', 'smtp_host',             '""',                 'string',  false, true,  'usergod', 'Host SMTP'),
  ('email', 'smtp_port',             '587',                'number',  false, false, 'usergod', 'Port SMTP'),
  ('email', 'smtp_username',         '""',                 'string',  false, true,  'usergod', 'Username SMTP — password di env SMTP_PASSWORD'),
  ('email', 'smtp_encryption',       '"tls"',              'string',  false, false, 'usergod', 'tls | ssl | none'),
  ('email', 'smtp_from_name',        '"KAI App"',          'string',  false, false, 'usergod', 'Nama pengirim email'),
  ('email', 'smtp_from_email',       '"noreply@kai.or.id"','string',  false, false, 'usergod', 'Alamat pengirim'),
  ('email', 'welcome_email_enabled', 'true',               'boolean', false, false, 'usergod', 'Kirim welcome email'),
  ('email', 'email_footer_text',     '"© 2026 Korea Asosiasi Indonesia. All rights reserved."', 'string', false, false, 'usergod', 'Footer email'),

  -- ── storage (usergod only; kredensial provider di env) ───────────────────
  ('storage', 'storage_provider',         '"local"', 'string', false, false, 'usergod', 'local | s3 | r2'),
  ('storage', 'max_upload_size_mb',       '10',      'number', false, false, 'usergod', 'Batas umum ukuran upload'),
  ('storage', 'max_avatar_size_mb',       '2',       'number', false, false, 'usergod', 'Batas ukuran avatar'),
  ('storage', 'max_payment_proof_size_mb','5',       'number', false, false, 'usergod', 'Batas bukti transfer'),
  ('storage', 'allowed_mime_types',
     '["image/jpeg","image/png","image/webp","application/pdf"]',
                                          'array',  false, false, 'usergod', 'Whitelist MIME upload'),
  ('storage', 'cdn_base_url',             '""',      'string', false, false, 'usergod', 'Prefix CDN (opsional)'),

  -- ── payment ──────────────────────────────────────────────────────────────
  ('payment', 'bank_name',                            '""', 'string', true,  false, 'superadmin', 'Bank tujuan transfer'),
  ('payment', 'bank_account_number',                  '""', 'string', true,  false, 'superadmin', 'Nomor rekening'),
  ('payment', 'bank_account_holder',                  '""', 'string', true,  false, 'superadmin', 'Nama pemilik rekening'),
  ('payment', 'payment_instructions',                 '""', 'string', true,  false, 'superadmin', 'Instruksi pembayaran (markdown)'),
  ('payment', 'payment_confirmation_deadline_hours',  '24', 'number', true,  false, 'superadmin', 'Batas upload bukti transfer'),
  ('payment', 'payment_provider',          '"manual"',  'string', true,  false, 'superadmin', 'manual | midtrans | both — flow pembayaran aktif'),
  ('payment', 'midtrans_environment',       '"sandbox"', 'string', false, false, 'usergod',    'sandbox | production — credentials di env MIDTRANS_SERVER_KEY/CLIENT_KEY'),
  ('payment', 'midtrans_integration_mode',  '"snap"',    'string', false, false, 'superadmin', 'snap | core — mode integrasi Midtrans aktif (snap=token+redirect, core=instruksi per-channel)'),
  ('payment', 'midtrans_enabled_channels',  '[]',        'array',  false, false, 'superadmin', 'Channel Midtrans aktif, mis. ["gopay","qris","bank_transfer"]'),

  -- ── moderation (nilai dibaca modul Reporting/Community) ──────────────────
  ('moderation', 'banned_keywords',            '[]',  'array',  false, false, 'superadmin', 'Kata terlarang — konten ditolak saat submit'),
  ('moderation', 'report_auto_flag_threshold', '5',   'number', false, false, 'superadmin', 'Report sebelum auto-flag (ref REPORTING_RULES)'),
  ('moderation', 'report_rate_limit_per_day',  '20',  'number', false, false, 'superadmin', 'Maks report per user per hari'),

  -- ── maintenance ──────────────────────────────────────────────────────────
  ('maintenance', 'maintenance_mode_enabled', 'false', 'boolean', true, false, 'superadmin', 'Aktifkan maintenance mode'),
  ('maintenance', 'maintenance_message',
     '"Kami sedang melakukan pemeliharaan sistem. Silakan coba beberapa saat lagi."',
                                              'string', true, false, 'superadmin', 'Pesan maintenance'),

  -- ── contact ──────────────────────────────────────────────────────────────
  ('contact', 'whatsapp_number', '""', 'string', true, false, 'superadmin', 'WhatsApp admin/CS'),
  ('contact', 'instagram_url',   '""', 'string', true, false, 'superadmin', 'Instagram resmi'),
  ('contact', 'website_url',     '""', 'string', true, false, 'superadmin', 'Website resmi')

ON CONFLICT (setting_key) DO NOTHING;
```

### Permission Baru (modul Role-Permission)

```sql
INSERT INTO permissions (key, display_name, description, scope, category, risk_level) VALUES
  ('manage_system_settings', 'Manage System Settings',
   'Mengubah konfigurasi global platform', 'global', 'system', 'high'),
  ('manage_legal_documents', 'Manage Legal Documents',
   'Mengelola dokumen legal & versinya', 'global', 'system', 'medium')
ON CONFLICT (key) DO NOTHING;
```

---

## Query Patterns

### Load semua settings ke cache (saat boot / setelah invalidate)

```sql
SELECT group_key, setting_key, value, value_type
FROM system_settings;
```

### Public config untuk mobile

```sql
SELECT group_key, setting_key, value
FROM system_settings
WHERE is_public = true
ORDER BY group_key, setting_key;
```

### Update satu setting (selalu lewat service: validasi type + editable_by + audit)

```sql
UPDATE system_settings
SET value = :new_value, updated_by = :actor_id, updated_at = NOW()
WHERE setting_key = :key;
-- Setelah commit: invalidate cache + insert audit log (old vs new).
```

### Versi legal published + cek pending re-acceptance user

```sql
-- Versi aktif per doc_type
SELECT d.doc_type, d.title, v.id AS version_id, v.version, v.content, v.effective_date
FROM legal_documents d
JOIN legal_document_versions v
  ON v.document_id = d.id AND v.status = 'published'
WHERE d.doc_type = :doc_type;

-- Dokumen yang butuh persetujuan ulang oleh user (dipanggil saat login)
SELECT d.doc_type, v.id AS version_id, v.version
FROM legal_documents d
JOIN legal_document_versions v
  ON v.document_id = d.id AND v.status = 'published'
WHERE v.require_reacceptance = true
  AND NOT EXISTS (
    SELECT 1 FROM user_legal_acceptances a
    WHERE a.user_id = :user_id AND a.version_id = v.id
  );
```

### Catat acceptance (idempotent)

```sql
INSERT INTO user_legal_acceptances (user_id, version_id, source, ip_address)
VALUES (:user_id, :version_id, :source, :ip)
ON CONFLICT (user_id, version_id) DO NOTHING;
```

---

## Catatan Implementasi

1. **Tidak ada INSERT/DELETE setting saat runtime** — daftar key dikunci oleh seed migration. Endpoint backoffice hanya `PATCH` value. Menambah setting baru = migration baru.
2. **Validasi dua lapis di service layer:** (a) cocokkan tipe JSONB dengan `value_type`; (b) validasi khusus per key — `smtp_port` 1–65535, semver valid untuk `mobile_app.*version*`, `min_version <= latest_version`, `default_language` ada di `system_languages`, dsb.
3. **Cache:** load semua ke Redis/in-memory saat boot; UPDATE → invalidate → reload. Fallback TTL 5 menit untuk multi-instance.
4. **Audit Log:** setiap UPDATE setting & transisi status legal version dicatat (`actor_id`, key, old, new, ip). Nilai `is_sensitive = true` dimask di log dan response list.
5. **Secrets:** `SMTP_PASSWORD`, kredensial S3/R2, Firebase service account, Google OAuth secret — semuanya env/secret manager. Tabel ini tidak menyimpan secret apa pun.
6. **Konsumsi lintas modul:** Auth membaca `security.*`; Reporting membaca `moderation.report_*`; Subscription membaca `payment.*`; Upload service membaca `storage.*` — semuanya lewat satu settings service/cache, bukan query langsung.
7. **`legal_document_versions.content` immutable setelah published** — enforce di service layer (tolak UPDATE content jika status != 'draft'); bisa ditambah trigger DB jika ingin hard-guarantee.

---

*Skema ini melengkapi `SYSTEM_SETTINGS_RULES.md`. API spec backoffice & mobile menyusul.*
