package console

import (
	"fmt"
	"path/filepath"
)

// The make:* commands generate feature scaffolding under internal/<pkg>/
// mirroring the internal/user feature structure. Flags: --force overwrites.

// MakeModel scaffolds internal/<pkg>/model.go.
func MakeModel(name string, force bool) error {
	data, err := newFeatureName(name)
	if err != nil {
		return err
	}
	return writeGenerated([]generatedFile{
		{relPath: filepath.Join("internal", data.Pkg, "model.go"), tmpl: "model.go.tmpl"},
	}, data, force)
}

// MakeDTO scaffolds internal/<pkg>/dto.go.
func MakeDTO(name string, force bool) error {
	data, err := newFeatureName(name)
	if err != nil {
		return err
	}
	return writeGenerated([]generatedFile{
		{relPath: filepath.Join("internal", data.Pkg, "dto.go"), tmpl: "dto.go.tmpl"},
	}, data, force)
}

// MakeRepository scaffolds the repository interface plus its Postgres
// implementation.
func MakeRepository(name string, force bool) error {
	data, err := newFeatureName(name)
	if err != nil {
		return err
	}
	return writeGenerated([]generatedFile{
		{relPath: filepath.Join("internal", data.Pkg, "repository.go"), tmpl: "repository.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "postgres_repository.go"), tmpl: "postgres_repository.go.tmpl"},
	}, data, force)
}

// MakeService scaffolds the service and its business errors.
func MakeService(name string, force bool) error {
	data, err := newFeatureName(name)
	if err != nil {
		return err
	}
	return writeGenerated([]generatedFile{
		{relPath: filepath.Join("internal", data.Pkg, "service.go"), tmpl: "service.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "errors.go"), tmpl: "errors.go.tmpl"},
	}, data, force)
}

// MakeHandler scaffolds the HTTP handler with Swagger annotations.
func MakeHandler(name string, force bool) error {
	data, err := newFeatureName(name)
	if err != nil {
		return err
	}
	return writeGenerated([]generatedFile{
		{relPath: filepath.Join("internal", data.Pkg, "handler.go"), tmpl: "handler.go.tmpl"},
	}, data, force)
}

// MakeFeature scaffolds the complete feature (model, dto, repository, service,
// handler, errors) plus a create-table migration, and prints the wiring
// snippet to mount it.
func MakeFeature(name string, force bool) error {
	data, err := newFeatureName(name)
	if err != nil {
		return err
	}

	files := []generatedFile{
		{relPath: filepath.Join("internal", data.Pkg, "model.go"), tmpl: "model.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "dto.go"), tmpl: "dto.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "repository.go"), tmpl: "repository.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "postgres_repository.go"), tmpl: "postgres_repository.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "service.go"), tmpl: "service.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "handler.go"), tmpl: "handler.go.tmpl"},
		{relPath: filepath.Join("internal", data.Pkg, "errors.go"), tmpl: "errors.go.tmpl"},
	}

	next := nextMigrationNumber(outputRoot)
	up := fmt.Sprintf("migrations/%06d_create_%s.up.sql", next, data.Table)
	down := fmt.Sprintf("migrations/%06d_create_%s.down.sql", next, data.Table)
	files = append(files,
		generatedFile{relPath: up, tmpl: "migration.up.sql.tmpl"},
		generatedFile{relPath: down, tmpl: "migration.down.sql.tmpl"},
	)

	if err := writeGenerated(files, data, force); err != nil {
		return err
	}

	printl("\nNext steps to mount the feature:")
	printl("")
	printf("1. Build it in internal/app/app.go (Build):\n")
	printf("     %sRepo := %s.NewPostgresRepository(db)\n", data.Pkg, data.Pkg)
	printf("     %sSvc := %s.NewService(%sRepo)\n", data.Pkg, data.Pkg, data.Pkg)
	printf("     %sHandler := %s.NewHandler(%sSvc, validator)\n", data.Pkg, data.Pkg, data.Pkg)
	printf("     // add %sHandler to server.Dependencies and pass it to server.New\n", data.Pkg)
	printl("")
	printl("2. Mount it in internal/server/server.go (inside the /api/v1 group):")
	printf("     if deps.%sHandler != nil {\n", data.ID)
	printf("         deps.%sHandler.RegisterRoutes(v1, %s.RouteOptions{\n", data.ID, data.Pkg)
	printf("             Auth:          deps.AuthMiddleware,\n")
	printf("             RequireCreate: middleware.RequirePermission(deps.RBACService, \"%s.create\"),\n", data.Pkg)
	printf("             RequireView:   middleware.RequirePermission(deps.RBACService, \"%s.view\"),\n", data.Pkg)
	printf("             RequireUpdate: middleware.RequirePermission(deps.RBACService, \"%s.update\"),\n", data.Pkg)
	printf("             RequireDelete: middleware.RequirePermission(deps.RBACService, \"%s.delete\"),\n", data.Pkg)
	printl("         })")
	printl("     }")
	printl("")
	printl("3. The permission names must exist. Append to migrations/000007_seed_rbac.up.sql:")
	printf("     INSERT INTO permissions (name)\n")
	printf("     VALUES\n")
	for _, p := range []string{"create", "view", "update", "delete"} {
		printf("         ('%s.%s'),\n", data.Pkg, p)
	}
	printl("     ON CONFLICT DO NOTHING;")
	printl("")
	return nil
}
