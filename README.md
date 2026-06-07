# flux7-tests

[![Integration](https://img.shields.io/github/actions/workflow/status/KTCrisis/flux7-tests/integration.yaml?style=flat-square&label=integration)](https://github.com/KTCrisis/flux7-tests/actions/workflows/integration.yaml)

Integration test suite for the flux7 stack: [flux7-mesh](https://github.com/KTCrisis/flux7-mesh) and [flux7-memory](https://github.com/KTCrisis/flux7-memory).

## Prerequisites

Binaries in `PATH` or `~/go/bin/`:

```bash
# mesh7
cd ~/flux7-mesh && go install ./cmd/mesh7/

# mem7 (optional — tests skip gracefully)
cd ~/flux7-memory && go install ./cmd/mem7/
```

Ports 19090 (mesh) and 19070 (mem7) must be free.

## Run

```bash
go test -tags=integration -v -timeout 2m
```

Single test:

```bash
go test -tags=integration -v -run TestAutoProxy
```

## Test matrix

| Test | Config | What it validates |
|---|---|---|
| `TestSoloMCP` | Solo MCP stdio | init, tools/list, tool call, unknown tool error |
| `TestServeDaemon` | `mesh7 serve` | health, /tools, /version endpoints |
| `TestAutoProxy` | Daemon + MCP auto-proxy | init, tools/list, tool call through proxy |
| `TestStreamableHTTP` | POST /mcp | session lifecycle (init, tools/list, DELETE) |
| `TestDurableGrants` | SQLite durable state | grant create, kill, restart, grant survives |
| `TestMeshWithMem7Daemon` | mesh + mem7 | both healthy, integration wired |

Full matrix (4 axes):

- **Mesh mode**: solo-mcp, serve-daemon, mcp-auto-proxy
- **mem7**: off, stdio, daemon
- **Transport**: mcp-stdio, http, mcp-streamable-http
- **Durable state**: off, on (kill + restart)

## Harness

The `harness/` package provides reusable test infrastructure:

- `Process` — start/stop binaries with context cancellation
- `WaitHealthy` — poll health endpoint until 200 or timeout
- `MCPClient` — stdio JSON-RPC client (send requests, read responses, skip non-JSON lines)
- `MeshConfig` — programmatic YAML config generation

## Structure

```
flux7-tests/
  harness/          # test infrastructure
    process.go      # binary lifecycle
    mcp.go          # MCP stdio client
    config.go       # YAML config builder
  setup_test.go     # TestMain, binary detection
  matrix_test.go    # matrix documentation
  solo_test.go      # Config 1: solo MCP
  serve_test.go     # Config 4: serve daemon
  autoproxy_test.go # Config 7: auto-proxy
  streamable_http_test.go # Config 8: streamable HTTP
  durable_test.go   # Durable state across restarts
  mem7_test.go      # mesh + mem7 integration
```
