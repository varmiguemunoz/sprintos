//go:build darwin

package tray

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/varmiguemunoz/sprintos/internal/config"
)

const plistID = "com.sprintos.tray"

const plistTmpl = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
	<key>Label</key>
	<string>{{.Label}}</string>
	<key>ProgramArguments</key>
	<array>
		<string>{{.BinaryPath}}</string>
		<string>tray</string>
	</array>
	<key>EnvironmentVariables</key>
	<dict>
		{{- range $k, $v := .Env}}
		<key>{{$k}}</key>
		<string>{{$v}}</string>
		{{- end}}
	</dict>
	<key>RunAtLoad</key>
	<true/>
	<key>KeepAlive</key>
	<true/>
	<key>StandardOutPath</key>
	<string>{{.LogOut}}</string>
	<key>StandardErrorPath</key>
	<string>{{.LogErr}}</string>
</dict>
</plist>
`

func plistPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, "Library", "LaunchAgents", plistID+".plist"), nil
}

func logDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	dir := filepath.Join(home, ".sprintos")
	return dir, os.MkdirAll(dir, 0755)
}

func IsInstalled() bool {
	path, err := plistPath()
	if err != nil {
		return false
	}
	_, err = os.Stat(path)
	return err == nil
}

func Install() error {
	binaryPath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("could not resolve binary path: %w", err)
	}

	ppath, err := plistPath()
	if err != nil {
		return err
	}

	logs, err := logDir()
	if err != nil {
		return err
	}

	env := map[string]string{}
	keys := []string{"DATABASE_URL", "GITHUB_CLIENT_ID", "GITHUB_CLIENT_SECRET",
		"SMTP_HOST", "SMTP_PORT", "SMTP_FROM", "SMTP_PASSWORD",
		"EVOLUTION_API_URL", "EVOLUTION_API_TOKEN"}
	for _, k := range keys {
		if v := os.Getenv(k); v != "" {
			env[k] = v
		}
	}
	if v := config.GetDatabaseURL(); v != "" {
		env["DATABASE_URL"] = v
	}
	if v := config.GetGitHubClientID(); v != "" {
		env["GITHUB_CLIENT_ID"] = v
	}
	if v := config.GetGitHubClientSecret(); v != "" {
		env["GITHUB_CLIENT_SECRET"] = v
	}

	data := map[string]interface{}{
		"Label":      plistID,
		"BinaryPath": binaryPath,
		"Env":        env,
		"LogOut":     filepath.Join(logs, "tray.log"),
		"LogErr":     filepath.Join(logs, "tray-error.log"),
	}

	tmpl, err := template.New("plist").Parse(plistTmpl)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(ppath), 0755); err != nil {
		return err
	}

	f, err := os.Create(ppath)
	if err != nil {
		return err
	}
	defer f.Close()

	if err := tmpl.Execute(f, data); err != nil {
		return err
	}

	_ = exec.Command("launchctl", "unload", ppath).Run()

	out, err := exec.Command("launchctl", "load", "-w", ppath).CombinedOutput()
	if err != nil {
		if !strings.Contains(string(out), "already loaded") {
			return fmt.Errorf("launchctl load: %s: %w", strings.TrimSpace(string(out)), err)
		}
	}

	return nil
}

func EnsureInstalled() error {
	if IsInstalled() {
		return nil
	}
	return Install()
}
