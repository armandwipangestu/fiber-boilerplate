package console

import (
	"encoding/json"
	"fmt"
	"sort"
	"text/tabwriter"

	"github.com/armandwipangestu/fiber-boilerplate/internal/app"
	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

type routeRow struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Name   string `json:"name,omitempty"`
}

// RouteList assembles the Fiber app without opening a database connection and
// prints every registered route. With jsonOutput the routes are emitted as a
// machine-readable JSON array instead of a table.
func RouteList(cfg config.Config, jsonOutput bool) error {
	fiberApp, err := app.BuildForRoutes(cfg)
	if err != nil {
		return err
	}

	routes := fiberApp.GetRoutes(true)
	rows := make([]routeRow, 0, len(routes))
	for _, r := range routes {
		if r.Method == "" || r.Path == "" || r.Method == "HEAD" {
			continue
		}
		rows = append(rows, routeRow{Method: r.Method, Path: r.Path, Name: r.Name})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].Path == rows[j].Path {
			return rows[i].Method < rows[j].Method
		}
		return rows[i].Path < rows[j].Path
	})

	if jsonOutput {
		enc := json.NewEncoder(out)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	fmt.Fprintln(w, "METHOD\tPATH")
	for _, r := range rows {
		if r.Name != "" {
			fmt.Fprintf(w, "%s\t%s\t(%s)\n", r.Method, r.Path, r.Name)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\n", r.Method, r.Path)
	}
	return w.Flush()
}
