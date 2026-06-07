package harness

import (
	"fmt"
	"os"
	"path/filepath"
)

type MeshConfig struct {
	Port                int
	StoragePath         string
	TraceFile           string
	MemoryURL           string
	MemoryToken         string
	ApprovalTimeoutSecs int    // approval.timeout_seconds (0 = mesh default)
	ApprovalChannel     string // approval.channel: queue | tty | tty-fallback ("" = mesh default)
	MCPServers          []MCPServerDef
	OpenAPIs            []OpenAPIDef
	Policies            []PolicyDef
}

type MCPServerDef struct {
	Name      string
	Transport string
	Command   string
	Args      []string
}

type OpenAPIDef struct {
	File       string
	BackendURL string
}

type PolicyDef struct {
	Name  string
	Agent string
	Rules []RuleDef
}

type RuleDef struct {
	Tools  []string
	Action string
}

func (c *MeshConfig) WriteYAML(dir string) (string, error) {
	path := filepath.Join(dir, "config.yaml")
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	fmt.Fprintf(f, "port: %d\n", c.Port)
	if c.StoragePath != "" {
		fmt.Fprintf(f, "storage_path: %s\n", c.StoragePath)
	}
	if c.TraceFile != "" {
		fmt.Fprintf(f, "trace_file: %s\n", c.TraceFile)
	}
	if c.MemoryURL != "" {
		fmt.Fprintf(f, "memory:\n  url: %s\n", c.MemoryURL)
		if c.MemoryToken != "" {
			fmt.Fprintf(f, "  token: %s\n", c.MemoryToken)
		}
	}
	if c.ApprovalTimeoutSecs > 0 || c.ApprovalChannel != "" {
		fmt.Fprintln(f, "approval:")
		if c.ApprovalTimeoutSecs > 0 {
			fmt.Fprintf(f, "  timeout_seconds: %d\n", c.ApprovalTimeoutSecs)
		}
		if c.ApprovalChannel != "" {
			fmt.Fprintf(f, "  channel: %s\n", c.ApprovalChannel)
		}
	}
	if len(c.OpenAPIs) > 0 {
		fmt.Fprintln(f, "openapi:")
		for _, o := range c.OpenAPIs {
			fmt.Fprintf(f, "  - file: %s\n", o.File)
			if o.BackendURL != "" {
				fmt.Fprintf(f, "    backend_url: %s\n", o.BackendURL)
			}
		}
	}

	if len(c.MCPServers) > 0 {
		fmt.Fprintln(f, "mcp_servers:")
		for _, s := range c.MCPServers {
			fmt.Fprintf(f, "  - name: %s\n    transport: %s\n    command: %s\n", s.Name, s.Transport, s.Command)
			if len(s.Args) > 0 {
				fmt.Fprintf(f, "    args: [")
				for i, a := range s.Args {
					if i > 0 {
						fmt.Fprintf(f, ", ")
					}
					fmt.Fprintf(f, "%q", a)
				}
				fmt.Fprintln(f, "]")
			}
		}
	}

	if len(c.Policies) > 0 {
		fmt.Fprintln(f, "policies:")
		for _, p := range c.Policies {
			fmt.Fprintf(f, "  - name: %s\n    agent: %q\n    rules:\n", p.Name, p.Agent)
			for _, r := range p.Rules {
				fmt.Fprintf(f, "      - tools: [")
				for i, t := range r.Tools {
					if i > 0 {
						fmt.Fprintf(f, ", ")
					}
					fmt.Fprintf(f, "%q", t)
				}
				fmt.Fprintf(f, "]\n        action: %s\n", r.Action)
			}
		}
	}

	return path, nil
}
