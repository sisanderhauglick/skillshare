package mcp

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

// readTOMLFile reads a TOML file into a map. Returns empty map if not found.
func readTOMLFile(path string) (map[string]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return make(map[string]any), nil
		}
		return nil, fmt.Errorf("read toml file %q: %w", path, err)
	}
	var doc map[string]any
	if err := toml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parse toml file %q: %w", path, err)
	}
	if doc == nil {
		doc = make(map[string]any)
	}
	return doc, nil
}

// writeTOMLFile encodes doc as TOML and writes it to path,
// creating parent directories as needed.
func writeTOMLFile(path string, data map[string]any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("create dir for %q: %w", path, err)
	}
	var buf bytes.Buffer
	enc := toml.NewEncoder(&buf)
	if err := enc.Encode(data); err != nil {
		return fmt.Errorf("encode toml for %q: %w", path, err)
	}
	if err := os.WriteFile(path, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("write toml file %q: %w", path, err)
	}
	return nil
}

// transformServerTOML converts MCPServer to Codex TOML format.
//
// Codex uses:
//   - command, args, env (nested table) for stdio servers
//   - url, bearer_token_env_var for remote servers (no headers field)
func transformServerTOML(srv MCPServer) map[string]any {
	m := make(map[string]any)
	if srv.IsRemote() {
		m["url"] = srv.URL
		// Codex uses bearer_token_env_var for authentication; if Authorization
		// header is present, extract the env var name from its value.
		if auth, ok := srv.Headers["Authorization"]; ok {
			m["bearer_token_env_var"] = auth
		}
	} else {
		m["command"] = srv.Command
		if len(srv.Args) > 0 {
			m["args"] = srv.Args
		}
	}
	if len(srv.Env) > 0 {
		env := make(map[string]any, len(srv.Env))
		for k, v := range srv.Env {
			env[k] = v
		}
		m["env"] = env
	}
	return m
}
