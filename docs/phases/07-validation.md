# Phase 7 — Validation

> Scope: hardening the request validation pipeline with go-playground/validator
> and JSON-aware field error messages.

## Deliverables

### 7.1 Validator Wrapper (`internal/pkg/validator.go`)
- A singleton `*validator.Validate` built lazily by `NewValidator()`.
- `RegisterTagNameFunc(jsonTagName)`: field errors are keyed by the `json`
  tag name (e.g. `per_page`), not the Go field name.
- Custom `uuid` tag backed by `github.com/google/uuid`.
- `Validator.ValidateStruct(s) (map[string][]string, error)`:
  - nil, nil when the struct is valid;
  - field-name → list of human-readable messages otherwise.

### 7.2 Error Messages
- `required`, `email`, `uuid`, `min`, `max`, `oneof` produce friendly
  messages (`must be at least 8 characters`, `must be one of: admin, user`, …).

### 7.3 Integration
- User handler already calls `h.validate.ValidateStruct` after `BodyParser`
  and returns `pkg.ValidationResponse` (422) with the field map.
- All validation failures share the `VALIDATION_ERROR` envelope code.

### 7.4 Tests (`internal/pkg/validator_test.go`)
- Valid struct → no errors.
- Prominent tag violations → correct field → message mapping.
- JSON tag name (not Go field name) is used as the response key.

## How to Reproduce

```bash
go get github.com/go-playground/validator/v10
go test ./internal/pkg/... ./internal/user/...
```

## Verify
- `go test ./internal/pkg/...` passes.
- Invalid request body returns `422 VALIDATION_ERROR` with `fields` map.