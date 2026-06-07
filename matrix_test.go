// Package flux7tests defines the integration test matrix for the flux7 stack.
//
// Run with: go test -tags=integration -v -timeout 5m
//
// Prerequisites:
//   - mesh7 binary in PATH or ~/go/bin/
//   - mem7 binary in PATH or ~/go/bin/
//   - No other process on ports 19090 (mesh) and 19070 (mem7)
//
// The matrix covers 4 axes:
//
//	Axis 1 — mesh mode:     solo-mcp | serve-daemon | mcp-auto-proxy
//	Axis 2 — mem7:          off | stdio | daemon
//	Axis 3 — transport:     mcp-stdio | http | mcp-streamable-http
//	Axis 4 — durable state: off | on (kill + restart)
//
// Each test case validates: init → tools/list → tool call (allow) →
// tool call (deny) → traces → health.
//
// Memory-enabled cases also validate: approval → decision write →
// auto-approve on repeat.

//go:build integration

package flux7tests
