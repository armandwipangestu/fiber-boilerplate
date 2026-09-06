# Phase 24 — Object Storage (S3 + local fallback)

> Scope: pluggable object storage for uploads (e.g. user avatars), mirroring the
> NestJS boilerplate `StorageProvider`: S3 when configured, local folder
> otherwise. Testable against MinIO locally.

## Deliverables

### 24.1 Storage package (`internal/storage`)
- `Storage` interface: `Upload(ctx, UploadOptions) (key, err)`, `Delete(ctx, key)`,
  `GetURL(key)`. Keys are `folder/<uuid><ext>`.
- `s3.go` — AWS SDK v2 client. A custom `S3_ENDPOINT` (MinIO) uses path-style
  addressing and `{endpoint}/{bucket}/{key}` URLs; blank endpoint targets AWS
  (`https://{bucket}.s3.{region}.amazonaws.com/{key}`). Bucket must pre-exist.
- `local.go` — fallback writing under `STORAGE_PATH` (default `storage/`), served
  from `/uploads`, URLs built from `PUBLIC_URL`.
- `New(cfg, logger)` picks S3 when access key/secret/bucket are all set, else
  local; logs the active provider.

### 24.2 Configuration
- `S3_REGION`, `S3_ACCESS_KEY_ID`, `S3_SECRET_ACCESS_KEY`, `S3_ENDPOINT`,
  `S3_BUCKET`, `S3_FORCE_PATH_STYLE`, `STORAGE_PATH`, `PUBLIC_URL`.

### 24.3 Example feature: avatar upload
- Migration `000008`: `avatar_url TEXT` on `users`.
- `POST /users/:id/avatar` (multipart `avatar`, ≤10MB, requires `users.update`)
  stores the image and replaces the previous object.
- `DELETE /users/:id/avatar` removes the stored object and clears the column.
- `avatar_url` in all user responses.

### 24.4 MinIO
- `docker-compose.yml` adds `minio-fiber-boilerplate` (ports 9000/9001,
  root `minioadmin/minioadmin`) for local S3-compatible testing.

## How to Test with MinIO

```bash
docker compose up -d minio-fiber-boilerplate
# create bucket + allow anonymous read (or via MinIO console)
S3_ACCESS_KEY_ID=minioadmin S3_SECRET_ACCESS_KEY=minioadmin \
S3_ENDPOINT=http://localhost:9000 S3_BUCKET=fiber-boilerplate \
go run cmd/app/main.go
cat avatar.png | curl -X POST -F "avatar=@-;type=image/png" \
  -H "Authorization: Bearer <token>" localhost:8080/api/v1/users/<id>/avatar
# avatar_url -> http://localhost:9000/fiber-boilerplate/avatars/<uuid>.png
```

Omit the S3 variables to exercise the local fallback; files land in
`storage/avatars/` and are served from `/uploads`.

## Verify
- Storage unit tests cover local round-trip, delete-missing no-op, default
  folder/URLs, S3 URL formats, provider selection.
- E2E against MinIO: upload → public URL (200), replace cleans the old object
  (404), delete removes the object (404). E2E local: file persisted,
  fetched via `/uploads`.