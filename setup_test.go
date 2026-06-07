//go:build integration

package flux7tests

import (
	"os"
	"os/exec"
	"testing"
)

var (
	meshBin string
	mem7Bin string
	sup7Bin string
)

func TestMain(m *testing.M) {
	meshBin = findBinary("mesh7")
	mem7Bin = findBinary("mem7")
	sup7Bin = findBinary("sup7")

	if meshBin == "" {
		panic("mesh7 binary not found in PATH or ~/go/bin/")
	}

	os.Exit(m.Run())
}

func findBinary(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	home, _ := os.UserHomeDir()
	candidate := home + "/go/bin/" + name
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

func skipIfNoMem7(t *testing.T) {
	t.Helper()
	if mem7Bin == "" {
		t.Skip("mem7 binary not found, skipping mem7 tests")
	}
}

func skipIfNoSup7(t *testing.T) {
	t.Helper()
	if sup7Bin == "" {
		t.Skip("sup7 binary not found, skipping supervisor tests")
	}
}
