# Plan: Observability Infra (Logs) — Loki + Alloy + Grafana + RequestID

## Context

k-forum-api sudah pakai `log/slog` (JSON di production, text di dev) tapi belum ada request ID untuk korelasi antar log satu request. Belum ada infra log aggregation sama sekali (tidak ada Loki/Grafana/Prometheus). Tujuannya: bikin stack observability logs sendiri (compose terpisah), tambahkan request ID ke semua log, dan bisa search log per-request di Grafana — termasuk log dari `worker` yang memproses job async akibat request itu. Metrics (Prometheus) sengaja TIDAK diimplementasi sekarang — cukup dipastikan arsitekturnya tidak menghalangi penambahan metrics nanti (Alloy dipilih karena bisa nambah `prometheus.scrape` belakangan tanpa ganti collector).

Keputusan yang sudah dikonfirmasi user:
- Fokus logs dulu, metrics future-ready (bukan diimplementasi sekarang).
- Compose file baru diletakkan di `k-forum-api/`, hanya observe container k-forum-api sendiri (tidak observe k-forum-backoffice).
- Request ID harus full correlation sampai ke worker: satu `request_id` yang sama harus bisa dipakai untuk search log API request DAN log worker yang memproses job async akibat request tersebut (bukan sekadar "worker lognya juga masuk Grafana" tanpa korelasi).

**Poin arsitektur penting**: k-forum-api pakai transactional outbox pattern — usecase tidak publish langsung ke RabbitMQ, tapi `outboxRepo.Save(ctx, outbox.NewOutboxEntry(...))` dalam transaksi DB yang sama dengan write bisnis (`internal/app/usecase/**`, ~40 lokasi). Publish AMQP yang sebenarnya baru terjadi belakangan, di proses **worker**, lewat outbox relay (`internal/interfaces/mq/relay/outbox_relay.go:88`). Konsekuensinya: request_id dari HTTP handler tidak bisa langsung "ikut" ke pesan AMQP — harus dititipkan lewat kolom di tabel `event_outbox`, lalu diambil lagi oleh relay saat publish, dan di-inject ulang ke context oleh worker saat consume.

## Bagian A — Kode Go (request ID + logger)

1. **`internal/logger/context.go`** (baru) — helper context: `WithRequestID(ctx, id) context.Context` dan `RequestIDFromContext(ctx) (string, bool)`, pakai unexported `ctxKeyRequestID` type sebagai key.

2. **`internal/logger/context_handler.go`** (baru) — `contextHandler` yang membungkus `slog.Handler`. Di `Handle(ctx, record)`, ambil request_id dari ctx via `RequestIDFromContext`, panggil `record.Clone()` dulu (wajib, supaya tidak mutate shared backing array punya record asli), baru `AddAttrs(slog.String("request_id", id))`. `WithAttrs`/`WithGroup` wajib mengembalikan `contextHandler` baru yang membungkus hasil delegasi, bukan handler polos, supaya wrapping tidak hilang kalau nanti ada kode pakai `logger.With(...)`. Efeknya: SEMUA pemanggilan `logger.InfoContext/WarnContext/ErrorContext/LogAttrs(ctx, ...)` di manapun (termasuk usecase yang meneruskan ctx) otomatis dapat `request_id` tanpa perlu ubah tiap call site.

3. **`internal/logger/logger.go`** (ubah) — ubah signature jadi `New(env, format string) *slog.Logger`. Kalau `format` diisi eksplisit ("json"/"text") pakai itu; kalau kosong, fallback ke behavior lama (json untuk production, text untuk lainnya). Bungkus handler akhir dengan `&contextHandler{next: baseHandler}` sebelum `slog.New(...)`.

4. **`internal/config/config.go`** — tambah field `LogFormat string` di `AppConfig` (dekat baris ~74), isi via `getEnv("LOG_FORMAT", "")` (dekat baris ~128).

