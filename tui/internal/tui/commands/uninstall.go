package commands

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"crona/shared/config"
	"crona/tui/internal/api"
	"crona/tui/internal/logger"
	appruntime "crona/tui/internal/tui/runtime"

	tea "github.com/charmbracelet/bubbletea"
)

func UninstallCrona(c *api.Client, currentExecutablePath, kernelExecutablePath string, kernelInfo *api.KernelInfo) tea.Cmd {
	return func() tea.Msg {
		mode := ""
		if kernelInfo != nil {
			mode = kernelInfo.Env
		}
		if reason := appruntime.NonStandardRuntimeReason(currentExecutablePath, config.TUIBinaryNameForMode(mode)); reason != "" {
			return ErrMsg{Err: fmt.Errorf("%s Uninstall manually from your install directory", reason)}
		}

		installDir := filepath.Dir(strings.TrimSpace(currentExecutablePath))
		targets := uninstallTargets(installDir, currentExecutablePath, kernelExecutablePath, mode)
		baseDir := uninstallBaseDir(kernelInfo)

		if err := c.WipeRuntimeData(); err != nil {
			logger.Errorf("UninstallCrona wipe: %v", err)
			return ErrMsg{Err: err}
		}
		if err := c.ShutdownKernel(); err != nil {
			logger.Errorf("UninstallCrona shutdown: %v", err)
			return ErrMsg{Err: err}
		}
		if err := startUninstallCleanup(targets, baseDir); err != nil {
			logger.Errorf("UninstallCrona cleanup: %v", err)
			return ErrMsg{Err: err}
		}
		return UninstallStartedMsg{}
	}
}

func uninstallTargets(installDir, currentExecutablePath, kernelExecutablePath, mode string) []string {
	candidates := []string{
		strings.TrimSpace(currentExecutablePath),
		strings.TrimSpace(kernelExecutablePath),
		filepath.Join(installDir, config.CLIBinaryNameForMode(mode)),
		filepath.Join(installDir, config.TUIBinaryNameForMode(mode)),
		filepath.Join(installDir, config.KernelBinaryNameForMode(mode)),
	}
	seen := map[string]struct{}{}
	out := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		candidate = strings.TrimSpace(candidate)
		if candidate == "" {
			continue
		}
		candidate = filepath.Clean(candidate)
		if _, ok := seen[candidate]; ok {
			continue
		}
		seen[candidate] = struct{}{}
		out = append(out, candidate)
	}
	return out
}

func uninstallBaseDir(info *api.KernelInfo) string {
	if info == nil || strings.TrimSpace(info.ScratchDir) == "" {
		return ""
	}
	baseDir := filepath.Clean(filepath.Dir(strings.TrimSpace(info.ScratchDir)))
	switch baseDir {
	case "", ".", string(filepath.Separator):
		return ""
	default:
		return baseDir
	}
}

func startUninstallCleanup(targets []string, baseDir string) error {
	if runtime.GOOS == "windows" {
		return startWindowsUninstallCleanup(targets, baseDir)
	}
	return startPOSIXUninstallCleanup(targets, baseDir)
}

func startPOSIXUninstallCleanup(targets []string, baseDir string) error {
	scriptDir, err := os.MkdirTemp("", "crona-uninstall-*")
	if err != nil {
		return err
	}
	scriptPath := filepath.Join(scriptDir, "cleanup.sh")
	lines := []string{
		"#!/bin/sh",
		"sleep 2",
	}
	for _, target := range targets {
		lines = append(lines, "rm -f -- "+shellQuote(target))
	}
	if strings.TrimSpace(baseDir) != "" {
		lines = append(lines, "rm -rf -- "+shellQuote(baseDir))
	}
	lines = append(lines,
		"rm -f -- "+shellQuote(scriptPath),
		"rmdir -- "+shellQuote(scriptDir)+" 2>/dev/null || true",
	)
	if err := os.WriteFile(scriptPath, []byte(strings.Join(lines, "\n")+"\n"), 0o700); err != nil {
		return err
	}
	shellPath, err := exec.LookPath("sh")
	if err != nil {
		return err
	}
	cmd := exec.Command(shellPath, scriptPath)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func startWindowsUninstallCleanup(targets []string, baseDir string) error {
	scriptDir, err := os.MkdirTemp("", "crona-uninstall-*")
	if err != nil {
		return err
	}
	scriptPath := filepath.Join(scriptDir, "cleanup.ps1")
	lines := []string{
		"Start-Sleep -Seconds 2",
	}
	for _, target := range targets {
		lines = append(lines, "Remove-Item -LiteralPath '"+powershellQuote(target)+"' -Force -ErrorAction SilentlyContinue")
	}
	if strings.TrimSpace(baseDir) != "" {
		lines = append(lines, "Remove-Item -LiteralPath '"+powershellQuote(baseDir)+"' -Recurse -Force -ErrorAction SilentlyContinue")
	}
	lines = append(lines, "Remove-Item -LiteralPath $PSCommandPath -Force -ErrorAction SilentlyContinue")
	if err := os.WriteFile(scriptPath, []byte(strings.Join(lines, "\n")+"\n"), 0o600); err != nil {
		return err
	}
	powershellPath, err := exec.LookPath("powershell.exe")
	if err != nil {
		return err
	}
	cmd := exec.Command(powershellPath, "-NoProfile", "-ExecutionPolicy", "Bypass", "-File", scriptPath)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Start()
}

func shellQuote(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "'\"'\"'") + "'"
}

func powershellQuote(value string) string {
	return strings.ReplaceAll(value, "'", "''")
}
