// Package docs embeds the generated OpenAPI (Swagger) spec into the binary so
// the server can serve its own documentation from any working directory:
// `go run`, a dev shell, or a downloaded standalone binary.
package docs

import "embed"

// SwaggerFS holds the generated OpenAPI document.
//
//go:embed swagger/swagger.json swagger/swagger.yaml
var SwaggerFS embed.FS
