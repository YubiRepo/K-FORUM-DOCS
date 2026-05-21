# Dokumentasi API Spec - FCM (Firebase Cloud Messaging) Push Notifications

Dokumentasi ini dibuat untuk kebutuhan tim Backend (Golang) dan Frontend (Flutter) untuk implementasi push notification menggunakan Firebase Cloud Messaging (FCM).

## Informasi Umum

- **Base URL Prefix**: `/api/v1` (Seluruh endpoint di bawah ini menggunakan prefix `/api/v1/mobile/fcm`)
- **Headers Global**:
  - `Content-Type: application/json`
  - `Accept: application/json`
  - `Authorization: Bearer <access_token>` (Required untuk semua endpoint FCM)

---

## Konsep FCM Token Management

### Prinsip Dasar
- **1 User = Multiple FCM Tokens** (satu user bisa login di banyak devices)
- **1 Device = 1 FCM Token** (unique per device)
- **Token bisa refresh/expired** (FCM auto-refresh token secara periodik)
- **Logout per device** (revoke token device yang logout saja, device lain tetap aktif)

### Flow Diagram
```
User Login Device A  ──→  Generate FCM Token A  ──→  POST /fcm/register
User Login Device B  ──→  Generate FCM Token B  ──→  POST /fcm/register
User Login Device C  ──→  Generate FCM Token C  ──→  POST /fcm/register

Token Refresh        ──→  PUT /fcm/update

User Logout Device A ──→  DELETE /fcm/revoke (Token A only)
                          ↓
                     Device B & C tetap aktif

Send Notification    ──→  Backend query active tokens
                          ↓
                     Send to Token B & Token C only
```

---

## Database Structure

### Table: `fcm_tokens`
Menyimpan FCM token untuk setiap device yang login.

```sql
CREATE TABLE fcm_tokens (
    id                  VARCHAR(36) PRIMARY KEY,              -- UUID
    user_id             VARCHAR(36) NOT NULL,                 -- FK to users.id
    fcm_token           TEXT NOT NULL,                        -- FCM token dari Firebase
    device_id           VARCHAR(255),                         -- Optional unique device identifier
    platform            VARCHAR(20) NOT NULL,                 -- 'android', 'ios', 'web'
    device_name         VARCHAR(255),                         -- e.g. "iPhone 13 Pro", "Samsung Galaxy S21"
    device_model        VARCHAR(255),                         -- e.g. "iPhone14,3", "SM-G991B"
    os_version          VARCHAR(50),                          -- e.g. "iOS 16.5", "Android 13"
    app_version         VARCHAR(50),                          -- e.g. "1.2.3"
    is_active           BOOLEAN NOT NULL DEFAULT TRUE,        -- Token status
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    last_used_at        TIMESTAMP,                            -- Last time notif dikirim ke token ini
    
    CONSTRAINT fk_user FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT unique_fcm_token UNIQUE (fcm_token)
);

-- Indexes untuk performance
CREATE INDEX idx_fcm_tokens_user_id ON fcm_tokens(user_id);
CREATE INDEX idx_fcm_tokens_is_active ON fcm_tokens(is_active);
CREATE INDEX idx_fcm_tokens_user_active ON fcm_tokens(user_id, is_active);
```

### Field Descriptions

| Field | Type | Description |
|-------|------|-------------|
| `id` | VARCHAR(36) | Primary key (UUID) |
| `user_id` | VARCHAR(36) | Foreign key ke users table |
| `fcm_token` | TEXT | FCM token dari Firebase (bisa panjang) |
| `device_id` | VARCHAR(255) | Optional. Unique identifier per device (Android: Android ID, iOS: identifierForVendor) |
| `platform` | VARCHAR(20) | Platform: `android`, `ios`, `web` |
| `device_name` | VARCHAR(255) | Human-readable device name untuk display di settings |
| `device_model` | VARCHAR(255) | Model identifier teknis |
| `os_version` | VARCHAR(50) | OS version untuk debugging |
| `app_version` | VARCHAR(50) | App version untuk debugging |
| `is_active` | BOOLEAN | Status token: `true` = aktif, `false` = revoked/invalid |
| `created_at` | TIMESTAMP | Waktu token pertama kali di-register |
| `updated_at` | TIMESTAMP | Waktu token terakhir diupdate |
| `last_used_at` | TIMESTAMP | Waktu terakhir notif dikirim ke token ini |

