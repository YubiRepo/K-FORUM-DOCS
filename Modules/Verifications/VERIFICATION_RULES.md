# KAI Verification Badge Module — Rules & Use Cases (v1.0)

Dokumentasi sistem rules **Verification Badge** (centang keaslian) KAI App. Modul ini ngasih badge "Verified" ke **User**, **Merchant**, dan **Community** yang keasliannya udah diverifikasi manual oleh Superadmin. Ini **layer trust terpisah** — bukan konfirmasi kontak, bukan gate publish, bukan ID Card.

> **Prinsip inti:** verifikasi = pernyataan keaslian yang di-grant selektif oleh KAI Pusat. Diajukan manual, di-review manual, bisa dicabut. Semua riwayat append-only buat audit. Status badge di-resolve live dari record verifikasi terbaru, bukan snapshot yang gampang basi.

---

## 1. WHAT IS VERIFICATION BADGE?

**Verification Badge** adalah centang yang nempel di entitas (User atau Merchant) sebagai bukti bahwa **KAI Pusat sudah memverifikasi keasliannya**. Fungsinya:

1. **Trust signal** — kasih tau member lain bahwa akun/bisnis ini asli & terpercaya.
2. **Anti-impersonation** — bedain tokoh/bisnis asli dari akun palsu/tiruan.
3. **Kurasi** — bukan hak semua orang; hanya yang lolos review.

### Prinsip Desain (PALING PENTING)

| Prinsip | Penjelasan |
|---------|-----------|
| **Badge User ≠ Badge Merchant** | Dua tipe berbeda (`user` / `merchant`) — beda label, beda syarat dokumen, beda tampilan. Satu mekanisme review, dua produk. |
| **Resolve live, jangan snapshot** | Status verified di-resolve dari record verifikasi aktif terbaru. Kolom cache `is_verified` boleh ada buat query cepat, tapi record `verifications` adalah sumber kebenaran. |
| **Append-only history** | Setiap pengajuan/approve/reject/revoke tercatat permanen. Ga ada hard-delete record verifikasi — buat audit & anti-sengketa. |
| **Superadmin-only** | Cuma Superadmin yang approve & revoke. Ga didelegasikan ke Admin Regional (beda dari merchant approval). |
| **Manual-first, hook buat otomasi** | Phase 1 full manual. Field/hook disiapin buat auto-KYC / integrasi Phase 2, tapi ga dibangun sekarang. |

### Berbeda dari (WAJIB DIPAHAMI — jangan ketuker):

| Konsep | Menjawab | Beda dari Verification |
|--------|----------|------------------------|
| `email_verified` / `phone_verified` | "Kontak ini valid?" (OTP) | Cuma teknis konfirmasi kontak. Bukan trust, bukan selektif. |
| Merchant `approval_status` | "Boleh tampil di listing?" (gate publish) | Semua merchant layak bisa published. Verified = **layer di atasnya**, selektif. |
| ID Card | "Siapa orang ini?" (identitas, semua member punya) | Universal, bukan selektif. Verified = badge kurasi. |
| Verification Badge | **"Keaslian entitas ini udah dijamin KAI?"** | — |

---

## 2. ENTITY TYPES & BADGE

Tiga tipe, satu mekanisme:

| `type` | Target | Diajukan oleh | Label Badge | Contoh Kasus |
|--------|--------|---------------|-------------|--------------|
| `user` | akun member | member itu sendiri | "Verified Member" / centang biru | Tokoh publik, influencer diaspora, pengurus KAI, figur komunitas |
| `merchant` | listing bisnis di Directory | owner merchant | "Verified Merchant" / centang bisnis | Bisnis resmi, brand official, partner KAI |
| `community` | komunitas di modul Community | **Leader / owner** komunitas | "Verified Community" / centang komunitas | Komunitas resmi, chapter organisasi, komunitas partner KAI |

> **Ketiganya di-approve Superadmin** (global trust). Bedanya cuma **siapa yang ngajuin** — buat community, requester-nya Leader/owner komunitas (makanya sering disebut "diverifikasi lewat Leader", tapi keputusan tetep di Superadmin).
>
> Badge-nya **beda secara visual & makna**. Frontend bedain dari field `type`, bukan dari warna badge yang di-hardcode.

### Syarat entitas boleh diajukan

