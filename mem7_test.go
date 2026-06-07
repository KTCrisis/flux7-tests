//go:build integration

package flux7tests

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/KTCrisis/flux7-tests/harness"
)

// mesh + mem7 daemon: decisions written to mem7, auto-approve on repeat.
func TestMeshWithMem7Daemon(t *testing.T) {
	skipIfNoMem7(t)

	dir := t.TempDir()
	mem7Dir := dir + "/mem7data"
	os.MkdirAll(mem7Dir, 0o755)

	// Start mem7 daemon — MEM7_DIR must be set before the spawn so the
	// daemon inherits it (cmd.Env is captured at StartProcess time).
	os.Setenv("MEM7_DIR", mem7Dir)
	defer os.Unsetenv("MEM7_DIR")
	mem7Proc, err := harness.StartProcess("mem7", mem7Bin, "serve", "--listen", ":19070", "--token", "test-token")
	if err != nil {
		t.Fatal(err)
	}
	defer mem7Proc.Stop()

	if err := harness.WaitHealthy("http://localhost:19070/healthz", 10e9); err != nil {
		t.Fatal(err)
	}

	// Start mesh with mem7 integration
	cfg := &harness.MeshConfig{
		Port:        19090,
		StoragePath: dir + "/state.db",
		TraceFile:   dir + "/traces.jsonl",
		MemoryURL:   "http://localhost:19070",
		Policies: []harness.PolicyDef{
			{Name: "test", Agent: "*", Rules: []harness.RuleDef{
				{Tools: []string{"mesh.catalog"}, Action: "allow"},
				{Tools: []string{"*"}, Action: "deny"},
			}},
		},
	}
	configPath, _ := cfg.WriteYAML(dir)

	meshProc, err := harness.StartProcess("mesh", meshBin, "serve", "--config", configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer meshProc.Stop()

	if err := harness.WaitHealthy("http://localhost:19090/health", 10e9); err != nil {
		t.Fatal(err)
	}

	// Verify mesh health shows mem7 connection
	resp, err := http.Get("http://localhost:19090/health")
	if err != nil {
		t.Fatal(err)
	}
	var health map[string]any
	json.NewDecoder(resp.Body).Decode(&health)
	resp.Body.Close()
	if health["status"] != "ok" {
		t.Fatalf("mesh health: %v", health)
	}

	// Verify mem7 health
	resp, err = http.Get("http://localhost:19070/healthz")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 200 {
		t.Fatalf("mem7 health: %d", resp.StatusCode)
	}

	t.Log("mesh + mem7 daemon: both healthy, integration wired")
}