---

## Model Data Utama

### 1. FCM Token Object
```json
{
  "id": "string (UUID)",
  "user_id": "string",
  "fcm_token": "string",
  "device_id": "string (nullable)",
  "platform": "string (enum: 'android', 'ios', 'web')",
  "device_name": "string (nullable)",
  "device_model": "string (nullable)",
  "os_version": "string (nullable)",
  "app_version": "string (nullable)",
  "is_active": "boolean",
  "created_at": "string (ISO 8601)",
  "updated_at": "string (ISO 8601)",
  "last_used_at": "string (nullable, ISO 8601)"
}
```

### 2. Device Info (Simplified for List)
```json
{
  "id": "string",
  "device_name": "string",
  "platform": "string",
  "is_active": "boolean",
  "last_used_at": "string (nullable)",
  "created_at": "string"
}
```

### 3. Error Responses
```json
{
  "message": "Pesan error deskriptif"
}
```

---

## Daftar Endpoint

### 1. Register FCM Token
Register FCM token baru untuk device yang baru login atau install app.

- **URL**: `POST /api/v1/mobile/fcm/register`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "fcm_token": "eXJd8F3zRDm9k...", 
    "device_id": "android_abc123xyz",
    "platform": "android",
    "device_name": "Samsung Galaxy S21",
    "device_model": "SM-G991B",
    "os_version": "Android 13",
    "app_version": "1.2.3"
  }
  ```
  **Field Requirements:**
  - `fcm_token`: **Required** (string, FCM token dari Firebase)
  - `platform`: **Required** (enum: `android`, `ios`, `web`)
  - `device_id`: Optional (string, unique device identifier)
  - `device_name`: Optional (string, human-readable name)
  - `device_model`: Optional (string, technical model identifier)
  - `os_version`: Optional (string)
  - `app_version`: Optional (string)

- **Backend Logic**:
  ```go
  // Check if token already exists
  existingToken := db.FindByFCMToken(fcmToken)
  
  if existingToken != nil {
    // Token sudah ada, update saja
    if existingToken.UserID != currentUserID {
      // Token pindah ke user lain (re-install app dengan akun beda)
      // Update owner
      existingToken.UserID = currentUserID
    }
    existingToken.IsActive = true
    existingToken.UpdatedAt = time.Now()
    db.Update(existingToken)
  } else {
    // Token baru, insert
    newToken := FCMToken{
      ID:          uuid.New(),
      UserID:      currentUserID,
      FCMToken:    fcmToken,
      DeviceID:    deviceID,
      Platform:    platform,
      DeviceName:  deviceName,
      DeviceModel: deviceModel,
      OSVersion:   osVersion,
      AppVersion:  appVersion,
      IsActive:    true,
      CreatedAt:   time.Now(),
      UpdatedAt:   time.Now(),
    }
    db.Insert(newToken)
  }
  ```

- **Response (Success 200)**:
  ```json
  {
    "message": "FCM token berhasil didaftarkan",
    "data": {
      "id": "fcm_tok_001",
      "user_id": "usr_90210",
      "platform": "android",
      "device_name": "Samsung Galaxy S21",
      "is_active": true,
      "created_at": "2026-05-21T10:30:00.000Z"
    }
  }
  ```

- **Response (Error 400 - Invalid Platform)**:
  ```json
  {
    "message": "Platform tidak valid. Harus: android, ios, atau web"
  }
  ```

---

### 2. Update FCM Token
Update FCM token ketika token di-refresh oleh Firebase (auto-refresh).

- **URL**: `PUT /api/v1/mobile/fcm/update`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "old_token": "eXJd8F3zRDm9k...",
    "new_token": "fYKe9G4aSEn0l...",
    "device_id": "android_abc123xyz"
  }
  ```
  **Field Requirements:**
  - `old_token`: **Required** (string, token lama yang akan di-replace)
  - `new_token`: **Required** (string, token baru dari Firebase)
  - `device_id`: Optional (untuk identifikasi device jika old_token tidak ditemukan)