5. Update 2 call site yang ada: **`cmd/app/main.go:89`** dan **`cmd/worker/main.go:58`** — ubah jadi `applogger.New(cfg.App.Env, cfg.App.LogFormat)`. (`cmd/seeder` dan `cmd/migrate` tidak pakai logger ini, tidak perlu diubah.)

6. **`internal/interfaces/http/middleware/request_id.go`** (baru) — Gin middleware `RequestID()`:
   - Server **selalu generate ID sendiri** via `uuid.NewString()` (package `github.com/google/uuid` sudah jadi dependency) — ini yang jadi `request_id` resmi dipakai untuk logging/korelasi/response header. Input client (`X-Request-Id` dari request) **tidak pernah dipakai/di-trust** sebagai request_id internal — menghindari risiko client bisa "memilih" ID-nya sendiri (spoofing/collision/log injection kalau ID itu dipakai mentah di query LogQL atau tempat lain).
   - Kalau client memang kirim header `X-Request-Id` di request, echo balik apa adanya di response header terpisah `X-Client-Request-Id` (murni untuk kenyamanan debugging client mencocokkan ID mereka sendiri) — bukan digabung/dipakai sebagai `request_id` server.
   - `c.Set("request_id", id)`.
   - `c.Writer.Header().Set("X-Request-Id", id)` (dan `X-Client-Request-Id` kalau ada input client) — **wajib sebelum** `c.Next()`, karena Gin flush header pada write pertama di handler chain; kalau di-set setelah `c.Next()` sudah terlambat untuk response sukses/4xx/5xx/panic sekalipun.
   - `c.Request = c.Request.WithContext(logger.WithRequestID(c.Request.Context(), id))`.
   - `c.Next()`.

7. **`internal/interfaces/http/router/router.go`** — daftarkan `middleware.RequestID()` sebagai `r.Use(...)` PALING PERTAMA (sebelum `CORSMiddleware()`), supaya request_id sudah ada untuk semua middleware setelahnya termasuk panic recovery di `httperror.Middleware`.

8. **`internal/interfaces/http/middleware/cors.go`** — tambahkan `"X-Request-Id"` ke `AllowHeaders` (supaya browser boleh kirim header ini di request cross-origin, meski server tidak men-trust nilainya) dan `"X-Request-Id"` + `"X-Client-Request-Id"` ke `ExposeHeaders` (supaya JS di browser bisa baca kedua response header ini). Tanpa ini, request ID praktis tidak berguna untuk client browser.

9. **`internal/interfaces/http/httperror/middleware.go`** — ubah `logger.Error(...)` di panic-recovery jadi `logger.ErrorContext(c.Request.Context(), ...)` supaya log panic juga ikut ke-tag request_id lewat context handler (sebelumnya pakai varian non-context sehingga tidak akan otomatis dapat request_id).

10. **`internal/interfaces/http/middleware/request_logger.go`** — TIDAK perlu diubah, karena sudah pakai `logger.LogAttrs(c.Request.Context(), ...)` yang otomatis dapat request_id lewat context handler.

11. **`.env.example`** — tambah `LOG_FORMAT=json`. Ini penting: tanpa ini, di dev (`APP_ENV=development`) log tetap pakai TextHandler, dan Grafana LogQL `| json` tidak bisa extract `request_id` dari log text. JSON format wajib aktif supaya fitur search-by-requestID di Grafana benar-benar jalan di local dev.

**Verifikasi kode**: `go build ./...` dan jalankan test handler yang ada (`go test ./internal/interfaces/http/handler/...`) untuk pastikan tidak ada regresi dari perubahan signature `logger.New`.

## Bagian B — Korelasi request_id sampai ke Worker (Outbox → RabbitMQ → Worker)

Dasarnya sudah dibangun di Bagian A (helper `logger.WithRequestID`/`RequestIDFromContext` + `contextHandler`). Bagian ini memakai helper yang **sama** (nama field tetap `request_id`, bukan field baru "correlation_id" di level log) supaya query Grafana `request_id="<id>"` bisa nemuin log API maupun log worker sekaligus.

