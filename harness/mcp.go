package harness

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"sync/atomic"
)

type MCPClient struct {
	cmd    *exec.Cmd
	stdin  io.WriteCloser
	stdout *bufio.Scanner
	nextID atomic.Int64
}

type RPCRequest struct {
	JSONRPC string `json:"jsonrpc"`
	ID      int64  `json:"id,omitempty"`
	Method  string `json:"method"`
	Params  any    `json:"params,omitempty"`
}

type RPCResponse struct {
	JSONRPC string          `json:"jsonrpc"`
	ID      int64           `json:"id"`
	Result  json.RawMessage `json:"result,omitempty"`
	Error   *RPCError       `json:"error,omitempty"`
}

type RPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func NewMCPClient(bin string, args ...string) (*MCPClient, error) {
	cmd := exec.Command(bin, args...)
	stdin, _ := cmd.StdinPipe()
	stdout, _ := cmd.StdoutPipe()
	cmd.Stderr = nil

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("start MCP client: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, 4<<20), 4<<20)

	return &MCPClient{
		cmd:    cmd,
		stdin:  stdin,
		stdout: scanner,
	}, nil
}

func (c *MCPClient) Call(method string, params any) (*RPCResponse, error) {
	id := c.nextID.Add(1)
	req := RPCRequest{JSONRPC: "2.0", ID: id, Method: method, Params: params}
	data, _ := json.Marshal(req)
	data = append(data, '\n')

	if _, err := c.stdin.Write(data); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	for c.stdout.Scan() {
		line := c.stdout.Bytes()
		if len(line) == 0 || line[0] != '{' {
			continue
		}
		var resp RPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}
		return &resp, nil
	}
	return nil, fmt.Errorf("no response (stdin closed?)")
}

func (c *MCPClient) Notify(method string, params any) error {
	req := RPCRequest{JSONRPC: "2.0", Method: method, Params: params}
	data, _ := json.Marshal(req)
	data = append(data, '\n')
	_, err := c.stdin.Write(data)
	return err
}

func (c *MCPClient) Close() error {
	c.stdin.Close()
	return c.cmd.Wait()
}