- **Backend Logic**:
  ```go
  // Find old token
  oldToken := db.FindByFCMToken(oldToken)
  
  if oldToken == nil {
    // Old token tidak ditemukan, treat as new registration
    // Call register logic
    return registerNewToken(newToken)
  }
  
  // Update token
  oldToken.FCMToken = newToken
  oldToken.UpdatedAt = time.Now()
  db.Update(oldToken)
  ```

- **Response (Success 200)**:
  ```json
  {
    "message": "FCM token berhasil diperbarui",
    "data": {
      "id": "fcm_tok_001",
      "updated_at": "2026-05-21T11:00:00.000Z"
    }
  }
  ```

- **Response (Error 404 - Old Token Not Found)**:
  ```json
  {
    "message": "Token lama tidak ditemukan. Silakan register token baru."
  }
  ```

---

### 3. Revoke FCM Token (Logout)
Revoke FCM token ketika user logout dari device tertentu.

- **URL**: `DELETE /api/v1/mobile/fcm/revoke`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Request Body**:
  ```json
  {
    "fcm_token": "eXJd8F3zRDm9k..."
  }
  ```
  **Field Requirements:**
  - `fcm_token`: **Required** (string, token yang akan di-revoke)

- **Backend Logic**:
  ```go
  // Find token
  token := db.FindByFCMToken(fcmToken)
  
  if token == nil || token.UserID != currentUserID {
    return NotFoundError("Token tidak ditemukan")
  }
  
  // Set inactive (soft delete)
  token.IsActive = false
  token.UpdatedAt = time.Now()
  db.Update(token)
  ```

- **Response (Success 200)**:
  ```json
  {
    "message": "FCM token berhasil direvoke"
  }
  ```

- **Response (Error 404 - Token Not Found)**:
  ```json
  {
    "message": "Token tidak ditemukan"
  }
  ```

> [!NOTE]
> **Catatan**: Endpoint ini hanya set `is_active = false` (soft delete), tidak menghapus record dari database. Ini penting untuk audit trail dan analytics.

---

### 4. Get My Devices
Mengambil daftar semua devices yang terdaftar untuk user (untuk halaman settings/manage devices).

- **URL**: `GET /api/v1/mobile/fcm/devices`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Query Parameters**:
  ```
  include_inactive : boolean (default: false, include revoked tokens)
  ```

- **Backend Logic**:
  ```go
  query := `
    SELECT id, device_name, platform, device_model, is_active, 
           last_used_at, created_at 
    FROM fcm_tokens 
    WHERE user_id = ?
  `
  
  if !includeInactive {
    query += " AND is_active = true"
  }
  
  query += " ORDER BY last_used_at DESC NULLS LAST, created_at DESC"
  
  devices := db.Query(query, currentUserID)
  ```

- **Response (Success 200)**:
  ```json
  {
    "data": [
      {
        "id": "fcm_tok_001",
        "device_name": "Samsung Galaxy S21",
        "platform": "android",
        "device_model": "SM-G991B",
        "is_active": true,
        "last_used_at": "2026-05-21T08:45:00.000Z",
        "created_at": "2026-04-01T10:30:00.000Z"
      },
      {
        "id": "fcm_tok_002",
        "device_name": "iPhone 13 Pro",
        "platform": "ios",
        "device_model": "iPhone14,3",
        "is_active": true,
        "last_used_at": "2026-05-20T22:15:00.000Z",
        "created_at": "2026-03-15T14:20:00.000Z"
      },
      {
        "id": "fcm_tok_003",
        "device_name": "iPad Air",
        "platform": "ios",
        "device_model": "iPad13,1",
        "is_active": false,
        "last_used_at": "2026-04-10T16:30:00.000Z",
        "created_at": "2026-02-20T09:00:00.000Z"
      }
    ]
  }
  ```

