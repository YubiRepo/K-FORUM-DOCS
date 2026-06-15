# Accounting Module — Rules & Use Cases (v1.1)

Dokumentasi sistem rules modul Accounting KAI App. Accounting adalah **buku besar pencatatan uang masuk & keluar (in/out)** untuk operasional KAI, dipakai di **backoffice oleh Superadmin (KAI Pusat) saja**.

> **Scope penting:** Modul ini adalah *ledger / pencatatan*, BUKAN payment gateway dan BUKAN sistem akuntansi penuh (neraca, PSAK, pajak). Modul lain (Subscription, Ads) menghasilkan uang; Accounting mencatat & merangkumnya, plus mencatat pengeluaran manual yang tidak punya modul sendiri.
>
> **Catatan v1.1:** Accounting **tidak menggunakan region**. Tidak ada scope/pemisahan per region dan tidak diakses Admin Region — murni tingkat Pusat oleh Superadmin.

---

## 1. APA ITU MODUL ACCOUNTING?

**Accounting** adalah sistem pencatatan keuangan internal KAI untuk melacak arus kas (cashflow) operasional Pusat.

### Karakteristik:
- **Surface:** Backoffice only (tidak ada di mobile member)
- **Pengguna:** Superadmin (KAI Pusat) saja
- **Fungsi inti:** Catat transaksi IN (pemasukan) / OUT (pengeluaran), kategorisasi, lihat saldo & laporan
- **Sumber data:** Manual (diketik Superadmin) + future auto-record dari modul lain

### Berbeda dari:
- **Subscription:** menghasilkan revenue (upgrade Pro). Accounting mencatat revenue tersebut.
- **Ads:** menghasilkan revenue (member bayar iklan). Accounting mencatat revenue tersebut.
- **Payment Gateway (future):** memproses pembayaran. Accounting mencatat hasil pembayaran yang sudah settle.

---

## 2. KENAPA MODUL INI ADA?

Saat ini revenue KAI tersebar dan tidak terpusat:
- Revenue Subscription hanya muncul sebagai angka MRR di dashboard subscription
- Revenue Ads tidak terangkum di satu tempat
- Pengeluaran operasional (sewa, gaji, biaya event) tidak tercatat sama sekali

Accounting mengisi gap ini: **satu tempat yang menjumlahkan semua uang masuk + keluar**, dengan kategorisasi yang rapi.

---

## 3. KONSEP INTI — TRANSACTION LEDGER

Setiap pencatatan = satu **entry** dengan arah eksplisit:

```
accounting_entry
  ├── direction      : IN (pemasukan) | OUT (pengeluaran)
  ├── amount         : nominal dalam currency asli
  ├── currency       : IDR (default), KRW, USD, dll
  ├── exchange_rate  : kurs ke IDR saat transaksi
  ├── amount_base    : amount × exchange_rate (dalam IDR, untuk agregasi)
  ├── category_id    : referensi ke accounting_categories
  ├── source         : manual | system
  ├── source_ref     : link balik ke transaksi asal (mis. subscription_request_id)
  ├── status         : recorded | verified | void
  ├── transaction_date: tanggal transaksi terjadi (bukan tanggal input)
  └── attachment_url : bukti (nota, kwitansi, screenshot transfer)
```

**Saldo** = `SUM(amount_base WHERE direction=IN) − SUM(amount_base WHERE direction=OUT)` (filter per period sesuai kebutuhan).

### Kenapa desain ini?
- **`direction`** → laporan in/out & saldo jadi query sederhana.
- **`source` + `source_ref`** → memisahkan entri manual dari entri otomatis (future auto-record), sekaligus jembatan integrasi.
- **`amount_base` (IDR)** → semua agregasi & laporan konsisten meski transaksi multi-currency.

---

## 4. KATEGORI (CATEGORY MASTER) — HIERARKIS

Kategori dikelola Superadmin, dipakai untuk mengelompokkan transaksi. Bersifat **fleksibel dan bertingkat (parent → child)** — Superadmin bisa tambah/edit/nonaktifkan kapan saja tanpa deploy ulang.

### Konsep hierarki
- **Parent category** (level 1) — kelompok besar, mis. "Operasional".
- **Child category** (level 2) — rincian di bawah parent, mis. "Listrik", "Internet", "ATK".
- Transaksi boleh di-assign ke parent **atau** ke child. Laporan bisa di-**roll-up** (jumlahkan semua child ke parent) atau di-**drill-down** (lihat per child).
- Maksimum 2 level (parent → child) untuk menjaga laporan tetap mudah dibaca. Child tidak punya child lagi.
- `direction` child **harus sama** dengan parent-nya (parent OUT → semua child OUT).

