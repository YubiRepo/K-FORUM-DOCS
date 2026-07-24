# Brainstorming: Multi-language (i18n) & Localization (l10n) — Platform K-Forum

> Dokumen eksploratif/diskusi. Untuk rencana eksekusi konkret lihat
> [`PLAN_I18N_L10N_PLATFORM.md`](./PLAN_I18N_L10N_PLATFORM.md) (dokumen hidup,
> di-update seiring progres). Untuk translasi pesan error API (sudah
> diimplementasikan, scope sempit), lihat
> [`../PLAN_I18N_L10N.md`](../PLAN_I18N_L10N.md).

## Latar Belakang

Review arsitektur menyeluruh untuk i18n/l10n di seluruh platform (k-forum-api,
backoffice, mobile app k_forum), dipicu oleh kebutuhan mengantisipasi
pertumbuhan jumlah locale (saat ini id/en/ko, berpotensi tumbuh ke 5-10+).
Bukan dipicu insiden spesifik — ini architecture review umum.

## Kondisi Saat Ini per Area

### 1. Translasi konten dinamis (database)

Hanya **News** yang punya mekanisme translasi konten, lewat tabel
`article_translations` (`internal/migrations/0012_create_news_tables.up.sql`):
kolom `article_id, language, title, content, summary, is_original,
translate_status, translated_by, provider_requested`, `UNIQUE(article_id,
language)`. Pipeline translasi multi-provider (`noop/google/aws/llm`) sudah
matang lewat `ProviderRegistry` (`internal/infrastructure/translation/registry.go`),
mendukung mode batch & sync.

Modul lain — Event, Community, Directory, QnA, Ads — **tidak punya mekanisme
translasi konten sama sekali**. Konten disimpan single-locale langsung di
tabel utama.

### 2. String statis UI — 3 mekanisme, duplikasi nyata di 2 dari 3

| Layer | Mekanisme | Cakupan | Isi |
|---|---|---|---|
| API (Go) | Bundle JSON custom (`internal/infrastructure/i18n/bundle.go`), fallback `lang → default → key` | Hanya pesan error/validasi | Murni kode error (`DOMAIN_AD_HEADLINE_REQUIRED`, `CONTENT_BANNED_KEYWORD`, dst) — **tidak overlap** secara konsep dengan label UI |
| Backoffice (Vue) | Composable + dictionary TS custom (`app/composables/useI18n.ts`, `en.ts/id.ts/ko.ts` ~77 baris) | Sangat terbatas — nav, login, preferences saja; mayoritas layar fitur hardcode English | `common.save: 'Simpan'`, `common.cancel: 'Batal'`, `common.delete: 'Hapus'`, `common.edit: 'Ubah'`, dll |
| Mobile (Flutter) | `flutter_localizations` + ARB resmi (`lib/l10n/app_{en,id,ko}.arb`, ~800+ key) | Penuh, seluruh app | `commonSave: "Simpan"`, `commonCancel: "Batal"`, `commonDelete: "Hapus"`, `commonEdit: "Ubah"`, dll |

Koreksi dari klaim awal: API bundle isinya murni kode error backend, jadi
**tidak** benar-benar tumpang tindih konsep dengan label UI seperti "Simpan"/
"Batal" — API bundle bukan bagian dari masalah duplikasi ini.

Duplikasi nyata dan terverifikasi ada **di antara backoffice dan mobile**:
kata "Simpan", "Batal", "Hapus", "Ubah" ditulis & diterjemahkan ke id/en/ko
secara **independen di 2 file berbeda** (`app/i18n/locales/id.ts` vs
`lib/l10n/app_id.arb`) — kebetulan masih sama persis hari ini, tapi tidak ada
mekanisme yang menjaga keduanya tetap sinkron kalau salah satu diubah/
diperbaiki di masa depan (mis. terjemahan Korea untuk "Edit" diperbaiki di
mobile tapi lupa di backoffice). Ini akar masalah nyata yang mendasari
Phase 2 (central source), bukan cuma "3 platform, 3 format" yang sebenarnya
wajar untuk stack teknologi berbeda.

