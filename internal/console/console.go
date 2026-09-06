// Package console implements the artisan-style commands of the app binary:
// migration control, seeding, code generation, configuration checks and route
// listing. CLI mechanics (flags, arguments, help) live in cmd/app (cobra);
// this package only contains business logic.
package console

import (
	"bytes"
	"embed"
	"errors"
	"fmt"
	"go/format"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"
	"unicode"
)

//go:embed templates
var templatesFS embed.FS

// outputRoot is the project root used by the code generators. It exists to let
// tests redirect generated files into a temporary directory.
var outputRoot = "."

// out is the writer used by all console commands for human-facing output. It
// is swappable in tests.
var out io.Writer = os.Stdout

func printf(format string, a ...any) { fmt.Fprintf(out, format, a...) }
func printl(a ...any)                { fmt.Fprintln(out, a...) }

// featureName carries every derived identifier needed by the code generators.
type featureName struct {
	ID     string // PascalCase singular, e.g. Book
	Pkg    string // snake_case singular, e.g. book
	Table  string // snake_case plural table/route, e.g. books
	Plural string // PascalCase plural type name, e.g. Books
	Upper  string // UPPER_SNAKE singular error codes, e.g. BOOK
}

// newFeatureName normalizes a user-supplied name into the identifier set used
// by the templates. Inputs may be PascalCase (User) or snake_case (user).
func newFeatureName(raw string) (featureName, error) {
	id := toPascal(raw)
	if id == "" {
		return featureName{}, errors.New("name must be a valid identifier, e.g. 'user' or 'User'")
	}
	pkg := toSnake(id)
	return featureName{
		ID:     id,
		Pkg:    pkg,
		Table:  pluralize(pkg),
		Plural: toPascal(pluralize(pkg)),
		Upper:  strings.ToUpper(pkg),
	}, nil
}

// toPascal converts any identifier-ish input into PascalCase (User, UserProfile).
func toPascal(raw string) string {
	var b strings.Builder
	upperNext := true
	for _, r := range raw {
		if r == '-' || r == '_' || r == ' ' {
			upperNext = true
			continue
		}
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) {
			continue
		}
		if b.Len() == 0 {
			b.WriteRune(unicode.ToUpper(r))
			upperNext = false
			continue
		}
		if upperNext {
			b.WriteRune(unicode.ToUpper(r))
			upperNext = false
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// toSnake converts PascalCase into snake_case (UserProfile -> user_profile).
func toSnake(s string) string {
	var b strings.Builder
	runes := []rune(s)
	for i, r := range runes {
		if unicode.IsUpper(r) {
			if i > 0 && (unicode.IsLower(runes[i-1]) || unicode.IsDigit(runes[i-1])) {
				b.WriteByte('_')
			}
			b.WriteRune(unicode.ToLower(r))
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// pluralize applies naive English pluralization (books, categories, boxes).
func pluralize(s string) string {
	if s == "" {
		return s
	}
	lower := strings.ToLower(s)
	if strings.HasSuffix(lower, "ch") || strings.HasSuffix(lower, "sh") ||
		strings.HasSuffix(lower, "x") || strings.HasSuffix(lower, "z") ||
		strings.HasSuffix(lower, "s") {
		return s + "es"
	}
	if strings.HasSuffix(lower, "y") && len(s) > 2 && !isVowel(s[len(s)-2]) {
		return s[:len(s)-1] + "ies"
	}
	return s + "s"
}

func isVowel(b byte) bool {
	switch b {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return true
	}
	return false
}

// generatedFile is a single file the code generators produce.
type generatedFile struct {
	relPath string // path relative to outputRoot
	tmpl    string // template name under templates/
}

// writeGenerated renders every template into its target file, refuses to
// overwrite existing files unless force is set, and gofmts each result.
func writeGenerated(files []generatedFile, data featureName, force bool) error {
	for _, gf := range files {
		abs := filepath.Join(outputRoot, filepath.FromSlash(gf.relPath))
		if err := os.MkdirAll(filepath.Dir(abs), 0o755); err != nil {
			return err
		}
		if _, err := os.Stat(abs); err == nil {
			if !force {
				return fmt.Errorf("%s already exists (use --force to overwrite)", abs)
			}
		} else if !os.IsNotExist(err) {
			return err
		}

		raw, err := templatesFS.ReadFile("templates/" + gf.tmpl)
		if err != nil {
			return err
		}
		tmpl, err := template.New(gf.tmpl).Parse(string(raw))
		if err != nil {
			return fmt.Errorf("parse template %s: %w", gf.tmpl, err)
		}
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, data); err != nil {
			return fmt.Errorf("render template %s: %w", gf.tmpl, err)
		}

		source := buf.Bytes()
		if formatted, err := format.Source(source); err == nil {
			source = formatted
		}

		if err := os.WriteFile(abs, source, 0o644); err != nil {
			return err
		}
		printl("created " + gf.relPath)
	}
	return nil
}