### Contoh struktur IN (pemasukan):
| Code | Nama | Parent | Keterangan |
|------|------|--------|-----------|
| `REV_SUBSCRIPTION` | Subscription Revenue | — | Pendapatan upgrade Pro |
| `REV_ADS` | Ads Revenue | — | Pendapatan member pasang iklan |
| `REV_EVENT` | Event Revenue | — | Pendapatan dari event |
| `REV_EVENT_TICKET` | Tiket | `REV_EVENT` | Penjualan tiket |
| `REV_EVENT_SPONSOR` | Sponsor | `REV_EVENT` | Dana sponsor event |
| `REV_DONATION` | Donation | — | Donasi/sumbangan |
| `REV_OTHER` | Other Income | — | Pemasukan lain |

### Contoh struktur OUT (pengeluaran):
| Code | Nama | Parent | Keterangan |
|------|------|--------|-----------|
| `EXP_SALARY` | Gaji & Honor | — | Gaji staff, honor narasumber |
| `EXP_RENT` | Sewa | — | Sewa kantor/venue |
| `EXP_EVENT` | Biaya Event | — | Operasional pelaksanaan event |
| `EXP_OPERATIONAL` | Operasional | — | Kelompok biaya operasional |
| `EXP_OP_ELECTRICITY` | Listrik | `EXP_OPERATIONAL` | |
| `EXP_OP_INTERNET` | Internet | `EXP_OPERATIONAL` | |
| `EXP_OP_SUPPLIES` | ATK & Perlengkapan | `EXP_OPERATIONAL` | |
| `EXP_MARKETING` | Marketing | — | Promosi, iklan eksternal |
| `EXP_OTHER` | Other Expense | — | Pengeluaran lain |

> Daftar di atas hanya **seed awal** — bukan daftar tetap. Superadmin bebas menambah parent/child baru sesuai kebutuhan.

### Aturan kategori
- Setiap kategori punya `code` standar (memudahkan export ke software akuntansi nanti).
- Kategori yang sudah dipakai transaksi **tidak bisa dihapus** — hanya dinonaktifkan (`is_active = false`) untuk jaga integritas laporan historis.
- Parent **tidak bisa dinonaktifkan** jika masih punya child aktif (nonaktifkan child dulu).
- `direction` dan `parent_id` tidak bisa diubah setelah kategori dipakai transaksi.

---

## 5. ACTOR & PERMISSION

Hanya **Superadmin (KAI Pusat)** yang mengakses modul ini. Tidak ada Admin Region.

| Aksi | Superadmin |
|------|:---:|
| Catat transaksi manual (IN/OUT) | ✅ |
| Lihat ledger | ✅ |
| Edit transaksi (sebelum `verified`) | ✅ |
| Void transaksi | ✅ |
| Verifikasi transaksi (jika diaktifkan) | ✅ |
| Kelola kategori master | ✅ |
| Lihat laporan & saldo | ✅ |
| Export | ✅ |
| Konfigurasi setting accounting | ✅ |

---

## 6. STATUS LIFECYCLE

```
recorded → verified        (jika verifikasi diaktifkan & disetujui Superadmin)
recorded → void            (dibatalkan, mis. salah input)
verified → void            (dengan alasan)
```

| Status | Deskripsi | Bisa diedit? |
|--------|-----------|:---:|
| `recorded` | Baru dicatat, default | ✅ |
| `verified` | Sudah diverifikasi | ❌ (terkunci, hanya bisa di-void) |
| `void` | Dibatalkan — tetap tersimpan untuk audit, tidak masuk perhitungan saldo | ❌ |

> **Verifikasi bersifat OPSIONAL.** Diatur via setting `verification_required`. Jika `false`, entri langsung final saat `recorded` (tetap bisa diedit/void). Jika `true`, entri perlu di-`verified` sebelum masuk laporan resmi.

---

## 7. USE CASES

### Use Case 1: Catat pengeluaran event
```
Superadmin → Accounting → "Catat Transaksi"
  ↓
Pilih: direction=OUT, category=EXP_EVENT, amount=Rp 5.000.000
       transaction_date=2026-06-10, attachment=foto kwitansi
  ↓
Submit → status=recorded
  ↓
(Jika verification_required) verifikasi → status=verified
```

### Use Case 2: Catat revenue subscription (manual, Phase 1)
```
Superadmin → Accounting → "Catat Transaksi"
  ↓
direction=IN, category=REV_SUBSCRIPTION, amount=Rp 80.000
source=manual, source_ref=req_123 (opsional, link ke subscription request)
  ↓
Submit → status=recorded
```

### Use Case 3: Lihat laporan cashflow
```
Superadmin → Accounting → Laporan
  ↓
Filter: period=Juni 2026
  ↓
Tampil: Total IN, Total OUT, Saldo, breakdown per kategori (roll-up / drill-down)
```

---

## 8. SETTING ACCOUNTING (BACKOFFICE, SUPERADMIN)

