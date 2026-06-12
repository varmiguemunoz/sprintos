package mcpdetector

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
)

type AITool struct {
	Name       string
	ConfigPath string
	Detected   bool
	Configured bool
}

func configPath(parts ...string) string {
	home, _ := os.UserHomeDir()
	return filepath.Join(append([]string{home}, parts...)...)
}

func DetectTools() []AITool {
	tools := []AITool{
		{
			Name:       "Claude Desktop",
			ConfigPath: claudeConfigPath(),
		},
		{
			Name:       "Cursor",
			ConfigPath: configPath(".cursor", "mcp.json"),
		},
		{
			Name:       "Windsurf",
			ConfigPath: configPath(".codeium", "windsurf", "mcp_config.json"),
		},
		{
			Name:       "Zed",
			ConfigPath: configPath(".config", "zed", "settings.json"),
		},
		{
			Name:       "OpenCode",
			ConfigPath: configPath(".config", "opencode", "opencode.json"),
		},
		{
			Name:       "Codex",
			ConfigPath: configPath(".codex", "config.toml"),
		},
		{
			Name:       "Antigravity",
			ConfigPath: configPath(".gemini", "config", "mcp_config.json"),
		},
	}

	for i := range tools {
		if _, err := os.Stat(tools[i].ConfigPath); err == nil {
			tools[i].Detected = true
			tools[i].Configured = isConfigured(tools[i].ConfigPath)
		}
	}

	return tools
}

func claudeConfigPath() string {
	home, _ := os.UserHomeDir()
	switch runtime.GOOS {
	case "darwin":
		return filepath.Join(home, "Library", "Application Support", "Claude", "claude_desktop_config.json")
	case "windows":
		return filepath.Join(os.Getenv("APPDATA"), "Claude", "claude_desktop_config.json")
	default:
		return filepath.Join(home, ".config", "Claude", "claude_desktop_config.json")
	}
}

func isConfigured(configPath string) bool {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return false
	}
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return false
	}
	servers, ok := raw["mcpServers"].(map[string]interface{})
	if !ok {
		return false
	}
	_, exists := servers["sprintos"]
	return exists
}

func InstallMCP(tool *AITool, binaryPath string) error {
	if err := os.MkdirAll(filepath.Dir(tool.ConfigPath), 0755); err != nil {
		return fmt.Errorf("could not create config directory: %w", err)
	}

	var raw map[string]interface{}

	data, err := os.ReadFile(tool.ConfigPath)
	if err != nil {
		raw = map[string]interface{}{}
	} else {
		if err := json.Unmarshal(data, &raw); err != nil {
			raw = map[string]interface{}{}
		}
	}

	servers, ok := raw["mcpServers"].(map[string]interface{})
	if !ok {
		servers = map[string]interface{}{}
	}

	servers["sprintos"] = map[string]interface{}{
		"command": binaryPath,
		"args":    []string{"mcp"},
	}

	raw["mcpServers"] = servers

	out, err := json.MarshalIndent(raw, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode config: %w", err)
	}

	if err := os.WriteFile(tool.ConfigPath, out, 0644); err != nil {
		return fmt.Errorf("could not write config: %w", err)
	}

	tool.Configured = true
	return nil
}