### 3. Locale resolution — 2 alur tidak sinkron

**Alur A** — HTTP middleware (`internal/interfaces/http/middleware/locale.go`):
`?lang=` query → `Accept-Language` header → default `i18n.DefaultLocale()`
(=`"en"`). Tidak melihat preferensi tersimpan user. Validasi ke daftar
hardcoded `i18n.supportedLocales`.

**Alur B** — usecase user-settings (`internal/app/usecase/usersettings/helpers.go:
resolveLanguageEffective`): preferensi bahasa tersimpan di profil user →
header `X-Locale` (baru `Accept-Language`) → setting sistem `default_language`
(fallback konstanta `"ko"`). Validasi ke tabel `system_languages` (`is_active`).

Beda nyata: (1) Alur A mengabaikan preferensi user, Alur B mengutamakannya;
(2) header device beda (`Accept-Language` saja vs `X-Locale` dulu); (3)
**default value beda: `"en"` vs `"ko"`** — potensi bug tersembunyi. Akibatnya
user yang sama bisa dapat bahasa berbeda tergantung endpoint mana yang
memproses request-nya.

### 4. Timezone (wall-clock data) — pola data benar, rendering salah

Backend punya pola yang baik: `PlatformTimezone` value object
(`internal/domain/system/valueobject/platform_timezone.go`) + field timezone
IANA eksplisit per entity (event, schedule) — hasil kerja `PLAN_TIMEZONE_WALLCLOCK_INPUTS.md`.
Tapi rendering di client belum konsisten:
- **Mobile**: `EventFormatters` memanggil `.toLocal()` (timezone device),
  bukan timezone yang tersimpan di entity — field timezone cuma dipakai
  untuk input/validasi, bukan tampilan. Ini bug aktif.
- **Backoffice**: tidak ada konversi sama sekali — waktu ditampilkan mentah
  plus label timezone (`⏰ ${event_time} (${timezone})`), bukan dikonversi.

### 5. Search (OpenSearch)

Field paralel per locale (`title_id`, `title_en`, `body_id`, `body_en`) di
index News, analyzer default (bukan linguistik per bahasa — tidak ada nori
untuk Korea, stemmer Indonesia). Aman di 3 bahasa, tapi field mapping akan
meledak linear terhadap jumlah locale kalau tumbuh ke 5-10+.

### 6. Locale/language master data

`system_languages` (code, is_ui_language, is_translate_target, is_active) +
`system_settings.default_locale`/`default_timezone`. Tapi Go bundle punya
`supportedLocales` **hardcoded terpisah** — bukan baca dari `system_languages`.
Tidak ada kolom `supported_locales` eksplisit di `system_settings`.

### 7. Format angka/currency

Tidak locale-aware secara konsisten. Mobile: beberapa formatter (`date_formatter.dart`,
`currency_formatter.dart`) hardcode `'id_ID'`/`'en_US'` terlepas dari locale
aktif app; hanya `event_formatters.dart` yang benar-benar locale-aware (peta
locale app → intl locale). Tidak ada `initializeDateFormatting()` dipanggil
di manapun — berisiko untuk locale non-`en`.

### 8. Business rule yang locale-independent (by design)

Banned keywords, permission, validasi format — global, tidak per-locale. Ini
**sudah benar** sesuai kebutuhan bisnis, bukan gap.

## Klasifikasi

| Elemen | Klasifikasi |
|---|---|
| Label tombol, nav, pesan error | i18n — resource translation |
| Isi artikel/listing/deskripsi entity | l10n — data translation (database) |
| `system_languages`, `system_settings` | Master data / konfigurasi platform |
| Format tanggal/angka/currency | l10n locale-dependent, tapi resource-based (pakai library CLDR/intl, bukan diterjemahkan manual) |
| Timezone IANA per entity | Locale-dependent tapi BUKAN translasi — domain wall-clock/lokasi, jangan dicampur dengan i18n string |
| Banned keywords, permission | Business data, locale-independent by design |

## Opsi Desain & Trade-off

### Translasi konten dinamis