---

### 5. Remove Device
Menghapus device tertentu dari daftar (revoke token by device ID).

- **URL**: `DELETE /api/v1/mobile/fcm/devices/:id`
- **Autentikasi**: Ya (`Bearer <access_token>`)
- **Path Parameters**:
  - `id`: FCM Token ID (e.g. `fcm_tok_001`)

- **Backend Logic**:
  ```go
  token := db.FindByID(tokenID)
  
  if token == nil || token.UserID != currentUserID {
    return NotFoundError("Device tidak ditemukan")
  }
  
  // Set inactive
  token.IsActive = false
  token.UpdatedAt = time.Now()
  db.Update(token)
  ```

- **Response (Success 200)**:
  ```json
  {
    "message": "Device berhasil dihapus"
  }
  ```

- **Response (Error 404 - Device Not Found)**:
  ```json
  {
    "message": "Device tidak ditemukan"
  }
  ```

---

## Backend: Sending Push Notifications

### Flow Mengirim Notifikasi

```go
// 1. Get all active tokens for target users
func GetActiveTokens(userIDs []string) []string {
    query := `
        SELECT fcm_token 
        FROM fcm_tokens 
        WHERE user_id IN (?) 
          AND is_active = true
        ORDER BY last_used_at DESC
    `
    tokens := db.Query(query, userIDs)
    return tokens
}

// 2. Send notification via FCM
func SendNotification(userIDs []string, notification Notification) {
    tokens := GetActiveTokens(userIDs)
    
    payload := &messaging.MulticastMessage{
        Tokens: tokens,
        Notification: &messaging.Notification{
            Title: notification.Title,
            Body:  notification.Body,
            ImageURL: notification.ImageURL,
        },
        Data: map[string]string{
            "type":       notification.Type,
            "entity_id":  notification.EntityID,
            "click_action": notification.ClickAction,
        },
        Android: &messaging.AndroidConfig{
            Priority: "high",
            Notification: &messaging.AndroidNotification{
                Sound: "default",
                ChannelID: "kai_app_notifications",
            },
        },
        APNS: &messaging.APNSConfig{
            Payload: &messaging.APNSPayload{
                Aps: &messaging.Aps{
                    Sound: "default",
                    Badge: notification.Badge,
                },
            },
        },
    }
    
    // Send to FCM
    response, err := fcmClient.SendMulticast(ctx, payload)
    
    // Handle response & cleanup invalid tokens
    handleFCMResponse(response, tokens)
}

// 3. Handle FCM response & cleanup
func handleFCMResponse(response *messaging.BatchResponse, tokens []string) {
    for i, sendResponse := range response.Responses {
        token := tokens[i]
        
        if sendResponse.Success {
            // Update last_used_at
            db.Exec(`
                UPDATE fcm_tokens 
                SET last_used_at = NOW() 
                WHERE fcm_token = ?
            `, token)
        } else {
            // Check error type
            errorCode := sendResponse.Error.ErrorCode
            
            if errorCode == "registration-token-not-registered" ||
               errorCode == "invalid-registration-token" {
                // Token invalid, deactivate
                db.Exec(`
                    UPDATE fcm_tokens 
                    SET is_active = false, updated_at = NOW() 
                    WHERE fcm_token = ?
                `, token)
            }
        }
    }
}
```

### Notification Types & Payload

Backend bisa kirim berbagai tipe notifikasi dengan data tambahan:

