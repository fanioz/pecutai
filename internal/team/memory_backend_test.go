package team

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/nex-crm/wuphf/internal/config"
)

func TestShouldPollNexNotificationsOnlyWhenNexIsActive(t *testing.T) {
	t.Setenv("WUPHF_NO_NEX", "")
	t.Setenv("WUPHF_API_KEY", "nex-test-key")
	t.Setenv("WUPHF_MEMORY_BACKEND", config.MemoryBackendNex)

	binDir := t.TempDir()
	nexMCP := filepath.Join(binDir, "nex-mcp")
	if err := os.WriteFile(nexMCP, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatalf("create fake nex-mcp: %v", err)
	}
	t.Setenv("PATH", binDir+string(os.PathListSeparator)+os.Getenv("PATH"))

	if !shouldPollNexNotifications() {
		t.Fatal("expected nex notification polling when nex backend is active")
	}
}
