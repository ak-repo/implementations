# AWS S3 File Upload System with Go

A production-grade file upload system using **AWS S3 Pre-signed URLs** with Go (Gin framework). This architecture offloads file uploads directly to S3, eliminating backend bandwidth bottlenecks and enabling scalable, cost-efficient file handling.

---

## Table of Contents

- [System Architecture](#system-architecture)
- [Mental Model](#mental-model)
- [Upload Workflow](#upload-workflow)
- [Project Structure](#project-structure)
- [API Documentation](#api-documentation)
- [Frontend Demo](#frontend-demo)
- [Environment Variables](#environment-variables)
- [How to Run](#how-to-run)
- [Security Best Practices](#security-best-practices)
- [Production Considerations](#production-considerations)
- [Common Mistakes](#common-mistakes)

---

## System Architecture

The system consists of three main components working together to enable secure, direct-to-S3 file uploads:

```
┌─────────────────┐      Request Upload URL       ┌─────────────────┐
│                 │ ─────────────────────────────>│                 │
│     Client      │                               │   Backend (Go)  │
│   (Browser/JS)  │<───────────────────────────── │                 │
│                 │      Pre-signed URL + Key       │  - Generate URL │
│                 │                               │  - IAM Auth     │
└────────┬────────┘                               └─────────────────┘
         │
         │  2. Upload File Directly to S3
         │
         ▼
┌─────────────────┐
│                 │
│    AWS S3       │
│                 │
│  - Object Store │
│  - Pre-signed   │
│    URL Auth     │
└─────────────────┘
```

### Components

| Component | Responsibility |
|-----------|--------------|
| **Client** | Requests upload permission, uploads file directly to S3 using pre-signed URL |
| **Backend (Go)** | Authenticates client, generates time-limited pre-signed URLs, tracks metadata |
| **AWS S3** | Stores files, validates pre-signed signatures, handles actual upload traffic |

---

## Mental Model

Understanding these core concepts is essential before implementing the system.

### What is S3?

**Amazon Simple Storage Service (S3)** is an object storage service. Unlike filesystem storage, S3 stores data as **objects** (files + metadata) in flat **buckets** (containers). There is no directory hierarchy—paths are just prefixes in object keys.

### What is a Bucket?

A **bucket** is a top-level container in S3 with a globally unique name. Think of it as the "database" for your files. All objects live within a bucket.

```
Bucket: my-app-uploads
├── uploads/abc123-image.png
├── uploads/def456-document.pdf
└── uploads/ghi789-video.mp4
```

### What is an Object Key?

An **object key** is the unique identifier for an object within a bucket. It can include path-like prefixes (e.g., `uploads/2024/03/file.png`) but S3 treats it as a single string.

### What is a Pre-signed URL?

A **pre-signed URL** is a temporary, signed URL that grants time-limited access to perform a specific S3 operation (like `PUT` or `GET`). The backend generates this URL using AWS credentials, but the client uses it directly—**no AWS credentials are exposed to the client**.

```
https://my-bucket.s3.amazonaws.com/uploads/abc123.png
?X-Amz-Algorithm=AWS4-HMAC-SHA256
&X-Amz-Credential=AKIA.../20240315/us-east-1/s3/aws4_request
&X-Amz-Date=20240315T120000Z
&X-Amz-Expires=300
&X-Amz-SignedHeaders=host
&X-Amz-Signature=...
```

### Why Backend Should NOT Upload Files

| Approach | Problem |
|----------|---------|
| **Backend Upload** | File streams through your server, consuming bandwidth, memory, and CPU. Becomes a bottleneck at scale. |
| **Pre-signed URL** | Backend only generates a URL (bytes). Client uploads directly to S3. Backend resources are unaffected by file size or upload volume. |

**Production rule:** Never proxy file uploads through your backend. Always use pre-signed URLs.

---

## Upload Workflow

### Sequence Diagram

```
Client                  Backend                 AWS S3
  |                       |                       |
  |  1. POST /generate    |                       |
  |     {filename, type} |                       |
  |──────────────────────>|                       |
  |                       |                       |
  |                       |  2. Generate          |
  |                       |     Pre-signed URL    |
  |                       |  (AWS SDK)            |
  |                       |< - - - - - - - - - - -|
  |                       |                       |
  |  3. Return            |                       |
  |     {upload_url, key} |                       |
  |<──────────────────────|                       |
  |                       |                       |
  |  4. PUT upload_url    |                       |
  |     (binary body)     |                       |
  |───────────────────────────────────────────────>|
  |                       |                       |
  |  5. HTTP 200 OK       |                       |
  |<───────────────────────────────────────────────|
  |                       |                       |
  |  6. POST /complete    |                       |
  |     {file_key}        |                       |
  |──────────────────────>|                       |
  |                       |  7. Store metadata    |
  |                       |     in database       |
  |  8. Success           |                       |
  |<──────────────────────|                       |
```

### Step-by-Step Explanation

1. **Client requests upload URL** — Sends filename and content type to backend
2. **Backend generates pre-signed URL** — Uses AWS SDK with IAM credentials to create a time-limited `PUT` URL
3. **Backend returns URL + key** — Client receives the direct S3 URL and the object key for later reference
4. **Client uploads directly to S3** — Uses HTTP `PUT` with the pre-signed URL (no authentication headers needed)
5. **S3 validates and stores** — Validates signature, stores file, returns success
6. **Client notifies backend** — Optional: informs backend upload is complete
7. **Backend stores metadata** — Saves file key, size, user ID, timestamp to database

---

## Project Structure

```
s3-upload-system/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── config/
│   │   └── config.go            # Environment configuration
│   ├── handler/
│   │   └── upload_handler.go    # HTTP handlers (Gin)
│   ├── service/
│   │   └── upload_service.go    # Business logic
│   ├── s3/
│   │   └── s3_client.go         # AWS S3 client wrapper
│   └── model/
│       └── upload.go            # Request/response structs
├── frontend/
│   └── index.html               # Demo upload page
├── .env.example                 # Environment template
├── go.mod                       # Go module definition
└── README.md                    # This file
```

### Directory Purposes

| Path | Purpose |
|------|---------|
| `cmd/server/` | Application bootstrap and HTTP server setup |
| `internal/handler/` | HTTP request/response handling (Gin routes) |
| `internal/service/` | Core business logic (validation, orchestration) |
| `internal/s3/` | AWS S3 SDK integration, pre-signed URL generation |
| `internal/config/` | Environment variable loading and validation |
| `frontend/` | Static HTML/JS demo for testing uploads |

---

## API Documentation

### POST /api/v1/upload/generate-url

Generate a pre-signed URL for direct S3 upload.

**Request:**

```http
POST /api/v1/upload/generate-url
Content-Type: application/json

{
  "filename": "profile-picture.png",
  "content_type": "image/png"
}
```

**Response:**

```http
HTTP/1.1 200 OK
Content-Type: application/json

{
  "upload_url": "https://my-bucket.s3.amazonaws.com/...",
  "file_key": "uploads/550e8400-e29b-41d4-a716-446655440000-profile-picture.png",
  "expires_in": 300
}
```

| Field | Type | Description |
|-------|------|-------------|
| `upload_url` | string | Pre-signed URL for HTTP PUT to S3 |
| `file_key` | string | Unique S3 object key (store this in your database) |
| `expires_in` | int | URL validity in seconds (default: 300) |

**Error Responses:**

```http
HTTP/1.1 400 Bad Request
{"error": "invalid content type"}

HTTP/1.1 500 Internal Server Error
{"error": "failed to generate upload URL"}
```

### POST /api/v1/upload/complete (Optional)

Notify backend that upload is complete. Useful for triggering post-processing.

**Request:**

```http
POST /api/v1/upload/complete
Content-Type: application/json

{
  "file_key": "uploads/550e8400-e29b-41d4-a716-446655440000-profile-picture.png",
  "metadata": {
    "user_id": "user_123",
    "original_filename": "profile-picture.png"
  }
}
```

**Response:**

```http
HTTP/1.1 200 OK
{"status": "upload recorded"}
```

---

## Frontend Demo

A minimal HTML page demonstrates the upload flow:

```html
<!-- frontend/index.html -->
<!DOCTYPE html>
<html>
<head>
  <title>S3 Upload Demo</title>
</head>
<body>
  <input type="file" id="fileInput" />
  <button onclick="upload()">Upload</button>

  <script>
    async function upload() {
      const file = document.getElementById('fileInput').files[0];

      // 1. Get pre-signed URL from backend
      const res = await fetch('http://localhost:8080/api/v1/upload/generate-url', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          filename: file.name,
          content_type: file.type
        })
      });
      const { upload_url, file_key } = await res.json();

      // 2. Upload directly to S3
      await fetch(upload_url, {
        method: 'PUT',
        body: file,
        headers: { 'Content-Type': file.type }
      });

      // 3. Notify backend (optional)
      await fetch('http://localhost:8080/api/v1/upload/complete', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ file_key })
      });

      alert('Upload complete!');
    }
  </script>
</body>
</html>
```

**Key points:**
- No AWS credentials in frontend code
- Direct browser-to-S3 upload (no backend proxy)
- Simple `fetch` API usage

---

## Environment Variables

Create a `.env` file from `.env.example`:

```bash
# AWS Configuration
AWS_ACCESS_KEY_ID=AKIAxxxxxxxxxxxx
AWS_SECRET_ACCESS_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
AWS_REGION=us-east-1
S3_BUCKET_NAME=my-app-uploads

# Server Configuration
PORT=8080
ENV=development

# Upload Settings
UPLOAD_URL_EXPIRY=300  # seconds
MAX_FILE_SIZE=10485760 # 10MB in bytes
```

### Required Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `AWS_ACCESS_KEY_ID` | IAM user access key | `AKIAIOSFODNN7EXAMPLE` |
| `AWS_SECRET_ACCESS_KEY` | IAM user secret key | `wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY` |
| `AWS_REGION` | S3 bucket region | `us-east-1` |
| `S3_BUCKET_NAME` | Target S3 bucket | `my-app-uploads` |

### IAM Permissions Required

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject",
        "s3:GetObject"
      ],
      "Resource": "arn:aws:s3:::my-app-uploads/*"
    }
  ]
}
```

---

## How to Run

### Prerequisites

- Go 1.21+
- AWS account with S3 bucket
- IAM user with S3 permissions

### Step-by-Step

1. **Clone the repository:**
   ```bash
   git clone https://github.com/yourusername/s3-upload-system.git
   cd s3-upload-system
   ```

2. **Install dependencies:**
   ```bash
   go mod download
   ```

3. **Configure environment:**
   ```bash
   cp .env.example .env
   # Edit .env with your AWS credentials
   ```

4. **Run the server:**
   ```bash
   go run cmd/server/main.go
   ```

5. **Open the demo:**
   ```bash
   open frontend/index.html
   # Or manually open in browser
   ```

6. **Test upload:**
   - Select a file
   - Click Upload
   - Verify file appears in S3 bucket

---

## Security Best Practices

### Why Pre-signed URLs Are Safe

| Security Aspect | Implementation |
|-----------------|----------------|
| **No credential exposure** | AWS credentials never leave the backend |
| **Time-limited** | URLs expire (default 5 minutes) |
| **Single-purpose** | Each URL is for one specific operation on one specific object |
| **Signature validation** | AWS validates the cryptographic signature on every request |

### IAM Permissions (Least Privilege)

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": [
        "s3:PutObject"
      ],
      "Resource": "arn:aws:s3:::my-app-uploads/uploads/*",
      "Condition": {
        "StringEquals": {
          "s3:x-amz-acl": "private"
        }
      }
    }
  ]
}
```

### File Validation Strategies

1. **Content-Type validation** — Whitelist allowed MIME types on backend
2. **File extension validation** — Check extension matches content type
3. **Size limits** — Enforce maximum file size before generating URL
4. **Virus scanning** — Use AWS Lambda + ClamAV for sensitive applications

### Never Do These

- Store AWS credentials in frontend code
- Use root AWS account credentials
- Make buckets public (`public-read` ACL)
- Trust client-provided content types without validation

---

## Production Considerations

### File Naming Strategy

Always generate unique filenames to prevent collisions and overwrites:

```go
// Pattern: {prefix}/{uuid}-{sanitized-filename}
key := fmt.Sprintf("uploads/%s-%s",
    uuid.New().String(),
    sanitizeFilename(originalFilename))

// Example: uploads/550e8400-e29b-41d4-a716-446655440000-profile.png
```

### Multipart Upload for Large Files

For files > 100MB, use S3 Multipart Upload:

1. Backend initiates multipart upload (`CreateMultipartUpload`)
2. Backend generates pre-signed URLs for each part
3. Client uploads parts in parallel
4. Client completes upload (`CompleteMultipartUpload`)

### CDN Integration (CloudFront)

For serving uploaded files:

```
User → CloudFront → S3 (for reads)
User → Backend → S3 (for writes via pre-signed URLs)
```

Configure CloudFront with Origin Access Control (OAC) for secure S3 access.

### Logging and Monitoring

```go
// Log upload events
logger.Info("upload_initiated",
    zap.String("file_key", key),
    zap.String("user_id", userID),
    zap.Int64("size", fileSize),
    zap.String("content_type", contentType))

// Metrics to track
- Upload success rate
- Average file size
- Upload duration
- S3 4xx/5xx errors
```

### Database Schema (Optional)

```sql
CREATE TABLE uploads (
    id UUID PRIMARY KEY,
    file_key VARCHAR(512) NOT NULL UNIQUE,
    original_filename VARCHAR(255),
    content_type VARCHAR(100),
    size_bytes BIGINT,
    user_id UUID REFERENCES users(id),
    status VARCHAR(50), -- pending, completed, failed
    created_at TIMESTAMP DEFAULT NOW(),
    completed_at TIMESTAMP
);

CREATE INDEX idx_uploads_user_id ON uploads(user_id);
CREATE INDEX idx_uploads_status ON uploads(status);
```

---

## Common Mistakes

### Uploading via Backend

```go
// WRONG: Streaming through backend
func wrongApproach(file io.Reader) {
    s3Client.PutObject(bucket, key, file) // Bandwidth bottleneck!
}

// CORRECT: Generate pre-signed URL
func correctApproach() string {
    url, _ := s3Client.PresignPutObject(bucket, key, expiry)
    return url // Client uploads directly
}
```

### Using Original Filenames

```go
// WRONG: Collision and security risk
key := filepath.Join("uploads", filename)

// CORRECT: UUID + sanitization
key := fmt.Sprintf("uploads/%s-%s", uuid.New(), sanitize(filename))
```

### Public Buckets

```go
// WRONG: Anyone can access
ACL: "public-read"

// CORRECT: Private bucket, pre-signed access only
ACL: "private"
```

### Ignoring Content-Type Validation

```go
// WRONG: Accept any content type
func generateURL(filename, contentType string) { ... }

// CORRECT: Whitelist validation
var allowedTypes = map[string]bool{
    "image/png": true,
    "image/jpeg": true,
    "application/pdf": true,
}
```

### Long-Lived URLs

```go
// WRONG: URL valid for 24 hours
cfg.Presign(expiry: 24 * time.Hour)

// CORRECT: Short expiry (5 minutes is standard)
cfg.Presign(expiry: 5 * time.Minute)
```

---

## License

MIT License - see LICENSE file for details.

---

## Resources

- [AWS S3 Pre-signed URLs Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/PresignedUrlUploadObject.html)
- [AWS SDK for Go v2](https://github.com/aws/aws-sdk-go-v2)
- [Gin Web Framework](https://gin-gonic.com/)
