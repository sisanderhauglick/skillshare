package mcp_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"

	"skillshare/internal/mcp"
)

// readTOML is a test helper that reads and parses a TOML file into map[string]any.
func readTOML(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("readTOML read: %v", err)
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("readTOML unmarshal: %v", err)
	}
	return doc
}

// getMCPSectionTOML returns the mcp_servers section from a parsed TOML doc.
func getMCPSectionTOML(t *testing.T, doc map[string]any) map[string]any {
	t.Helper()
	raw, ok := doc["mcp_servers"]
	if !ok {
		t.Fatalf("expected key 'mcp_servers' in TOML document, got keys: %v", keys(doc))
	}
	section, ok := raw.(map[string]any)
	if !ok {
		t.Fatalf("'mcp_servers' is not a table, got: %T", raw)
	}
	return section
}

// codexTarget returns a test MCPTargetSpec for codex.
func codexTarget() mcp.MCPTargetSpec {
	return mcp.MCPTargetSpec{
		Name:   "codex",
		Key:    "mcp_servers",
		Format: "toml",
		Shared: true,
	}
}

// ---- readTOMLFile / writeTOMLFile (tested via MergeToTarget) ----

func TestMergeToTargetTOML_NewFile(t *testing.T) {
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "config.toml")

	servers := map[string]mcp.MCPServer{
		"fetch": {Command: "uvx", Args: []string{"mcp-server-fetch"}},
	}

	result, err := mcp.MergeToTarget(targetFile, servers, nil, codexTarget(), false)
	if err != nil {
		t.Fatalf("MergeToTarget error: %v", err)
	}

	if result.Target != "codex" {
		t.Errorf("Target = %q, want %q", result.Target, "codex")
	}
	if result.Error != "" {
		t.Errorf("unexpected error in result: %q", result.Error)
	}
	if len(result.Added) != 1 || result.Added[0] != "fetch" {
		t.Errorf("Added = %v, want [fetch]", result.Added)
	}
	if len(result.Updated) != 0 {
		t.Errorf("Updated = %v, want []", result.Updated)
	}
	if len(result.Removed) != 0 {
		t.Errorf("Removed = %v, want []", result.Removed)
	}

	doc := readTOML(t, targetFile)
	section := getMCPSectionTOML(t, doc)
	if _, ok := section["fetch"]; !ok {
		t.Error("expected 'fetch' in mcp_servers section")
	}
}

func TestMergeToTargetTOML_PreservesOtherSections(t *testing.T) {
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "config.toml")

	// Write a pre-existing config.toml with non-MCP content.
	existing := `model = "o4-mini"

[mcp_servers.user-tool]
command = "my-tool"
`
	if err := os.WriteFile(targetFile, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	servers := map[string]mcp.MCPServer{
		"fetch": {Command: "uvx", Args: []string{"mcp-server-fetch"}},
	}

	_, err := mcp.MergeToTarget(targetFile, servers, nil, codexTarget(), false)
	if err != nil {
		t.Fatalf("MergeToTarget error: %v", err)
	}

	doc := readTOML(t, targetFile)

	// Non-MCP key must be preserved.
	if doc["model"] != "o4-mini" {
		t.Errorf("model = %v, want 'o4-mini'", doc["model"])
	}

	section := getMCPSectionTOML(t, doc)

	// User entry must be preserved.
	if _, ok := section["user-tool"]; !ok {
		t.Error("user-tool should be preserved")
	}
	// Skillshare entry must be added.
	if _, ok := section["fetch"]; !ok {
		t.Error("fetch should have been added")
	}
}

