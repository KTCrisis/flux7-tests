//go:build integration

package flux7tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/KTCrisis/flux7-tests/harness"
)

// Durable state: create grant → kill daemon → restart → grant survives.
func TestDurableGrants(t *testing.T) {
	dir := t.TempDir()
	cfg := &harness.MeshConfig{
		Port:        19090,
		StoragePath: dir + "/state.db",
		TraceFile:   dir + "/traces.jsonl",
		Policies: []harness.PolicyDef{
			{Name: "test", Agent: "*", Rules: []harness.RuleDef{
				{Tools: []string{"*"}, Action: "human_approval"},
			}},
		},
	}
	configPath, _ := cfg.WriteYAML(dir)

	// Start daemon
	proc, err := harness.StartProcess("mesh", meshBin, "serve", "--config", configPath)
	if err != nil {
		t.Fatal(err)
	}
	if err := harness.WaitHealthy("http://localhost:19090/health", 10e9); err != nil {
		proc.Stop()
		t.Fatal(err)
	}

	// Create a grant via HTTP
	grantBody, _ := json.Marshal(map[string]any{
		"agent":    "test-agent",
		"tools":    "mesh.*",
		"duration": "10m",
	})
	resp, err := http.Post("http://localhost:19090/grants", "application/json", bytes.NewReader(grantBody))
	if err != nil {
		proc.Stop()
		t.Fatal(err)
	}
	var grant map[string]any
	json.NewDecoder(resp.Body).Decode(&grant)
	resp.Body.Close()
	grantID, _ := grant["id"].(string)
	if grantID == "" {
		proc.Stop()
		t.Fatal("expected grant ID")
	}

	// Kill daemon
	proc.Stop()
	time.Sleep(500 * time.Millisecond)

	// Restart daemon
	proc2, err := harness.StartProcess("mesh", meshBin, "serve", "--config", configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer proc2.Stop()
	if err := harness.WaitHealthy("http://localhost:19090/health", 10e9); err != nil {
		t.Fatal(err)
	}

	// Check grant survived
	resp, err = http.Get("http://localhost:19090/grants")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var grants []map[string]any
	json.NewDecoder(resp.Body).Decode(&grants)

	found := false
	for _, g := range grants {
		if g["id"] == grantID {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("grant %s should survive restart, got %d grants", grantID, len(grants))
	}
}