| `type` | Prasyarat | Siapa yang boleh ngajuin |
|--------|-----------|--------------------------|
| `user` | Akun `status = active`. Ga lagi suspended/banned. | User ybs |
| `merchant` | Merchant `approval_status = approved` **dan** `status = published`. Merchant draft/pending/rejected/banned **ga bisa** diajukan. | Owner merchant |
| `community` | Komunitas `status = active`. Ga lagi archived/suspended. | Leader/owner komunitas itu (butuh permission community, lihat §5) |

---

## 3. SYARAT DOKUMEN (rules-as-data)

Syarat dokumen per tipe **disimpan sebagai config** (`verification_requirements`), bukan di-hardcode — biar Superadmin bisa ubah tanpa deploy. Model penerimaannya beda per tipe:

- **User → `any_of` (fleksibel):** cukup lampirin **minimal 1** dokumen pendukung dari daftar yang diterima. **KTP TIDAK wajib** — banyak diaspora ga punya KTP Indonesia, dan trust badge komunitas ga perlu KYC pemerintah. Superadmin tetep yang nilai layak/enggak.
- **Merchant → `all_of` (lebih ketat):** dokumen legalitas usaha wajib, karena taruhannya lebih tinggi (transaksi/bisnis).
- **Community → `any_of` (fleksibel):** cukup **minimal 1** bukti legitimasi organisasi. Leader yang ngajuin atas nama komunitasnya.

Contoh default Phase 1:

| `type` | Model | Dokumen diterima | Catatan |
|--------|-------|------------------|---------|
| `user` | `any_of` (min 1) | KTA KAI (kartu anggota), KTA organisasi lain, kartu identitas lain (KTP/Paspor/SIM — **opsional**, bukan wajib), bukti jabatan/keanggotaan, link sosial media terverifikasi | Selfie/foto pemegang bersifat opsional, diminta Superadmin kalo perlu klarifikasi |
| `merchant` | `all_of` | NIB / akta perusahaan / izin usaha, foto identitas pemilik | Opsional: bukti alamat usaha, akun sosmed bisnis |
| `community` | `any_of` (min 1) | Akta/SK organisasi, bukti afiliasi/partner KAI, akun sosmed resmi komunitas, surat keterangan pengurus | Diajukan Leader; Superadmin verifikasi legitimasi komunitas |

> **KTA nyambung ke modul ID Card:** kartu anggota KAI yang udah ada bisa langsung dipakai sebagai bukti pendukung user. (Follow-up: pertimbangkan auto-attach data ID Card ke request user di Phase 2 — hook aja dulu.)

> Dokumen di-upload dulu via media service (`context: "verification"`), lalu URL-nya dilampirin ke request. **Dokumen sensitif (kartu identitas, NIB, dll) disimpan private, bukan CDN publik** — akses cuma buat Superadmin saat review. (Follow-up: konfirmasi bucket/policy sama tim storage.)

---

## 4. LIFECYCLE & STATUS

```
[User/Merchant owner] → Ajukan verifikasi (lampirkan dokumen)
        │
        ▼
   status: pending  ──────────────► notif ke Superadmin: "Ada pengajuan verifikasi baru"
        │
        │  [Superadmin review]
        │
        ├── Approve → status: approved
        │            → is_verified = true (cache di users/merchants)
        │            → badge tampil
        │            → notif ke pemohon: "Selamat, akun/bisnis lu udah terverifikasi"
        │
        └── Reject  → status: rejected + rejection_reason (wajib)
                     → is_verified tetap false
                     → notif ke pemohon: "Pengajuan ditolak: {reason}"
                     → pemohon boleh resubmit (bikin request baru)

[Superadmin] → Revoke (kapan aja, dari status approved)
              → status: revoked + revoke_reason (wajib)
              → is_verified = false
              → badge hilang
              → notif ke pemilik: "Verifikasi dicabut: {reason}"
```

### Status enum (`verifications.status`)

| Status | Arti | Badge tampil? |
|--------|------|:---:|
| `pending` | Diajukan, nunggu review | ❌ |
| `approved` | Disetujui Superadmin | ✅ |
| `rejected` | Ditolak, bisa resubmit | ❌ |
| `revoked` | Dicabut setelah pernah approved | ❌ |

### Aturan state

