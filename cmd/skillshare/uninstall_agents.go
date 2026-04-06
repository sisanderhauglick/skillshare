package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"skillshare/internal/oplog"
	"skillshare/internal/resource"
	"skillshare/internal/trash"
	"skillshare/internal/ui"
)

// cmdUninstallAgents removes agents from the source directory by moving them to agent trash.
func cmdUninstallAgents(agentsDir string, opts *uninstallOptions, cfgPath string, start time.Time) error {
	if _, err := os.Stat(agentsDir); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("agents source directory does not exist: %s", agentsDir)
		}
		return fmt.Errorf("cannot access agents source: %w", err)
	}

	// Resolve agent names
	var names []string
	if opts.all {
		discovered, err := resource.AgentKind{}.Discover(agentsDir)
		if err != nil {
			return fmt.Errorf("failed to discover agents: %w", err)
		}
		for _, d := range discovered {
			names = append(names, d.Name)
		}
		if len(names) == 0 {
			ui.Info("No agents found")
			return nil
		}
	} else {
		names = opts.skillNames
	}

	if len(names) == 0 {
		return fmt.Errorf("specify agent name(s) or --all")
	}

	// Validate all agents exist before removing any
	for _, name := range names {
		agentFile := filepath.Join(agentsDir, name+".md")
		if _, err := os.Stat(agentFile); err != nil {
			return fmt.Errorf("agent %q not found in %s", name, agentsDir)
		}
	}

	// Confirmation (unless --force or --json)
	if !opts.force && !opts.jsonOutput {
		ui.Warning("This will remove %d agent(s): %s", len(names), strings.Join(names, ", "))
		fmt.Print("Continue? [y/N] ")
		var input string
		fmt.Scanln(&input)
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "y" && input != "yes" {
			ui.Info("Cancelled")
			return nil
		}
	}

	trashBase := trash.AgentTrashDir()
	var removed []string
	var failed []string

	for _, name := range names {
		agentFile := filepath.Join(agentsDir, name+".md")
		metaFile := filepath.Join(agentsDir, name+".skillshare-meta.json")

		if opts.dryRun {
			ui.Info("[dry-run] Would remove agent: %s", name)
			removed = append(removed, name)
			continue
		}

		_, err := trash.MoveAgentToTrash(agentFile, metaFile, name, trashBase)
		if err != nil {
			ui.Error("Failed to remove %s: %v", name, err)
			failed = append(failed, name)
			continue
		}

		ui.Success("Removed agent: %s", name)
		removed = append(removed, name)
	}

	// JSON output
	if opts.jsonOutput {
		output := struct {
			Removed  []string `json:"removed"`
			Failed   []string `json:"failed"`
			DryRun   bool     `json:"dry_run"`
			Duration string   `json:"duration"`
		}{
			Removed:  removed,
			Failed:   failed,
			DryRun:   opts.dryRun,
			Duration: formatDuration(start),
		}
		var jsonErr error
		if len(failed) > 0 {
			jsonErr = fmt.Errorf("%d agent(s) failed to uninstall", len(failed))
		}
		return writeJSONResult(&output, jsonErr)
	}

	// Summary
	if !opts.dryRun {
		fmt.Println()
		ui.Info("%d agent(s) removed, %d failed", len(removed), len(failed))
		if len(removed) > 0 {
			ui.Info("Run 'skillshare sync agents' to update targets")
		}
	}

	// Oplog
	logUninstallAgentOp(cfgPath, names, len(removed), len(failed), opts.dryRun, start)

	if len(failed) > 0 {
		return fmt.Errorf("%d agent(s) failed to uninstall", len(failed))
	}
	return nil
}

func logUninstallAgentOp(cfgPath string, names []string, removed, failed int, dryRun bool, start time.Time) {
	status := "ok"
	if failed > 0 && removed > 0 {
		status = "partial"
	} else if failed > 0 {
		status = "error"
	}
	e := oplog.NewEntry("uninstall", status, time.Since(start))
	e.Args = map[string]any{
		"resource_kind": "agent",
		"names":         names,
		"removed":       removed,
		"failed":        failed,
		"dry_run":       dryRun,
	}
	oplog.WriteWithLimit(cfgPath, oplog.OpsFile, e, logMaxEntries()) //nolint:errcheck
}