```go
type Notification struct {
    Title       string            // Notification title
    Body        string            // Notification body
    ImageURL    string            // Optional image
    Type        string            // "event", "news", "community", "message", etc
    EntityID    string            // ID of related entity
    ClickAction string            // Deep link or screen to open
    Badge       int               // Badge count for iOS
    Data        map[string]string // Additional custom data
}

// Example: New Event Notification
notification := Notification{
    Title:       "Event Baru: Korean Cultural Festival",
    Body:        "Event akan diadakan pada 15 Juni 2026 di Jakarta",
    ImageURL:    "https://example.com/events/evt_001_thumb.jpg",
    Type:        "event",
    EntityID:    "evt_001",
    ClickAction: "kai://events/evt_001",
    Badge:       1,
}

// Example: News Notification
notification := Notification{
    Title:       "Berita Terbaru dari KAI Pusat",
    Body:        "Pendaftaran member baru dibuka hingga akhir bulan",
    Type:        "news",
    EntityID:    "news_123",
    ClickAction: "kai://news/news_123",
}
```

---

## Flutter Implementation Guide

### 1. Setup Firebase Messaging

```dart
// main.dart
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';

Future<void> main() async {
  WidgetsFlutterBinding.ensureInitialized();
  await Firebase.initializeApp();
  
  // Request permission (iOS)
  await FirebaseMessaging.instance.requestPermission(
    alert: true,
    badge: true,
    sound: true,
  );
  
  // Setup FCM
  await setupFCM();
  
  runApp(MyApp());
}

Future<void> setupFCM() async {
  // Get FCM token
  String? token = await FirebaseMessaging.instance.getToken();
  
  if (token != null) {
    // Register token ke backend
    await registerFCMToken(token);
  }
  
  // Listen to token refresh
  FirebaseMessaging.instance.onTokenRefresh.listen((newToken) {
    updateFCMToken(oldToken: token, newToken: newToken);
  });
}
```

### 2. Register Token ke Backend

```dart
Future<void> registerFCMToken(String fcmToken) async {
  // Get device info
  final deviceInfo = await getDeviceInfo();
  
  final response = await http.post(
    Uri.parse('$baseUrl/api/v1/mobile/fcm/register'),
    headers: {
      'Authorization': 'Bearer $accessToken',
      'Content-Type': 'application/json',
    },
    body: jsonEncode({
      'fcm_token': fcmToken,
      'device_id': deviceInfo.deviceId,
      'platform': deviceInfo.platform, // 'android' or 'ios'
      'device_name': deviceInfo.deviceName,
      'device_model': deviceInfo.model,
      'os_version': deviceInfo.osVersion,
      'app_version': deviceInfo.appVersion,
    }),
  );
  
  if (response.statusCode == 200) {
    print('FCM token registered successfully');
  }
}
```

### 3. Handle Token Refresh

```dart
Future<void> updateFCMToken({required String oldToken, required String newToken}) async {
  final response = await http.put(
    Uri.parse('$baseUrl/api/v1/mobile/fcm/update'),
    headers: {
      'Authorization': 'Bearer $accessToken',
      'Content-Type': 'application/json',
    },
    body: jsonEncode({
      'old_token': oldToken,
      'new_token': newToken,
    }),
  );
  
  if (response.statusCode == 200) {
    print('FCM token updated successfully');
  }
}
```

### 4. Revoke Token on Logout

```dart
Future<void> logout() async {
  // Get current FCM token
  String? token = await FirebaseMessaging.instance.getToken();
  
  if (token != null) {
    // Revoke token
    await http.delete(
      Uri.parse('$baseUrl/api/v1/mobile/fcm/revoke'),
      headers: {
        'Authorization': 'Bearer $accessToken',
        'Content-Type': 'application/json',
      },
      body: jsonEncode({
        'fcm_token': token,
      }),
    );
  }
  
  // Clear local auth & logout
  await clearAuthToken();
  navigateToLoginScreen();
}
```

### 5. Handle Incoming Notifications