- **Satu entitas cuma boleh punya 1 request aktif** (`pending`) dalam satu waktu. Ga boleh spam ajuin.
- Kalo udah `approved`, ga bisa ajuin lagi (udah verified). Kecuali di-revoke dulu.
- `rejected` / `revoked` → boleh bikin request baru (resubmit). Request lama tetep tersimpan (append-only).
- Resubmit = **record baru**, bukan update record lama. Riwayat penolakan ga dihapus.

---

## 5. ACTOR & PERMISSION MATRIX

| Action | Pemohon (owner user/merchant, Leader komunitas) | Admin Regional | Superadmin |
|--------|:---:|:---:|:---:|
| Ajukan verifikasi entitas sendiri | ✅ | — | — |
| Lihat status pengajuan sendiri | ✅ | — | — |
| Resubmit setelah reject/revoke | ✅ | — | — |
| Lihat antrian pengajuan | ❌ | ❌ | ✅ |
| Lihat dokumen pengajuan | ❌ | ❌ | ✅ |
| Approve pengajuan | ❌ | ❌ | ✅ |
| Reject pengajuan | ❌ | ❌ | ✅ |
| Revoke badge | ❌ | ❌ | ✅ |
| Lihat riwayat verifikasi entitas | ❌ | 👀 (read, region sendiri) | ✅ |

> Admin Regional **cuma** boleh lihat (read) status verified entitas di region-nya buat monitoring — **ga bisa** approve/reject/revoke. Semua aksi kurasi di tangan Superadmin.
>
> Buat `community`: yang boleh ngajuin cuma **Leader/owner** komunitas itu (dicek via `user_roles` scope community + permission `verification.request_community`). Member biasa ga bisa ngajuin verifikasi komunitas.

### Permission keys (buat didaftarin di modul Role-Permission)

| Key | Deskripsi | Scope |
|-----|-----------|-------|
| `verification.request` | Ajukan verifikasi entitas sendiri (user/merchant) | member |
| `verification.request_community` | Ajukan verifikasi komunitas (khusus Leader) | community |
| `verification.review` | Approve/reject pengajuan | global (Superadmin) |
| `verification.revoke` | Cabut badge | global (Superadmin) |
| `verification.view_queue` | Lihat antrian + dokumen | global (Superadmin) |

---

## 6. BADGE DISPLAY RULES

- Badge tampil di: profil user, kartu merchant (listing + detail), **header/kartu komunitas (list + detail)**, author byline (news/announcement kalo relevan), komentar/post.
- Frontend nentuin ikon/warna dari `type` + `is_verified = true`.
- Kalo `is_verified = false` → **ga ada badge sama sekali** (ga ada state "pending badge").
- Public/member-facing response **cuma** expose `is_verified` + `verification_type`. **Jangan bocorin** dokumen, `rejection_reason`, `reviewer_id`, atau detail internal ke response member-facing.

---

## 7. NOTIFICATION EVENTS (buat didaftarin di modul Notification)

| Event | Penerima | Channel | Bypass? |
|-------|----------|---------|:---:|
| `verification.submitted` | Superadmin | In-app backoffice | ✅ |
| `verification.approved` | Pemohon | Push + In-app | ✅ (transaksional) |
| `verification.rejected` | Pemohon | Push + In-app | ✅ (transaksional) |
| `verification.revoked` | Pemilik | Push + In-app | ✅ (transaksional) |

> Semua bypass user-preference (transaksional/status akun, kaya pola Subscription).

---

## 8. USE CASES

### Use Case 1: User ajukan verified
```
Tokoh komunitas → Profile > Ajukan Verifikasi
  → Upload KTP + selfie, isi alasan
  → Submit → status: pending
Superadmin → Backoffice > Verification Queue
  → Buka dokumen, cek keaslian
  → Approve → badge muncul di profil user
```

### Use Case 2: Merchant ajukan verified
```
Owner bisnis (merchant udah published) → Merchant > Ajukan Verifikasi
  → Upload NIB + identitas pemilik
  → Submit → status: pending
Superadmin → review → Approve → "Verified Merchant" muncul di listing
```

### Use Case 3: Community ajukan verified (oleh Leader)
```
Leader komunitas (community status=active) → Community Settings > Ajukan Verifikasi
  → Upload akta/SK organisasi / bukti partner KAI
  → Submit → status: pending
Superadmin → review legitimasi komunitas → Approve
  → "Verified Community" muncul di header & list komunitas
```

