package console

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// setOutputRoot redirects the generators into dir and restores the previous
// root afterwards.
func setOutputRoot(t *testing.T, dir string) {
	t.Helper()
	old := outputRoot
	outputRoot = dir
	t.Cleanup(func() { outputRoot = old })
}

// TestMakeFeature scaffolds a full feature into a temp dir and validates that
// every generated Go file parses.
func TestMakeFeature(t *testing.T) {
	root := t.TempDir()
	setOutputRoot(t, root)

	if err := MakeFeature("book", false); err != nil {
		t.Fatalf("MakeFeature: %v", err)
	}

	wantFiles := []string{
		"internal/book/model.go",
		"internal/book/dto.go",
		"internal/book/repository.go",
		"internal/book/postgres_repository.go",
		"internal/book/service.go",
		"internal/book/handler.go",
		"internal/book/errors.go",
		"migrations/000001_create_books.up.sql",
		"migrations/000001_create_books.down.sql",
	}
	for _, f := range wantFiles {
		if _, err := os.Stat(filepath.Join(root, filepath.FromSlash(f))); err != nil {
			t.Errorf("expected generated file %s: %v", f, err)
		}
	}

	for _, f := range []string{
		"internal/book/model.go",
		"internal/book/dto.go",
		"internal/book/repository.go",
		"internal/book/postgres_repository.go",
		"internal/book/service.go",
		"internal/book/handler.go",
		"internal/book/errors.go",
	} {
		parseGoFile(t, filepath.Join(root, filepath.FromSlash(f)))
	}
}

// TestMakeFeatureMigrationNumberedAfterExistingMigrations ensures the feature
// scaffold picks the next free migration version.
func TestMakeFeatureMigrationNumberedAfterExistingMigrations(t *testing.T) {
	root := t.TempDir()
	setOutputRoot(t, root)
	dir := filepath.Join(root, "migrations")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "000003_existing.up.sql"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := MakeFeature("album", false); err != nil {
		t.Fatalf("MakeFeature: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "000004_create_albums.up.sql")); err != nil {
		t.Errorf("expected migration 000004: %v", err)
	}
}

// TestMakeFeatureRefusesOverwrite checks that existing files are not clobbered
// without --force.
func TestMakeFeatureRefusesOverwrite(t *testing.T) {
	root := t.TempDir()
	setOutputRoot(t, root)
	if err := MakeFeature("pen", false); err != nil {
		t.Fatalf("first MakeFeature: %v", err)
	}
	if err := MakeFeature("pen", false); err == nil {
		t.Fatal("expected overwrite error without --force")
	} else if !strings.Contains(err.Error(), "already exists") {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := MakeFeature("pen", true); err != nil {
		t.Fatalf("MakeFeature --force: %v", err)
	}
}

// TestMakeSingleFiles ensures each make:x command generates its expected files.
func TestMakeSingleFiles(t *testing.T) {
	root := t.TempDir()
	setOutputRoot(t, root)

	tests := []struct {
		name string
		run  func(string, bool) error
		want []string
	}{
		{"model", MakeModel, []string{"internal/tag/model.go"}},
		{"dto", MakeDTO, []string{"internal/tag/dto.go"}},
		{"repository", MakeRepository, []string{"internal/tag/repository.go", "internal/tag/postgres_repository.go"}},
		{"service", MakeService, []string{"internal/tag/service.go", "internal/tag/errors.go"}},
		{"handler", MakeHandler, []string{"internal/tag/handler.go"}},
	}
	for _, tt := range tests {
		if err := tt.run("tag", false); err != nil {
			t.Fatalf("Make%s: %v", tt.name, err)
		}
		for _, f := range tt.want {
			abs := filepath.Join(root, filepath.FromSlash(f))
			if _, err := os.Stat(abs); err != nil {
				t.Errorf("make:%s: expected %s: %v", tt.name, f, err)
			}
			parseGoFile(t, abs)
		}
	}
}

func parseGoFile(t *testing.T, path string) {
	t.Helper()
	src, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	fset := token.NewFileSet()
	if _, err := parser.ParseFile(fset, path, src, parser.AllErrors); err != nil {
		t.Fatalf("generated %s does not parse: %v\n%s", path, err, src)
	}
}