1. **`internal/app/service/outbox/outbox_entry.go`** — tambah field `CorrelationID string` di struct `OutboxEntry`. Ubah constructor jadi `NewOutboxEntry(ctx context.Context, routingKey string, payload json.RawMessage) *OutboxEntry`: ambil `id, ok := logger.RequestIDFromContext(ctx)`; kalau tidak ada (job yang tidak berasal dari HTTP request, misal scheduler/cron), generate `uuid.NewString()` sebagai fallback — supaya tetap traceable meski bukan dari request HTTP.

2. **Migration baru** (jangan edit migration `0014_create_system_tables.up.sql` yang sudah ada — buat file migration baru sesuai nomor urut berikutnya) — tambah kolom nullable `correlation_id VARCHAR(64)` ke tabel `event_outbox`.

3. **`internal/infrastructure/persistence/postgres_outbox_repository.go`** — update `Save` (baris ~27-37) untuk ikut menyimpan `correlation_id`, dan `scanOutboxEntry` (baris ~90-110) untuk membacanya kembali.

4. **42 call site pemanggil `outbox.NewOutboxEntry(...)`** (dihitung eksak via `grep -rn "outbox.NewOutboxEntry(" --include="*.go" | wc -l` — tersebar di 37 file, sebagian file punya lebih dari 1 pemanggilan) di `internal/app/usecase/**`, `internal/app/service/notification/dispatcher.go`, dan `internal/interfaces/mq/relay/scheduler_relay.go` — perubahan mekanis: tambahkan `ctx` sebagai argumen pertama (ctx sudah tersedia di scope yang sama karena baris berikutnya selalu `outboxRepo.Save(ctx, ...)`). Jalankan ulang grep yang sama persis sebelum eksekusi untuk pastikan angka belum berubah karena commit lain. Lokasi yang tidak berasal dari HTTP context (`scheduler_relay.go`, `internal/app/usecase/newssource/trigger_scrape_now.go`) akan dapat correlation ID sintetis dari fallback di poin 1.

5. **`internal/interfaces/mq/message.go`** — tambah field `CorrelationID string \`json:"correlation_id"\`` ke struct `JobMessage`.

6. **`internal/interfaces/mq/relay/outbox_relay.go:80-86`** — saat membangun `mq.JobMessage{...}` dari outbox entry yang mau dipublish, ikut set `CorrelationID: entry.CorrelationID`.

7. **`internal/interfaces/mq/middleware/job_lifecycle.go:34-38`** — ini titik krusial di sisi worker. Saat ini baris 35 langsung `ctx := context.Background()` sebelum unmarshal `msg`. Ubah urutan: unmarshal `msg` dulu, baru bangun `ctx := logger.WithRequestID(context.Background(), msg.CorrelationID)`. Pakai `ctx` yang sudah dibungkus ini untuk semua pemanggilan setelahnya di fungsi ini (`jobRepo.FindByID/Save/Update`, dan yang paling penting `registry.Dispatch(ctx, msg)` di baris 66) — karena `registry.Dispatch`/`HandlerFunc` sudah meneruskan `ctx` ke semua domain handler, satu titik injeksi ini otomatis menyebar ke semua consumer tanpa ubah signature handler lain.

8. **`internal/interfaces/mq/middleware/job_lifecycle.go`** (lanjutan) — ubah semua pemanggilan `logger.Info/Warn/Error(...)` di file ini (baris ~54, 63, 76, dst.) jadi varian `*Context` (`InfoContext`/`WarnContext`/`ErrorContext`) supaya `request_id` yang sudah dititip di `ctx` benar-benar ikut ter-attach lewat `contextHandler` — kalau tetap pakai varian non-context, log tidak akan dapat atribut `request_id` walau ctx-nya sudah benar.

9. Domain handler di `internal/interfaces/mq/handler/**` yang melakukan logging sendiri — pastikan juga pakai `logger.*Context(ctx, ...)` (bukan varian polos) di titik manapun mereka log, supaya konsisten ikut ter-tag `request_id`.