| Setting | Tipe | Default | Keterangan |
|---------|------|---------|-----------|
| `verification_required` | boolean | `false` | Jika true, entri perlu verifikasi sebelum final |
| `default_currency` | string | `IDR` | Mata uang default saat input |
| `require_attachment_for_out` | boolean | `false` | Wajib lampirkan bukti untuk transaksi OUT |
| `fiscal_year_start_month` | integer | `1` | Bulan awal tahun fiskal (untuk laporan tahunan) |

---

## 9. ATURAN BISNIS & EDGE CASES

- **Transaksi void tidak dihapus** — tetap tersimpan untuk jejak audit, hanya dikeluarkan dari perhitungan saldo & laporan.
- **Edit setelah verified** — tidak diizinkan. Harus void lalu buat entri baru (jaga integritas audit).
- **Multi-currency** — `amount_base` (IDR) selalu dihitung saat input pakai `exchange_rate` yang berlaku. Laporan agregat selalu pakai `amount_base`.
- **transaction_date vs created_at** — laporan keuangan pakai `transaction_date` (kapan transaksi terjadi), bukan `created_at` (kapan diinput). Penting untuk transaksi yang diinput terlambat.
- **Kategori nonaktif** — tidak muncul di pilihan input baru, tapi entri lama yang memakainya tetap valid.
- **Roll-up vs drill-down** — laporan per kategori bisa dijumlahkan ke level parent (roll-up) atau dirinci per child (drill-down). Transaksi yang di-assign langsung ke parent tetap dihitung di level parent.
- **Child harus sama arah dengan parent** — tidak boleh ada child OUT di bawah parent IN.

---

## 10. FUTURE-PROOF — INTEGRASI YANG DISIAPKAN

Pola: **simpan handle/hook sekarang, fungsionalitas menyusul tanpa refactor.** Entri accounting dirancang sebagai *event yang bisa di-push dari sumber manapun* — selama bentuknya konsisten (`direction`, `amount`, `category`, `source`, `source_ref`), semua masuk ke ledger yang sama.

| Integrasi masa depan | Yang sudah disiapkan sekarang |
|----------------------|-------------------------------|
| **Auto-record revenue** dari Subscription & Ads | Field `source`+`source_ref` + endpoint internal `POST /accounting/entries/internal` (dipanggil modul lain saat transaksi approve) |
| **Payment gateway** (Stripe/Midtrans — "coming soon" di Subscription) | Field `external_txn_id` + `payment_provider`, siap diisi webhook settlement |
| **Budgeting** | Tabel `accounting_budgets` (period + planned amount); entri dibandingkan vs budget |
| **Export ke software akuntansi** (Accurate, Jurnal, Xero) | Kategori pakai `code` standar + endpoint export CSV/JSON |
| **Audit / locking periode** | `status` (recorded→verified→void) + immutable record; bisa ditambah `locked` per periode |
| **Recurring entries** (gaji, sewa bulanan) | Tabel `recurring_templates` yang generate entri otomatis via cron |
| **Multi-currency & FX** | `currency` + `exchange_rate` + `amount_base` (sudah ada di v1.0) |
| **Bank reconciliation** | Field `reconciled` + `reconciled_at` (cocokkan ledger vs rekening) |

### Roadmap
- **Phase 1:** Input manual (IN/OUT), kategori hierarkis, laporan & saldo, multi-currency, export CSV. Field hook (`source`, `source_ref`, `external_txn_id`) disiapkan tapi belum dipakai.
- **Phase 2:** Auto-record dari Subscription & Ads (hook ke event approval). Verifikasi flow diaktifkan jika perlu.
- **Phase 3:** Budgeting, recurring entries, payment gateway settlement, bank reconciliation.

---

## 11. INTEGRATION POINTS

| Modul | Hubungan |
|-------|----------|
| **Subscription** | Sumber revenue (`REV_SUBSCRIPTION`). Future: auto-push entri saat upgrade di-approve, `source_ref=subscription_request_id`. |
| **Ads** | Sumber revenue (`REV_ADS`). Future: auto-push entri saat ads dibayar. |
| **Role/Permission** | Membatasi akses hanya ke Superadmin. |
| **Payment Gateway (future)** | Webhook settlement → auto-push entri dengan `external_txn_id`. |
| **System Settings** | Setting accounting (verification, currency, dll). |

---

## 12. SUMMARY

Accounting = **buku besar in/out terpusat** untuk operasional KAI di backoffice, diakses hanya oleh Superadmin (Pusat). Setiap entri adalah event yang bisa berasal dari input manual maupun (nanti) auto-push modul lain. Desain `direction` + `source/source_ref` + `amount_base` + kategori hierarkis membuatnya sederhana untuk dipakai sekarang tapi siap berkembang ke auto-record, payment gateway, budgeting, dan export — tanpa mengubah struktur inti.
