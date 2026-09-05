package e2e

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestVersionEndpoint(t *testing.T) {
	resp, raw, err := api("GET", "/version", nil, "")
	if err != nil {
		t.Fatalf("get version: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("version status %d: %s", resp.StatusCode, raw)
	}
	var v struct {
		App     string `json:"app"`
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &v); err != nil {
		t.Fatalf("version decode: %v (%s)", err, raw)
	}
	if v.Version == "" || v.App == "" {
		t.Fatalf("version/app missing: %s", raw)
	}
}