| Pendekatan | Query/perf | Evolusi skema | Integritas | Fit search saat ini | Fit ProviderRegistry | Biaya migrasi |
|---|---|---|---|---|---|---|
| (a) Tabel bespoke per modul (replikasi pola `article_translations`) | Terbaik | Sedang (N tabel, tapi mekanis) | Terkuat (FK asli) | Selaras | Tinggal daftar writer baru | Rendah |
| (b) Tabel polymorphic generik (`content_translations`) | Terburuk utk listing/fallback (perlu pivot row→kolom) | Termurah | Terlemah (no FK) | Perlu reshape indexer | Paling cocok secara shape | Tinggi kalau News ikut dimigrasi |
| (c) Kolom JSONB di tabel utama | Bagus utk single-read, kehilangan status/provenance | Murah utk bahasa, mahal utk field | Kehilangan audit trail | Perlu flatten | Perlu ubah row-write→JSONB-merge | Tinggi (regresi fungsional) |
| (d) Hybrid (terstruktur utk News/Event, EAV utk Directory/QnA) | Campuran | Baik | Campuran | Dua jalur indexer selamanya | Dua jalur logic selamanya | Sedang, maintenance ganda |

**Catatan skala locale**: (a) row-based per bahasa — menambah bahasa =
insert baris, bukan ubah skema. Jadi (a) sudah scalable terhadap **jumlah
bahasa** meski tumbuh ke 5-10+; yang butuh desain ulang justru search
indexing (lihat bawah), bukan struktur tabel translasi.

### String statis

- **Sederhana**: pertahankan 3 bundle terpisah (kondisi sekarang) — drift terus berlanjut.
- **Umum**: satu sumber JSON/YAML sentral, di-generate ke 3 format native (ARB, Go JSON, TS dict) lewat script build.
- **Enterprise**: vendor TMS (Phrase/Lokalise/Crowdin) sebagai source of truth, CI menarik key ke tiap repo.
- **Scalable**: kombinasi sumber sentral + konvensi key seragam lintas platform.

Untuk 3→5-10 bahasa dengan tim kecil, opsi **Umum** (central JSON + generator,
tanpa vendor) adalah titik keseimbangan biaya/manfaat terbaik.

### Search indexing

- **Sederhana (saat ini)**: field paralel per locale — berhenti scalable di locale banyak.
- **Umum**: field tunggal + attribute `language` + analyzer per keluarga bahasa yang benar-benar aktif dipakai.
- **Enterprise**: index terpisah per bahasa dengan analyzer dioptimalkan penuh, digabung lewat alias saat query.

Untuk roadmap 5-10+ bahasa, opsi **Umum** cukup — index-per-bahasa baru
sepadan kalau volume dokumen per bahasa besar dan relevansi jadi hard
requirement.

## Rekomendasi Arah (ringkas — detail eksekusi di PLAN_I18N_L10N_PLATFORM.md)

1. Bakukan pola row-based per-modul (opsi a) sebagai konvensi translasi konten platform, dengan helper generik untuk fallback & writer pipeline (reuse code tanpa penalti opsi b/c).
2. Satukan locale resolution jadi satu helper (pola serupa `ResolvePlatformTimezone` yang sudah terbukti baik untuk timezone).
3. Central source untuk string statis, di-generate ke 3 platform.
4. Redesain index search: field tunggal + language filter + analyzer per keluarga bahasa, sebelum locale bertambah signifikan.
5. Perbaiki bug timezone-display di mobile & backoffice (item terpisah, bukan keputusan desain baru — sudah bug produksi).

## Pertanyaan Terbuka

1. Urutan prioritas modul yang mendapat translasi konten (Event dulu? Directory dulu? semua sekaligus?) — **belum dikonfirmasi user**.
2. Currency: perlu multi-currency per locale, atau cukup satu currency dengan format angka locale-aware?
3. Central-JSON-source cukup, atau ada preferensi ke vendor TMS tertentu?
4. Apakah backoffice benar-benar perlu dilokalkan penuh, atau tetap English-only untuk admin chrome selamanya (hanya konten yang multi-bahasa)?
