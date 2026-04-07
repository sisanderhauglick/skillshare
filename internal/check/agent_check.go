package check

import (
	"encoding/json"
	"os"
	"path/filepath"

	"skillshare/internal/install"
	"skillshare/internal/resource"
	"skillshare/internal/utils"
)

// AgentCheckResult holds the check result for a single agent.
type AgentCheckResult struct {
	Name    string `json:"name"`
	Source  string `json:"source,omitempty"`
	Version string `json:"version,omitempty"`
	RepoURL string `json:"repoUrl,omitempty"`
	Status  string `json:"status"` // "up_to_date", "drifted", "local", "error", "update_available"
	Message string `json:"message,omitempty"`
}

// CheckAgents scans the agents source directory for installed agents and
// compares their file hashes against metadata to detect drift.
// Uses resource.AgentKind{}.Discover() to recurse into subdirectories.
func CheckAgents(agentsDir string) []AgentCheckResult {
	discovered, err := resource.AgentKind{}.Discover(agentsDir)
	if err != nil {
		return nil
	}

	var results []AgentCheckResult
	for _, d := range discovered {
		result := checkOneAgent(d.SourcePath, d.RelPath)
		results = append(results, result)
	}

	return results
}

// checkOneAgent checks a single agent file. sourcePath is the absolute path
// to the .md file; relPath is relative to the agents root (e.g. "demo/code-reviewer.md").
func checkOneAgent(sourcePath, relPath string) AgentCheckResult {
	fileName := filepath.Base(relPath)
	agentName := fileName[:len(fileName)-len(".md")]
	result := AgentCheckResult{Name: relPath[:len(relPath)-len(".md")]}

	// Look for sidecar metadata: <basename>.skillshare-meta.json alongside the .md file
	dir := filepath.Dir(sourcePath)
	metaPath := filepath.Join(dir, agentName+".skillshare-meta.json")
	metaData, err := os.ReadFile(metaPath)
	if err != nil {
		result.Status = "local"
		return result
	}

	var meta install.SkillMeta
	if err := json.Unmarshal(metaData, &meta); err != nil {
		result.Status = "error"
		result.Message = "invalid metadata"
		return result
	}

	result.Source = meta.Source
	result.Version = meta.Version
	result.RepoURL = meta.RepoURL

	// Compare file hash
	if meta.FileHashes == nil || meta.FileHashes[fileName] == "" {
		result.Status = "local"
		return result
	}

	currentHash, err := utils.FileHashFormatted(sourcePath)
	if err != nil {
		result.Status = "error"
		result.Message = "cannot hash file"
		return result
	}

	if currentHash == meta.FileHashes[fileName] {
		result.Status = "up_to_date"
	} else {
		result.Status = "drifted"
		result.Message = "file content changed since install"
	}

	return result
}

// EnrichAgentResultsWithRemote checks agents that have RepoURL + Version
// against their remote HEAD to detect available updates.
// Uses ParallelCheckURLs for efficient batched remote probing.
func EnrichAgentResultsWithRemote(results []AgentCheckResult, onDone func()) {
	// Collect unique repo URLs that have version info
	type agentRef struct {
		repoURL string
		version string
		indices []int
	}
	urlMap := make(map[string]*agentRef)
	for i, r := range results {
		if r.RepoURL == "" || r.Version == "" {
			continue
		}
		if ref, ok := urlMap[r.RepoURL]; ok {
			ref.indices = append(ref.indices, i)
		} else {
			urlMap[r.RepoURL] = &agentRef{
				repoURL: r.RepoURL,
				version: r.Version,
				indices: []int{i},
			}
		}
	}

	if len(urlMap) == 0 {
		return
	}

	// Build URL check inputs
	var inputs []URLCheckInput
	var refs []*agentRef
	for _, ref := range urlMap {
		inputs = append(inputs, URLCheckInput{RepoURL: ref.repoURL})
		refs = append(refs, ref)
	}

	outputs := ParallelCheckURLs(inputs, onDone)

	// Apply results
	for i, out := range outputs {
		ref := refs[i]
		if out.Err != nil {
			continue
		}
		if out.RemoteHash != "" && out.RemoteHash != ref.version {
			for _, idx := range ref.indices {
				if results[idx].Status == "up_to_date" {
					results[idx].Status = "update_available"
					results[idx].Message = "newer version available"
				}
			}
		}
	}
}
