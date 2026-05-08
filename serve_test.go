//go:build integration

package flux7tests

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/KTCrisis/flux7-tests/harness"
)

// Config 4: serve daemon — standalone HTTP, no MCP stdio.
func TestServeDaemon(t *testing.T) {
	dir := t.TempDir()
	cfg := &harness.MeshConfig{
		Port:        19090,
		StoragePath: dir + "/state.db",
		TraceFile:   dir + "/traces.jsonl",
		Policies: []harness.PolicyDef{
			{Name: "test", Agent: "*", Rules: []harness.RuleDef{
				{Tools: []string{"mesh.catalog"}, Action: "allow"},
				{Tools: []string{"*"}, Action: "deny"},
			}},
		},
	}
	configPath, _ := cfg.WriteYAML(dir)

	proc, err := harness.StartProcess("mesh", meshBin, "serve", "--config", configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer proc.Stop()

	if err := harness.WaitHealthy("http://localhost:19090/health", 10e9); err != nil {
		t.Fatal(err)
	}

	// Health
	resp, err := http.Get("http://localhost:19090/health")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var health map[string]any
	json.NewDecoder(resp.Body).Decode(&health)
	if health["status"] != "ok" {
		t.Fatalf("health: %v", health)
	}

	// Tools list — REST /tools only shows MCP server tools (empty when no servers configured)
	resp, err = http.Get("http://localhost:19090/tools")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	var tools []any
	json.Unmarshal(body, &tools)
	// No MCP servers → empty list is correct
	if tools == nil {
		t.Fatal("expected tools array (even if empty)")
	}

	// Version
	resp, err = http.Get("http://localhost:19090/version")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("version: status %d", resp.StatusCode)
	}
}