### Use Case 4: Reject & resubmit
```
Superadmin → dokumen buram/ga valid → Reject (reason: "Foto KTA tidak terbaca")
  → Pemohon dapat notif + alasan
  → Pemohon upload ulang → Submit request baru → pending lagi
```

### Use Case 5: Revoke
```
Superadmin → temukan pelanggaran / bisnis tutup / impersonation / komunitas bubar
  → Revoke (reason wajib)
  → Badge hilang, is_verified = false, notif ke pemilik/Leader
  → Record revoke tersimpan permanen
```

---

## 9. DO / DON'T

### ✅ DO
- ✅ Simpan semua pengajuan append-only (audit trail lengkap).
- ✅ Resolve badge dari record aktif; cache `is_verified` cuma buat performa.
- ✅ Wajibin `rejection_reason` & `revoke_reason`.
- ✅ Simpan dokumen sensitif di storage private, akses cuma Superadmin.
- ✅ Bedain badge user / merchant / community lewat `type`.
- ✅ Batasin 1 request `pending` per entitas.

### ❌ DON'T
- ❌ Hard-delete record verifikasi.
- ❌ Bocorin dokumen / reason internal ke response member-facing.
- ❌ Ngasih Admin Regional hak approve/revoke.
- ❌ Ngeramu badge dari `email_verified`/`approval_status` — itu konsep beda.
- ❌ Snapshot status di banyak tempat; satu sumber kebenaran = tabel `verifications`.
- ❌ Verifikasi merchant yang belum `published` / komunitas yang ga `active`.
- ❌ Ngasih member biasa hak ngajuin verifikasi komunitas (cuma Leader).

---

## 10. CROSS-MODULE FOLLOW-UPS

Ditandai, dikerjain pas masuk modul terkait:

- **Role-Permission:** daftarin 5 permission keys (§5), termasuk `verification.request_community` (scope community, default Leader).
- **Notification:** daftarin 4 event (§7) + payload.
- **User Management:** `ALTER TABLE users ADD COLUMN is_verified BOOLEAN DEFAULT false` (additive).
- **Directory / Merchant:** `ALTER TABLE merchants ADD COLUMN is_verified BOOLEAN DEFAULT false` (additive).
- **Community:** `ALTER TABLE communities ADD COLUMN is_verified BOOLEAN DEFAULT false` (additive) + entry point "Ajukan Verifikasi" di Community Settings (khusus Leader).
- **Storage:** konfirmasi bucket private + akses policy buat dokumen verifikasi.
- **Media service:** tambah `context: "verification"`.

---

## 11. FUTURE FEATURES (OUT OF SCOPE Phase 1)

- ❌ **Verifikasi member scoped per-komunitas** (Leader nge-verify member di dalam komunitasnya, badge cuma berlaku di komunitas itu). Butuh dimensi `scope_type`/`scope_id` di `verifications` + cache di `community_members`. **Sengaja ditunda** — badge global (user/merchant/community, semua Superadmin) dijaga tetap seragam & bernilai. Baru dibangun kalau ada kebutuhan konkret.
- ❌ Auto-KYC / integrasi verifikasi identitas pihak ketiga.
- ❌ Tier badge (mis. "Official Organization" vs "Verified Individual").
- ❌ Delegasi review ke Admin Regional.
- ❌ Auto-expiry / re-verification berkala.
- ❌ Badge berbayar / di-gate benefit key subscription (hook disiapin, ga dibangun).
- ❌ Banding (appeal) terstruktur atas penolakan.

---

## 12. SUMMARY

Verification Badge = **trust signal selektif yang di-grant manual oleh Superadmin**. Diajukan pemohon (user sendiri / owner merchant / Leader komunitas) → di-review → approve/reject/revoke, semua append-only. Badge **User, Merchant, & Community** = 3 produk (label, dokumen, tampilan beda) tapi **satu mesin** (tabel `verifications`, dibedakan `type`). Semua di-approve Superadmin. Status di-resolve live dari record aktif; `is_verified` cuma cache. Terpisah tegas dari `email_verified`, merchant approval, dan ID Card.

---

*Verification Badge Module Rules v1.0 — KAI App. Step 1 dari pipeline (RULES → DB Schema → API Spec). Last updated: 2026-07-13*
