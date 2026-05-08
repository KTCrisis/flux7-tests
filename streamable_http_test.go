//go:build integration

package flux7tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/KTCrisis/flux7-tests/harness"
)

// Config 8: MCP Streamable HTTP — POST /mcp with session management.
func TestStreamableHTTP(t *testing.T) {
	dir := t.TempDir()
	cfg := &harness.MeshConfig{
		Port:      19090,
		TraceFile: dir + "/traces.jsonl",
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

	// Initialize via POST /mcp
	initReq, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      1,
		"method":  "initialize",
		"params": map[string]any{
			"protocolVersion": "2024-11-05",
			"capabilities":    map[string]any{},
			"clientInfo":      map[string]any{"name": "http-test", "version": "1.0"},
		},
	})

	resp, err := http.Post("http://localhost:19090/mcp", "application/json", bytes.NewReader(initReq))
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	sessionID := resp.Header.Get("Mcp-Session-Id")
	if sessionID == "" {
		t.Fatal("expected Mcp-Session-Id header")
	}

	var initResp map[string]any
	json.Unmarshal(body, &initResp)
	if initResp["error"] != nil {
		t.Fatalf("initialize error: %v", initResp["error"])
	}

	// tools/list via POST /mcp with session
	listReq, _ := json.Marshal(map[string]any{
		"jsonrpc": "2.0",
		"id":      2,
		"method":  "tools/list",
		"params":  map[string]any{},
	})
	req, _ := http.NewRequest("POST", "http://localhost:19090/mcp", bytes.NewReader(listReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Mcp-Session-Id", sessionID)

	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ = io.ReadAll(resp.Body)
	resp.Body.Close()

	var listResp struct {
		Result struct {
			Tools []struct{ Name string } `json:"tools"`
		} `json:"result"`
	}
	json.Unmarshal(body, &listResp)
	if len(listResp.Result.Tools) == 0 {
		t.Fatal("expected tools via Streamable HTTP")
	}

	// DELETE /mcp — cleanup session
	delReq, _ := http.NewRequest("DELETE", "http://localhost:19090/mcp", nil)
	delReq.Header.Set("Mcp-Session-Id", sessionID)
	resp, err = http.DefaultClient.Do(delReq)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != 204 {
		t.Fatalf("expected 204 on DELETE, got %d", resp.StatusCode)
	}
}
