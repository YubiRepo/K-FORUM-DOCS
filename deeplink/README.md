# Deep Link (App Links / Universal Links) — Server Setup

**Option B**: dedicated mobile subdomain `app.k-forum.yubicom.co.id`, served as static
files by Nginx (no Nuxt/Node involved). The Flutter app side is already wired
(Android manifest, iOS entitlements, routing, share links) — only these server
files remain.

| App | Value |
| --- | --- |
| Android package | `com.yubitech.k_forum` |
| iOS bundle | `com.yubitech.kForum` |
| Domain | `app.k-forum.yubicom.co.id` |
| Deep-linkable paths | `/news/*`, `/qna/*`, `/directory/*` |

## Files in this folder

| Repo path | Server destination |
| --- | --- |
| `well-known/assetlinks.json` | `/var/www/kforum-mobile-link/.well-known/assetlinks.json` |
| `well-known/apple-app-site-association` | `/var/www/kforum-mobile-link/.well-known/apple-app-site-association` (NO extension) |
| `nginx/kforum-mobile-link.conf` | `/etc/nginx/sites-available/kforum-mobile-link` → symlink to `sites-enabled/` |

## Before deploying — fill in 2 placeholders

1. **iOS `appID`** in `apple-app-site-association`: replace `YOUR_APPLE_TEAM_ID`
   with the 10-char Team ID (developer.apple.com → Membership Details) →
   e.g. `AB12CD34EF.com.yubitech.kForum`.
2. **Android fingerprint** in `assetlinks.json`: the value shipped is the **debug**
   keystore SHA-256, which is enough for App Links to work on debug/local test builds.
   This app will be distributed via the **Play Store**, so production uses **Play App
   Signing** (Google re-signs the AAB). The production SHA-256 only exists *after* the
   first upload — you cannot generate it locally.

   **After the first Play Store upload**, get it and add it to the array:
   - Play Console → your app → **Test and release → Setup → App signing**
     (a.k.a. *App Integrity*) → **App signing key certificate** → copy **SHA-256
     certificate fingerprint**.
   - Append it (keep the debug one too, so both internal-test and store builds verify):

   ```json
   [
     {
       "relation": ["delegate_permission/common.handle_all_urls"],
       "target": {
         "namespace": "android_app",
         "package_name": "com.yubitech.k_forum",
         "sha256_cert_fingerprints": [
           "16:E2:90:53:12:46:2A:26:B9:19:01:3F:66:E1:EF:CB:71:00:C9:2D:22:F2:DB:7C:81:56:6B:7A:D6:BB:87:0F",
           "PASTE_PLAY_APP_SIGNING_SHA256_HERE"
         ]
       }
     }
   ]
   ```
   Re-upload the file to `/.well-known/assetlinks.json` after editing — no app rebuild needed.

## Deploy steps (on the VPS)

```bash
sudo mkdir -p /var/www/kforum-mobile-link/.well-known
# copy the two verification files into .well-known/
sudo cp assetlinks.json                /var/www/kforum-mobile-link/.well-known/
sudo cp apple-app-site-association      /var/www/kforum-mobile-link/.well-known/

# install the nginx server block
sudo cp kforum-mobile-link.conf /etc/nginx/sites-available/kforum-mobile-link
sudo ln -s /etc/nginx/sites-available/kforum-mobile-link /etc/nginx/sites-enabled/
sudo nginx -t && sudo systemctl reload nginx
```

> If SSL is terminated at Cloudflare, the `listen 80` block is enough (Cloudflare
> handles 443). If Nginx terminates TLS itself, add the `listen 443 ssl` +
> certificate directives. The verification fetch by Android/iOS happens over HTTPS,
> so the domain MUST be reachable on https in production.

## Verify

```bash
# Files must return 200 + Content-Type: application/json
curl -I https://app.k-forum.yubicom.co.id/.well-known/assetlinks.json
curl -I https://app.k-forum.yubicom.co.id/.well-known/apple-app-site-association

# Android: after a fresh install, check verification + test a link
adb shell pm get-app-links com.yubitech.k_forum          # host should be "verified"
adb shell am start -a android.intent.action.VIEW \
  -d "https://app.k-forum.yubicom.co.id/news/SOME_ID"     # opens news detail in-app
```

Google's tester: `https://digitalassetlinks.googleapis.com/v1/statements:list?source.web.site=https://app.k-forum.yubicom.co.id&relation=delegate_permission/common.handle_all_urls`

## Notes

- `/qna/*` currently lands on the Q&A **list** (no Q&A detail route in the app yet).
- For Community & Event share links to open in-app, the **backend** `share_url` /
  `shareLink` must also use `https://app.k-forum.yubicom.co.id/...` with a matching path.