**Catatan scope**: pendekatan ini menitipkan correlation ID lewat JSON envelope (`JobMessage.CorrelationID`), bukan lewat native AMQP header/`CorrelationId` field — lebih sederhana, tidak perlu ubah signature `port.EventPublisher.Publish` maupun `RabbitMQPublisher`. Cukup untuk kebutuhan logging/observability saat ini.

**Verifikasi kode**: `go build ./...` — perhatikan compile error di 42 call site `NewOutboxEntry` (signature berubah), pastikan semua sudah di-update. Jalankan test yang menyentuh usecase terkait outbox kalau ada.

## Bagian C — Docker Compose Observability (baru, file terpisah)

Lokasi semua file baru di dalam `k-forum-api/`:

- **`docker/observability/loki/loki-config.yaml`** — konfigurasi Loki single-binary, storage filesystem, schema TSDB (v13), retention 7 hari (`limits_config.retention_period: 168h`, bisa disesuaikan). **Penting, sering kelewat**: `retention_period` saja TIDAK cukup untuk benar-benar menghapus data lama — di single-binary Loki, penghapusan chunk lama dieksekusi oleh compactor, dan itu harus diaktifkan eksplisit lewat `compactor.retention_enabled: true` (plus `compactor.delete_request_store: filesystem` untuk storage filesystem) di block `compactor:`. Kalau cuma set `retention_period` tanpa `compactor.retention_enabled: true`, data akan terus menumpuk tanpa pernah kehapus meski sudah lewat 7 hari. Validasi dengan `docker compose run --rm loki -config.file=... -verify-config` sebelum dipakai, karena field schema berubah antar versi Loki.

- **`docker/observability/alloy/config.alloy`** — pipeline Grafana Alloy (River syntax): `discovery.docker` (baca `/var/run/docker.sock`) → `discovery.relabel` (filter container yang namanya match regex `^/?(k-forum-api_.*)` — filter berdasarkan `container_name:` yang sudah ada, tidak perlu ubah `docker-compose.dev.yml`) → `loki.source.docker` (tail log via Docker API) → `loki.write` (push ke `http://loki:3100/loki/api/v1/push`).

- **`docker/observability/grafana/provisioning/datasources/loki.yaml`** — provisioning datasource Loki otomatis (`type: loki`, `url: http://loki:3100`, `isDefault: true`).

- **`docker-compose.observability.yml`** (di root `k-forum-api/`) — service `loki`, `alloy`, `grafana`:
  - Semua di network baru `observability-net` (bridge, dideklarasikan lokal di file ini). **Tidak perlu** join `k-forum-net` — Alloy baca log container lewat Docker socket/API (seperti `docker logs`), bukan lewat jaringan container-to-container, jadi tidak butuh network yang sama.
  - `alloy` mount `/var/run/docker.sock:ro`. Catatan: `:ro` cuma cegah tulis ke file socket, bukan membatasi API call yang bisa dilakukan Alloy (akses setara root ke Docker Engine) — ini risiko yang diterima untuk stack observability lokal/dev, bukan sesuatu yang benar-benar dimitigasi oleh `:ro`.
  - `grafana` port `${GRAFANA_EXPOSE_PORT:-3030}:3000` — **bukan** default 3000, karena `k-forum-backoffice`'s `web-production` service sudah pakai host port 3000, akan collide kalau default dipakai.
  - `loki` port `${LOKI_EXPOSE_PORT:-3100}:3100` — 3100 aman, tidak dipakai project lain.
  - Volume `loki_data`, `grafana_data`.
  - Ikuti gaya file `docker-compose.dev.yml` yang sudah ada: tanpa `version:` key, `${VAR:-default}` interpolation, healthcheck blocks, `container_name: k-forum-api_loki` / `_alloy` / `_grafana`.

- **`.env.observability.example`** (di root `k-forum-api/`) — dokumentasikan `LOKI_PORT`/`LOKI_EXPOSE_PORT`, `GRAFANA_PORT`/`GRAFANA_EXPOSE_PORT`, `GRAFANA_ADMIN_USER`, `GRAFANA_ADMIN_PASSWORD` (placeholder, bukan secret asli).

