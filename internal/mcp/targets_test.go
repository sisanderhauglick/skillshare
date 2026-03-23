package mcp_test

import (
	"strings"
	"testing"

	"skillshare/internal/mcp"
)

func TestMCPTargets_Load(t *testing.T) {
	targets, err := mcp.MCPTargets()
	if err != nil {
		t.Fatalf("MCPTargets() error: %v", err)
	}
	if len(targets) == 0 {
		t.Fatal("MCPTargets() returned empty slice")
	}
}

func TestMCPTargets_ClaudeFound(t *testing.T) {
	target, ok := mcp.LookupMCPTarget("claude")
	if !ok {
		t.Fatal("LookupMCPTarget(\"claude\") not found")
	}
	if target.Name != "claude" {
		t.Errorf("got Name=%q, want %q", target.Name, "claude")
	}
	if target.ProjectConfig != ".mcp.json" {
		t.Errorf("got ProjectConfig=%q, want %q", target.ProjectConfig, ".mcp.json")
	}
	if target.Key != "mcpServers" {
		t.Errorf("got Key=%q, want %q", target.Key, "mcpServers")
	}
}

func TestMCPTargets_CursorFound(t *testing.T) {
	target, ok := mcp.LookupMCPTarget("cursor")
	if !ok {
		t.Fatal("LookupMCPTarget(\"cursor\") not found")
	}
	if target.Name != "cursor" {
		t.Errorf("got Name=%q, want %q", target.Name, "cursor")
	}
	if target.GlobalConfig == "" {
		t.Error("cursor GlobalConfig should not be empty")
	}
	if target.ProjectConfig == "" {
		t.Error("cursor ProjectConfig should not be empty")
	}
}

func TestMCPTargets_LookupMissing(t *testing.T) {
	_, ok := mcp.LookupMCPTarget("nonexistent-tool")
	if ok {
		t.Error("LookupMCPTarget(\"nonexistent-tool\") should return false")
	}
}

func TestEffectiveURLKey_Default(t *testing.T) {
	// cursor has no url_key set — should default to "url"
	target, ok := mcp.LookupMCPTarget("cursor")
	if !ok {
		t.Fatal("cursor not found")
	}
	if got := target.EffectiveURLKey(); got != "url" {
		t.Errorf("EffectiveURLKey() = %q, want %q", got, "url")
	}
}

func TestEffectiveURLKey_Override(t *testing.T) {
	// Test with a constructed spec (no windsurf in MVP targets)
	spec := mcp.MCPTargetSpec{URLKey: "serverUrl"}
	if got := spec.EffectiveURLKey(); got != "serverUrl" {
		t.Errorf("EffectiveURLKey() = %q, want %q", got, "serverUrl")
	}
}

func TestGlobalConfigPath_ExpandsTilde(t *testing.T) {
	target, ok := mcp.LookupMCPTarget("cursor")
	if !ok {
		t.Fatal("cursor not found")
	}
	path := target.GlobalConfigPath()
	if strings.HasPrefix(path, "~") {
		t.Errorf("GlobalConfigPath() still has tilde: %q", path)
	}
	if path == "" {
		t.Error("GlobalConfigPath() returned empty string for cursor")
	}
}

func TestProjectConfigPath(t *testing.T) {
	target, ok := mcp.LookupMCPTarget("cursor")
	if !ok {
		t.Fatal("cursor not found")
	}
	projectRoot := "/some/project"
	path := target.ProjectConfigPath(projectRoot)
	expected := projectRoot + "/.cursor/mcp.json"
	if path != expected {
		t.Errorf("ProjectConfigPath() = %q, want %q", path, expected)
	}
}

func TestMCPTargets_CodexFound(t *testing.T) {
	target, ok := mcp.LookupMCPTarget("codex")
	if !ok {
		t.Fatal("LookupMCPTarget(\"codex\") not found")
	}
	if target.Name != "codex" {
		t.Errorf("got Name=%q, want %q", target.Name, "codex")
	}
	if target.GlobalConfig == "" {
		t.Error("codex GlobalConfig should not be empty")
	}
	if target.ProjectConfig == "" {
		t.Error("codex ProjectConfig should not be empty")
	}
	if target.Format != "toml" {
		t.Errorf("got Format=%q, want %q", target.Format, "toml")
	}
	if !target.Shared {
		t.Error("codex should be marked as shared")
	}
	if target.Key != "mcp_servers" {
		t.Errorf("got Key=%q, want %q", target.Key, "mcp_servers")
	}
}

func TestMCPTargets_FormatAndSharedFields(t *testing.T) {
	// claude: json + shared
	claude, ok := mcp.LookupMCPTarget("claude")
	if !ok {
		t.Fatal("claude not found")
	}
	if claude.Format != "json" {
		t.Errorf("claude.Format = %q, want %q", claude.Format, "json")
	}
	if !claude.Shared {
		t.Error("claude should be marked as shared")
	}

	// cursor: json + not shared
	cursor, ok := mcp.LookupMCPTarget("cursor")
	if !ok {
		t.Fatal("cursor not found")
	}
	if cursor.Format != "json" {
		t.Errorf("cursor.Format = %q, want %q", cursor.Format, "json")
	}
	if cursor.Shared {
		t.Error("cursor should NOT be marked as shared")
	}
}

func TestMCPTargetsForMode_GlobalExcludesProjectOnly(t *testing.T) {
	// In the current targets.yaml all three targets have global_config set,
	// so this test just validates the filtering logic doesn't break.
	globalTargets := mcp.MCPTargetsForMode(false)
	if len(globalTargets) == 0 {
		t.Error("expected at least one global target")
	}
}

func TestMCPTargetsForMode_ProjectExcludesGlobalOnly(t *testing.T) {
	// All current targets have project_config set; verify at least one is returned.
	projectTargets := mcp.MCPTargetsForMode(true)
	if len(projectTargets) == 0 {
		t.Error("expected at least one project-mode target")
	}
	// Sanity: a hypothetical global-only target (empty project_config) should be absent.
	projectNames := make(map[string]bool)
	for _, tgt := range projectTargets {
		projectNames[tgt.Name] = true
	}
	if projectNames["windsurf"] {
		t.Error("project mode should not include windsurf (global-only target)")
	}
}

func TestMCPTargetNames(t *testing.T) {
	names := mcp.MCPTargetNames()
	if len(names) == 0 {
		t.Fatal("MCPTargetNames() returned empty slice")
	}
	// Check a few known names
	nameSet := make(map[string]bool)
	for _, n := range names {
		nameSet[n] = true
	}
	for _, expected := range []string{"claude", "cursor", "codex"} {
		if !nameSet[expected] {
			t.Errorf("MCPTargetNames() missing %q", expected)
		}
	}
}