func TestMergeToTargetTOML_RemovesPreviouslySynced(t *testing.T) {
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "config.toml")

	existing := `model = "o4-mini"

[mcp_servers.fetch]
command = "uvx"

[mcp_servers.old-tool]
command = "old"

[mcp_servers.user-tool]
command = "my-tool"
`
	if err := os.WriteFile(targetFile, []byte(existing), 0644); err != nil {
		t.Fatal(err)
	}

	servers := map[string]mcp.MCPServer{
		"fetch": {Command: "uvx", Args: []string{"mcp-server-fetch"}},
	}
	previousServers := []string{"fetch", "old-tool"}

	result, err := mcp.MergeToTarget(targetFile, servers, previousServers, codexTarget(), false)
	if err != nil {
		t.Fatalf("MergeToTarget error: %v", err)
	}

	if len(result.Removed) != 1 || result.Removed[0] != "old-tool" {
		t.Errorf("Removed = %v, want [old-tool]", result.Removed)
	}

	doc := readTOML(t, targetFile)
	section := getMCPSectionTOML(t, doc)

	if _, ok := section["old-tool"]; ok {
		t.Error("old-tool should have been removed")
	}
	if _, ok := section["fetch"]; !ok {
		t.Error("fetch should still be present")
	}
	if _, ok := section["user-tool"]; !ok {
		t.Error("user-tool should be preserved (not a managed entry)")
	}
	// Non-MCP key must survive.
	if doc["model"] != "o4-mini" {
		t.Errorf("model = %v, want 'o4-mini'", doc["model"])
	}
}

func TestMergeToTargetTOML_DryRun(t *testing.T) {
	dir := t.TempDir()
	targetFile := filepath.Join(dir, "config.toml")

	servers := map[string]mcp.MCPServer{
		"fetch": {Command: "uvx"},
	}

	result, err := mcp.MergeToTarget(targetFile, servers, nil, codexTarget(), true)
	if err != nil {
		t.Fatalf("MergeToTarget error: %v", err)
	}

	if len(result.Added) != 1 || result.Added[0] != "fetch" {
		t.Errorf("Added = %v, want [fetch]", result.Added)
	}

	// File must NOT have been created.
	if _, err := os.Stat(targetFile); !os.IsNotExist(err) {
		t.Error("dry run should not create the target file")
	}
}

// ---- transformServerTOML (tested via GenerateTargetTOML) ----

func TestTransformServerTOML_Stdio(t *testing.T) {
	servers := map[string]mcp.MCPServer{
		"context7": {
			Command: "npx",
			Args:    []string{"-y", "@upstash/context7-mcp"},
			Env:     map[string]string{"KEY": "VALUE"},
		},
	}

	data, err := mcp.GenerateTargetTOML(servers, codexTarget())
	if err != nil {
		t.Fatalf("GenerateTargetTOML error: %v", err)
	}

	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid TOML: %v", err)
	}

	section := getMCPSectionTOML(t, doc)
	srv, ok := section["context7"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'context7' table, got: %T", section["context7"])
	}

	if srv["command"] != "npx" {
		t.Errorf("command = %v, want 'npx'", srv["command"])
	}

	args, ok := srv["args"].([]any)
	if !ok || len(args) != 2 {
		t.Errorf("args = %v, want [-y @upstash/context7-mcp]", srv["args"])
	}

	env, ok := srv["env"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'env' table, got: %T", srv["env"])
	}
	if env["KEY"] != "VALUE" {
		t.Errorf("env.KEY = %v, want 'VALUE'", env["KEY"])
	}
}

func TestTransformServerTOML_Remote(t *testing.T) {
	servers := map[string]mcp.MCPServer{
		"remote-srv": {
			URL:     "https://mcp.example.com/sse",
			Headers: map[string]string{"Authorization": "MY_TOKEN_ENV"},
		},
	}

	data, err := mcp.GenerateTargetTOML(servers, codexTarget())
	if err != nil {
		t.Fatalf("GenerateTargetTOML error: %v", err)
	}

	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		t.Fatalf("invalid TOML: %v", err)
	}

	section := getMCPSectionTOML(t, doc)
	srv, ok := section["remote-srv"].(map[string]any)
	if !ok {
		t.Fatalf("expected 'remote-srv' table, got: %T", section["remote-srv"])
	}

	if srv["url"] != "https://mcp.example.com/sse" {
		t.Errorf("url = %v, want 'https://mcp.example.com/sse'", srv["url"])
	}
	if srv["bearer_token_env_var"] != "MY_TOKEN_ENV" {
		t.Errorf("bearer_token_env_var = %v, want 'MY_TOKEN_ENV'", srv["bearer_token_env_var"])
	}
	// Codex has no headers field.
	if _, ok := srv["headers"]; ok {
		t.Error("Codex TOML should not have 'headers' field")
	}
}