```dart
// Setup notification handlers
void setupNotificationHandlers() {
  // Foreground notification
  FirebaseMessaging.onMessage.listen((RemoteMessage message) {
    print('Foreground notification: ${message.notification?.title}');
    
    // Show local notification atau in-app banner
    showInAppNotification(message);
  });
  
  // Background notification tap
  FirebaseMessaging.onMessageOpenedApp.listen((RemoteMessage message) {
    print('Notification opened: ${message.data}');
    
    // Navigate to relevant screen
    handleNotificationClick(message.data);
  });
  
  // App terminated, notification tap
  FirebaseMessaging.instance.getInitialMessage().then((message) {
    if (message != null) {
      handleNotificationClick(message.data);
    }
  });
}

void handleNotificationClick(Map<String, dynamic> data) {
  String type = data['type'];
  String entityId = data['entity_id'];
  
  switch (type) {
    case 'event':
      navigateToEventDetail(entityId);
      break;
    case 'news':
      navigateToNewsDetail(entityId);
      break;
    case 'community':
      navigateToCommunityDetail(entityId);
      break;
    default:
      navigateToHome();
  }
}
```

---

## Best Practices

### Backend

1. **Token Cleanup**: Jalankan cron job untuk cleanup inactive tokens yang sudah lama (e.g. > 90 hari):
   ```sql
   DELETE FROM fcm_tokens 
   WHERE is_active = false 
     AND updated_at < NOW() - INTERVAL '90 days';
   ```

2. **Batch Sending**: Untuk efisiensi, kirim notifikasi dalam batch max 500 tokens per request (FCM limit).

3. **Retry Logic**: Implement retry untuk FCM request yang gagal (network issue, timeout).

4. **Rate Limiting**: Implement rate limiting untuk prevent spam notification.

5. **Analytics**: Track notification metrics (sent, delivered, opened, conversion).

### Flutter

1. **Permission**: Selalu request notification permission di iOS sebelum register token.

2. **Token Refresh**: Always listen to `onTokenRefresh` dan update ke backend.

3. **Logout**: Always revoke token saat logout untuk prevent notifikasi ke device yang sudah logout.

4. **Deep Linking**: Implement proper deep link handling untuk notification click.

5. **Local Storage**: Save FCM token di local storage untuk avoid duplicate registration.

---

## Status Code Reference

| Code | Meaning |
|------|---------|
| `200` | Success - Request berhasil |
| `400` | Bad Request - Invalid input |
| `401` | Unauthorized - Token invalid/expired |
| `404` | Not Found - FCM token atau device tidak ditemukan |
| `409` | Conflict - FCM token sudah terdaftar (bisa diabaikan) |
| `422` | Unprocessable Entity - Validation error |
| `500` | Internal Server Error - Error di backend |

---

## Security Notes

1. **FCM Token adalah Sensitive**: Jangan expose FCM token di logs atau error messages.
2. **Authorization**: Semua endpoint FCM require valid access token.
3. **User Ownership**: Validate bahwa user hanya bisa revoke/manage token milik mereka sendiri.
4. **Token Rotation**: FCM token bisa rotate/refresh, backend harus handle dengan graceful.
5. **HTTPS Only**: Semua communication harus via HTTPS.

---

## Testing Checklist

### Backend Testing
- [ ] Register token baru
- [ ] Update token (refresh)
- [ ] Revoke token (logout)
- [ ] Get devices list
- [ ] Remove device
- [ ] Send notification ke single user
- [ ] Send notification ke multiple users
- [ ] Handle invalid token response dari FCM
- [ ] Cleanup inactive tokens

### Flutter Testing
- [ ] Request notification permission (iOS)
- [ ] Get FCM token on app start
- [ ] Register token ke backend
- [ ] Handle token refresh
- [ ] Revoke token on logout
- [ ] Receive foreground notification
- [ ] Receive background notification
- [ ] Handle notification click (app foreground)
- [ ] Handle notification click (app background)
- [ ] Handle notification click (app terminated)
- [ ] Deep link navigation
