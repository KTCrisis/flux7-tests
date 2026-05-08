//go:build integration

package flux7tests

import (
	"encoding/json"
	"testing"

	"github.com/KTCrisis/flux7-tests/harness"
)

// Config 1: Solo MCP — Claude spawns agent-mesh, no daemon, no mem7.
func TestSoloMCP(t *testing.T) {
	dir := t.TempDir()
	cfg := &harness.MeshConfig{
		Port:      19090,
		TraceFile: dir + "/traces.jsonl",
		Policies: []harness.PolicyDef{
			{Name: "test", Agent: "claude", Rules: []harness.RuleDef{
				{Tools: []string{"mesh.catalog"}, Action: "allow"},
				{Tools: []string{"*"}, Action: "deny"},
			}},
		},
	}
	configPath, err := cfg.WriteYAML(dir)
	if err != nil {
		t.Fatal(err)
	}

	client, err := harness.NewMCPClient(meshBin, "--mcp", "--config", configPath)
	if err != nil {
		t.Fatal(err)
	}
	defer client.Close()

	// Initialize
	resp, err := client.Call("initialize", map[string]any{
		"protocolVersion": "2024-11-05",
		"capabilities":    map[string]any{},
		"clientInfo":      map[string]any{"name": "test", "version": "1.0"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("initialize error: %s", resp.Error.Message)
	}

	// tools/list
	resp, err = client.Call("tools/list", map[string]any{})
	if err != nil {
		t.Fatal(err)
	}
	var toolsResult struct {
		Tools []struct {
			Name string `json:"name"`
		} `json:"tools"`
	}
	json.Unmarshal(resp.Result, &toolsResult)
	if len(toolsResult.Tools) == 0 {
		t.Fatal("expected at least virtual tools")
	}

	// tool call (allow) — mesh.catalog is always allowed
	resp, err = client.Call("tools/call", map[string]any{
		"name":      "mesh.catalog",
		"arguments": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error != nil {
		t.Fatalf("mesh.catalog should be allowed, got error: %s", resp.Error.Message)
	}

	// Unknown tool returns JSON-RPC error
	resp, err = client.Call("tools/call", map[string]any{
		"name":      "nonexistent.tool",
		"arguments": map[string]any{},
	})
	if err != nil {
		t.Fatal(err)
	}
	if resp.Error == nil {
		t.Fatal("expected error for unknown tool")
	}
}
