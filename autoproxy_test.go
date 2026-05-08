//go:build integration

package flux7tests

import (
	"encoding/json"
	"testing"

	"github.com/KTCrisis/flux7-tests/harness"
)

// Config 7 (solved): Two sessions — serve daemon + MCP auto-proxy.
func TestAutoProxy(t *testing.T) {
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

	// Start daemon
	proc, err := harness.StartProcess("mesh", meshBin, "serve", "--config", configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer proc.Stop()

	if err := harness.WaitHealthy("http://localhost:19090/health", 10e9); err != nil {
		t.Fatal(err)
	}

	// Start MCP client — should auto-proxy to daemon
	client, err := harness.NewMCPClient(meshBin, "--mcp", "--config", configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Initialize through proxy
	resp, err := client.Call("initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "proxy-test", "version": "1.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("initialize error: %s", resp.Error.Message)
	}

	// tools/list through proxy
	resp, err = client.Call("tools/list", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	var toolsResult struct {
		Tools []struct{ Name string } `json:"tools"`
	}
	json.Unmarshal(resp.Result, &toolsResult)
	if len(toolsResult.Tools) == 0 {
		t.Fatal("expected tools via proxy")
	}

	// tool call through proxy
	resp, err = client.Call("tools/call", map[string]any{
		"name":      "mesh.catalog",
		"arguments": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("mesh.catalog via proxy should work, got: %s", resp.Error.Message)
	}
}