**Antisipasi risiko (dieksekusi sebagai bagian dari task ini, bukan sekadar catatan):**

- **Log rotation di `docker-compose.dev.yml`** — Docker default `json-file` log driver tidak punya rotation limit, jadi file log container bisa membesar tanpa batas di host, terpisah dari retention Loki (yang cuma retensi data DI Loki, bukan file mentah `json-file` di host). Untuk antisipasi ini, tambahkan block berikut ke service `app`, `worker`, dan service jangka-panjang lain (`postgres`, `rabbitmq`, `opensearch`, `redis`, `minio`) di `docker-compose.dev.yml`:
  ```yaml
  logging:
    driver: json-file
    options:
      max-size: "10m"
      max-file: "3"
  ```
  Ini hanya menambah config `logging:` per service, tidak mengubah topologi/network/volume yang sudah ada, jadi aman dan tetap konsisten dengan gaya file yang ada.

- **Duplikasi log saat panic** — `httperror.Middleware` akan tetap menghasilkan 2 baris log untuk 1 panic yang sama (`"panic recovered"` dari `httperror.Middleware` dan `"server error"` dari `RequestLogger`). Ini **disengaja, bukan bug**: baris pertama membawa detail teknis (`recovered` value, stack context) untuk diagnosis, baris kedua adalah ringkasan standar tiap request (method/path/status/latency) yang konsisten untuk semua request termasuk yang sukses. Antisipasinya: pastikan kedua baris membawa `request_id` yang sama (sudah dijamin oleh fix di Bagian A poin 9 — `logger.ErrorContext`), sehingga di Grafana operator cukup search 1 `request_id` untuk melihat kedua baris berdampingan, bukan bingung menganggap itu 2 error terpisah. Tidak perlu dedupe/hilangkan salah satu baris karena keduanya punya informasi berbeda dan saling melengkapi.

## Cara Search by RequestID di Grafana

Setelah stack jalan (`docker compose -f docker-compose.observability.yml up -d`), buka Grafana Explore, pilih datasource Loki, query:

```
{container=~"k-forum-api.*"} | json | request_id="<id>"
```

Ambil `<id>` dari response header `X-Request-Id` yang dikembalikan API ke client, lalu tempel ke query di atas untuk lihat semua log (termasuk panic/error dari API, DAN log worker yang memproses job async akibat request tersebut — karena regex container sudah mencakup `k-forum-api_worker` dan field `request_id`-nya sama).

## Verifikasi End-to-End

1. `go build ./...` — pastikan kompilasi sukses setelah semua perubahan signature (termasuk ~40 call site `NewOutboxEntry` yang nambah param `ctx`).
2. Jalankan `docker compose -f docker-compose.dev.yml up -d` (pastikan `LOG_FORMAT=json` di `.env`), hit salah satu endpoint API, cek response header `X-Request-Id` ada.
3. Jalankan `docker compose -f docker-compose.observability.yml up -d`, buka Grafana di `http://localhost:${GRAFANA_EXPOSE_PORT:-3030}`, pastikan datasource Loki sudah keprovisioning otomatis.
4. Di Grafana Explore, jalankan query LogQL di atas dengan request_id dari langkah 2, pastikan muncul log request tersebut (termasuk baris dari `request_logger.go`).
5. Trigger error 500 (misal lewat endpoint yang sengaja panic atau lewat test), pastikan 2 baris log (panic + server error) sama-sama muncul dengan `request_id` yang sama saat di-search.
6. Hit endpoint yang memicu outbox event (misal create post/register/dll yang menulis ke `event_outbox`), catat `request_id`-nya dari response header. Tunggu worker memproses job (cek `event_outbox` row sudah ke-relay), lalu search `request_id` yang sama di Grafana — pastikan muncul juga log dari container `k-forum-api_worker` untuk job tersebut, membuktikan korelasi API → outbox → RabbitMQ → worker berjalan end-to-end.
