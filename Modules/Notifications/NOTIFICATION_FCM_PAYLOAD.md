# FCM Payload Pattern

Dokumen ini mendefinisikan struktur payload FCM yang dikirim backend (Golang) ke Flutter untuk setiap event notifikasi. Ini adalah kontrak antara backend dan Flutter — wajib konsisten di kedua sisi.

---

## Daftar Isi

- [Message Type Strategy](#message-type-strategy)
- [Struktur Payload Standar](#struktur-payload-standar)
- [Field Reference](#field-reference)
- [Payload Per Modul](#payload-per-modul)
- [Flutter Handling Guide](#flutter-handling-guide)

---

## Message Type Strategy

FCM mendukung dua tipe message. Kita pakai keduanya tergantung kondisi:

| Kondisi | Message Type | Alasan |
|---|---|---|
| App di **background / terminated** | `notification` + `data` | FCM otomatis tampilkan notif di system tray. Flutter handle tap via `data`. |
| App di **foreground** | `data-only` | FCM tidak tampilkan notif otomatis. Flutter yang render (misal: in-app snackbar / banner). |

**Backend tidak tahu app sedang foreground atau tidak** — jadi backend selalu kirim keduanya (`notification` + `data`). Flutter yang memutuskan cara render berdasarkan app state.

```
Backend kirim:
{
  "notification": { "title": "...", "body": "..." },   ← untuk background/terminated
  "data": { "module": "...", "event_type": "...", ... } ← selalu ada, untuk Flutter logic
}

Flutter terima:
- Background/terminated → FCM auto-render notif, Flutter handle tap dari data
- Foreground → Flutter intercept, gunakan data untuk render in-app UI
```

---

## Struktur Payload Standar

Semua notifikasi menggunakan struktur yang sama. Field `data` selalu flat key-value string (FCM requirement).

### Backend → FCM (yang dikirim backend)

```json
{
  "notification": {
    "title": "string",
    "body": "string",
    "image": "string | null"
  },
  "data": {
    "module": "string",
    "event_type": "string",
    "entity_id": "string",
    "click_action": "string",
    "bypass": "string (true | false)",
    "extra": "string (JSON encoded, nullable)"
  },
  "android": {
    "priority": "high | normal",
    "notification": {
      "channel_id": "kai_notifications",
      "sound": "default"
    }
  },
  "apns": {
    "payload": {
      "aps": {
        "sound": "default",
        "badge": 1
      }
    },
    "headers": {
      "apns-priority": "10 | 5"
    }
  }
}
```

### Flutter terima (dari `RemoteMessage`)

```dart
RemoteMessage {
  notification: RemoteNotification {
    title: "string",
    body: "string",
    imageUrl: "string | null"
  },
  data: {
    "module": "string",
    "event_type": "string",
    "entity_id": "string",
    "click_action": "string",
    "bypass": "true | false",
    "extra": "string (JSON encoded) | null"
  }
}
```

---

## Field Reference

### `notification` block

| Field | Type | Keterangan |
|---|---|---|
| `title` | string | Judul notifikasi |
| `body` | string | Isi notifikasi, max 300 karakter |
| `image` | string \| null | URL thumbnail. Hanya untuk notif yang punya visual relevan |

### `data` block (semua value adalah string)

| Field | Type | Keterangan |
|---|---|---|
| `module` | string | Modul sumber notif: `announcement`, `news`, `community`, `event`, `qna`, `subscription`, `report`, `region` |
| `event_type` | string | Event spesifik dalam modul. Lihat daftar per modul di bawah |
| `entity_id` | string | UUID entitas terkait (article_id, community_id, event_id, dll). `""` jika tidak ada |
| `click_action` | string | Deep link untuk navigasi Flutter. Format: `kai://...` |
| `bypass` | string | `"true"` jika notif ini bypass preferences. `"false"` jika tidak |
| `extra` | string | JSON-encoded string untuk data tambahan yang tidak cukup di field standar. `""` jika tidak ada |

### Android priority

| Kondisi | `android.priority` | `apns-priority` |
|---|---|---|
| Bypass / CRITICAL / HIGH | `high` | `10` |
| Normal / MEDIUM | `normal` | `5` |
| Low / reminder | `normal` | `5` |

---

## Payload Per Modul

---

### 1. Announcement

#### `announcement.info_published` — Announcement info dipublish

```json
{
  "notification": {
    "title": "Pengumuman",
    "body": "Pendaftaran member baru dibuka mulai 1 Juli 2026"
  },
  "data": {
    "module": "announcement",
    "event_type": "info_published",
    "entity_id": "ann_uuid",
    "click_action": "kai://announcements/ann_uuid",
    "bypass": "false",
    "extra": ""
  }
}
```

#### `announcement.critical_published` — Announcement darurat (bypass)

```json
{
  "notification": {
    "title": "⚠️ Peringatan Darurat",
    "body": "Terdapat gangguan layanan di area Jakarta Pusat"
  },
  "data": {
    "module": "announcement",
    "event_type": "critical_published",
    "entity_id": "ann_uuid",
    "click_action": "kai://announcements/ann_uuid",
    "bypass": "true",
    "extra": "{\"priority\":\"CRITICAL\",\"type\":\"disaster\"}"
  }
}
```

---

### 2. News

#### `news.article_published` — Artikel baru published

```json
{
  "notification": {
    "title": "Berita Terbaru",
    "body": "Jadwal Perjalanan KAI Semester II 2026 Resmi Dirilis",
    "image": "https://cdn.kai.id/news/thumb_uuid.jpg"
  },
  "data": {
    "module": "news",
    "event_type": "article_published",
    "entity_id": "article_uuid",
    "click_action": "kai://news/article_uuid",
    "bypass": "false",
    "extra": ""
  }
}
```

#### `news.article_approved` — Artikel Pro Member di-approve (bypass)

```json
{
  "notification": {
    "title": "Artikel Disetujui",
    "body": "Artikel \"Tips Perjalanan Hemat\" kamu telah dipublikasikan"
  },
  "data": {
    "module": "news",
    "event_type": "article_approved",
    "entity_id": "article_uuid",
    "click_action": "kai://news/article_uuid",
    "bypass": "true",
    "extra": ""
  }
}
```

#### `news.article_rejected` — Artikel Pro Member di-reject (bypass)

```json
{
  "notification": {
    "title": "Artikel Tidak Disetujui",
    "body": "Artikel \"Tips Perjalanan Hemat\" memerlukan revisi"
  },
  "data": {
    "module": "news",
    "event_type": "article_rejected",
    "entity_id": "article_uuid",
    "click_action": "kai://news/my-articles/article_uuid",
    "bypass": "true",
    "extra": "{\"reason\":\"Konten perlu dilengkapi dengan sumber yang valid\"}"
  }
}
```

---

### 3. Community

#### `community.post_new` — Post baru di komunitas

```json
{
  "notification": {
    "title": "Post Baru di Komunitas Pecinta Kereta",
    "body": "Budi Santoso: \"Foto perjalanan Argo Bromo kemarin keren banget!\""
  },
  "data": {
    "module": "community",
    "event_type": "post_new",
    "entity_id": "post_uuid",
    "click_action": "kai://communities/comm_uuid/posts/post_uuid",
    "bypass": "false",
    "extra": "{\"community_id\":\"comm_uuid\",\"community_name\":\"Pecinta Kereta\"}"
  }
}
```

#### `community.join_request_approved` — Join request di-approve

```json
{
  "notification": {
    "title": "Bergabung Berhasil",
    "body": "Kamu telah diterima di komunitas Pecinta Kereta"
  },
  "data": {
    "module": "community",
    "event_type": "join_request_approved",
    "entity_id": "comm_uuid",
    "click_action": "kai://communities/comm_uuid",
    "bypass": "false",
    "extra": "{\"community_name\":\"Pecinta Kereta\"}"
  }
}
```

#### `community.join_request_rejected` — Join request di-reject (bypass)

```json
{
  "notification": {
    "title": "Permintaan Bergabung Ditolak",
    "body": "Permintaan bergabung ke komunitas Pecinta Kereta tidak disetujui"
  },
  "data": {
    "module": "community",
    "event_type": "join_request_rejected",
    "entity_id": "comm_uuid",
    "click_action": "kai://communities",
    "bypass": "true",
    "extra": "{\"community_name\":\"Pecinta Kereta\"}"
  }
}
```

#### `community.member_kicked` — Kena kick (bypass)

```json
{
  "notification": {
    "title": "Kamu Dikeluarkan dari Komunitas",
    "body": "Kamu telah dikeluarkan dari komunitas Pecinta Kereta"
  },
  "data": {
    "module": "community",
    "event_type": "member_kicked",
    "entity_id": "comm_uuid",
    "click_action": "kai://communities",
    "bypass": "true",
    "extra": "{\"community_name\":\"Pecinta Kereta\"}"
  }
}
```

#### `community.member_banned` — Kena ban (bypass)

```json
{
  "notification": {
    "title": "Akses Komunitas Dibatasi",
    "body": "Kamu telah di-ban dari komunitas Pecinta Kereta"
  },
  "data": {
    "module": "community",
    "event_type": "member_banned",
    "entity_id": "comm_uuid",
    "click_action": "kai://communities",
    "bypass": "true",
    "extra": "{\"community_name\":\"Pecinta Kereta\",\"reason\":\"Melanggar aturan komunitas\"}"
  }
}
```

#### `community.member_joined` — Member lain bergabung

```json
{
  "notification": {
    "title": "Member Baru",
    "body": "Andi Pratama bergabung ke komunitas Pecinta Kereta"
  },
  "data": {
    "module": "community",
    "event_type": "member_joined",
    "entity_id": "comm_uuid",
    "click_action": "kai://communities/comm_uuid/members",
    "bypass": "false",
    "extra": "{\"community_id\":\"comm_uuid\",\"community_name\":\"Pecinta Kereta\"}"
  }
}
```

#### `community.member_left` — Member lain keluar

```json
{
  "notification": {
    "title": "Member Keluar",
    "body": "Andi Pratama meninggalkan komunitas Pecinta Kereta"
  },
  "data": {
    "module": "community",
    "event_type": "member_left",
    "entity_id": "comm_uuid",
    "click_action": "kai://communities/comm_uuid/members",
    "bypass": "false",
    "extra": "{\"community_id\":\"comm_uuid\",\"community_name\":\"Pecinta Kereta\"}"
  }
}
```

---

### 4. Event

#### `event.new_published` — Event baru dipublish

```json
{
  "notification": {
    "title": "Event Baru",
    "body": "Korean Cultural Festival — 15 Juni 2026 di Jakarta",
    "image": "https://cdn.kai.id/events/thumb_uuid.jpg"
  },
  "data": {
    "module": "event",
    "event_type": "new_published",
    "entity_id": "event_uuid",
    "click_action": "kai://events/event_uuid",
    "bypass": "false",
    "extra": "{\"category\":\"culture\",\"start_time\":\"2026-06-15T09:00:00Z\"}"
  }
}
```

#### `event.reminder` — Reminder sebelum event

```json
{
  "notification": {
    "title": "Pengingat Event",
    "body": "Korean Cultural Festival dimulai 24 jam lagi"
  },
  "data": {
    "module": "event",
    "event_type": "reminder",
    "entity_id": "event_uuid",
    "click_action": "kai://events/event_uuid",
    "bypass": "false",
    "extra": "{\"hours_before\":24,\"start_time\":\"2026-06-15T09:00:00Z\"}"
  }
}
```

---

### 5. Q&A

#### `qna.question_answered` — Pertanyaan dijawab

```json
{
  "notification": {
    "title": "Pertanyaan Kamu Dijawab",
    "body": "Admin telah menjawab pertanyaanmu tentang cara upgrade ke Pro"
  },
  "data": {
    "module": "qna",
    "event_type": "question_answered",
    "entity_id": "question_uuid",
    "click_action": "kai://qna/questions/question_uuid",
    "bypass": "false",
    "extra": ""
  }
}
```

#### `qna.question_rejected` — Pertanyaan ditolak

```json
{
  "notification": {
    "title": "Pertanyaan Tidak Dapat Diproses",
    "body": "Pertanyaanmu tentang cara upgrade ke Pro tidak dapat dijawab saat ini"
  },
  "data": {
    "module": "qna",
    "event_type": "question_rejected",
    "entity_id": "question_uuid",
    "click_action": "kai://qna/questions/question_uuid",
    "bypass": "false",
    "extra": "{\"reason\":\"Pertanyaan sudah dijawab di FAQ\"}"
  }
}
```

#### `qna.question_converted_faq` — Pertanyaan dikonversi ke FAQ

```json
{
  "notification": {
    "title": "Pertanyaanmu Jadi FAQ!",
    "body": "Pertanyaanmu tentang cara upgrade ke Pro telah ditambahkan ke FAQ"
  },
  "data": {
    "module": "qna",
    "event_type": "question_converted_faq",
    "entity_id": "question_uuid",
    "click_action": "kai://qna/faq/faq_uuid",
    "bypass": "false",
    "extra": "{\"faq_id\":\"faq_uuid\"}"
  }
}
```

---

### 6. Subscription

#### `subscription.upgrade_submitted` — Request upgrade dikonfirmasi (bypass)

```json
{
  "notification": {
    "title": "Permintaan Upgrade Diterima",
    "body": "Permintaan upgrade ke Pro sedang diproses, estimasi 1x24 jam"
  },
  "data": {
    "module": "subscription",
    "event_type": "upgrade_submitted",
    "entity_id": "subscription_uuid",
    "click_action": "kai://subscription",
    "bypass": "true",
    "extra": ""
  }
}
```

#### `subscription.upgrade_approved` — Upgrade di-approve (bypass)

```json
{
  "notification": {
    "title": "Selamat! Kamu Sekarang Pro Member 🎉",
    "body": "Akun kamu telah berhasil diupgrade ke paket Pro. Nikmati semua fitur eksklusif!"
  },
  "data": {
    "module": "subscription",
    "event_type": "upgrade_approved",
    "entity_id": "subscription_uuid",
    "click_action": "kai://subscription",
    "bypass": "true",
    "extra": "{\"plan\":\"pro\",\"expires_at\":\"2027-06-08T00:00:00Z\"}"
  }
}
```

#### `subscription.upgrade_rejected` — Upgrade di-reject (bypass)

```json
{
  "notification": {
    "title": "Upgrade Tidak Disetujui",
    "body": "Permintaan upgrade ke Pro tidak dapat diproses saat ini"
  },
  "data": {
    "module": "subscription",
    "event_type": "upgrade_rejected",
    "entity_id": "subscription_uuid",
    "click_action": "kai://subscription",
    "bypass": "true",
    "extra": "{\"reason\":\"Dokumen verifikasi tidak lengkap\"}"
  }
}
```

#### `subscription.expired` — Plan expired (bypass)

```json
{
  "notification": {
    "title": "Paket Pro Kamu Telah Berakhir",
    "body": "Akun kamu telah kembali ke paket Standard. Perpanjang sekarang untuk tetap menikmati fitur Pro."
  },
  "data": {
    "module": "subscription",
    "event_type": "expired",
    "entity_id": "subscription_uuid",
    "click_action": "kai://subscription",
    "bypass": "true",
    "extra": ""
  }
}
```

#### `subscription.expiry_reminder` — Expiry reminder

```json
{
  "notification": {
    "title": "Paket Pro Segera Berakhir",
    "body": "Paket Pro kamu akan berakhir dalam 7 hari. Segera perpanjang agar tidak terputus."
  },
  "data": {
    "module": "subscription",
    "event_type": "expiry_reminder",
    "entity_id": "subscription_uuid",
    "click_action": "kai://subscription",
    "bypass": "false",
    "extra": "{\"days_remaining\":7,\"expires_at\":\"2026-06-15T00:00:00Z\"}"
  }
}
```

---

### 7. Region

#### `region.join_approved` — Member disetujui masuk region (bypass)

```json
{
  "notification": {
    "title": "Kamu Bergabung ke Region Baru",
    "body": "Selamat datang di Region Jawa Tengah!"
  },
  "data": {
    "module": "region",
    "event_type": "join_approved",
    "entity_id": "region_uuid",
    "click_action": "kai://regions/region_uuid",
    "bypass": "true",
    "extra": "{\"region_name\":\"Jawa Tengah\"}"
  }
}
```

---

## Flutter Handling Guide

### Setup: Pisahkan foreground vs background

```dart
void setupNotificationHandlers() {

  // 1. App TERMINATED — user tap notif → app buka
  FirebaseMessaging.instance.getInitialMessage().then((message) {
    if (message != null) _handleNotificationTap(message.data);
  });

  // 2. App BACKGROUND — user tap notif
  FirebaseMessaging.onMessageOpenedApp.listen((message) {
    _handleNotificationTap(message.data);
  });

  // 3. App FOREGROUND — notif masuk, render sendiri
  FirebaseMessaging.onMessage.listen((message) {
    _handleForegroundNotification(message);
  });
}
```

### Handler: Navigasi dari `click_action`

```dart
void _handleNotificationTap(Map<String, dynamic> data) {
  final clickAction = data['click_action'] as String? ?? '';
  // Arahkan ke screen sesuai deep link
  NavigationService.navigateTo(clickAction);
}
```

### Handler: Foreground — render in-app

```dart
void _handleForegroundNotification(RemoteMessage message) {
  final module = message.data['module'] as String? ?? '';
  final eventType = message.data['event_type'] as String? ?? '';

  // Bypass → tampilkan sebagai banner/snackbar yang lebih prominent
  final isBypass = message.data['bypass'] == 'true';

  // Render in-app notification UI
  NotificationOverlay.show(
    title: message.notification?.title ?? '',
    body: message.notification?.body ?? '',
    prominent: isBypass,
    onTap: () => _handleNotificationTap(message.data),
  );

  // Update badge / notification center
  NotificationCenter.add(module: module, eventType: eventType, data: message.data);
}
```

### Parse `extra` field

```dart
Map<String, dynamic> parseExtra(Map<String, dynamic> data) {
  final extra = data['extra'] as String? ?? '';
  if (extra.isEmpty) return {};
  try {
    return jsonDecode(extra) as Map<String, dynamic>;
  } catch (_) {
    return {};
  }
}

// Contoh pakai:
final extra = parseExtra(message.data);
final reason = extra['reason'] as String? ?? '';
final daysRemaining = extra['days_remaining'] as int? ?? 0;
```

### Enum referensi: `event_type` per modul

```dart
// Untuk type-safe handling di Flutter
class NotificationEventType {
  // Announcement
  static const announcementInfoPublished   = 'info_published';
  static const announcementCritical        = 'critical_published';

  // News
  static const newsArticlePublished        = 'article_published';
  static const newsArticleApproved         = 'article_approved';
  static const newsArticleRejected         = 'article_rejected';

  // Community
  static const communityPostNew            = 'post_new';
  static const communityJoinApproved       = 'join_request_approved';
  static const communityJoinRejected       = 'join_request_rejected';
  static const communityMemberKicked       = 'member_kicked';
  static const communityMemberBanned       = 'member_banned';
  static const communityMemberJoined       = 'member_joined';
  static const communityMemberLeft         = 'member_left';

  // Event
  static const eventNewPublished           = 'new_published';
  static const eventReminder               = 'reminder';

  // Q&A
  static const qnaQuestionAnswered         = 'question_answered';
  static const qnaQuestionRejected         = 'question_rejected';
  static const qnaQuestionConvertedFaq     = 'question_converted_faq';

  // Subscription
  static const subscriptionSubmitted       = 'upgrade_submitted';
  static const subscriptionApproved        = 'upgrade_approved';
  static const subscriptionRejected        = 'upgrade_rejected';
  static const subscriptionExpired         = 'expired';
  static const subscriptionExpiryReminder  = 'expiry_reminder';

  // Region
  static const regionJoinApproved          = 'join_approved';
}
```

---

*Referensi terkait: `NOTIFICATION_RULES_ENGINE.md` (kapan payload ini dikirim), `API_SPEC_FCM.md` (token management), `API_SPEC_NOTIFICATION_PREFERENCES.md` (user control)*
