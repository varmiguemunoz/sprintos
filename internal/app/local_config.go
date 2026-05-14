package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type LocalConfig struct {
	ProjectID   uint   `json:"project_id"`
	ProjectName string `json:"project_name"`
	OrgID       uint   `json:"org_id"`
}

const localConfigFile = ".sprintos"

func SaveLocalConfig(cfg LocalConfig) error {
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("could not encode config: %w", err)
	}
	return os.WriteFile(localConfigFile, data, 0644)
}

func LoadLocalConfig() (*LocalConfig, error) {
	dir, err := os.Getwd()
	if err != nil {
		return nil, err
	}

	for {
		path := filepath.Join(dir, localConfigFile)
		data, err := os.ReadFile(path)
		if err == nil {
			var cfg LocalConfig
			if err := json.Unmarshal(data, &cfg); err != nil {
				return nil, fmt.Errorf("invalid .sprintos file: %w", err)
			}
			return &cfg, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return nil, fmt.Errorf("no .sprintos file found — run `sprintos init`")
}

func ResolveProjectID(flagValue uint, _ *StateService) (uint, error) {
	if flagValue > 0 {
		return flagValue, nil
	}
	cfg, err := LoadLocalConfig()
	if err != nil {
		return 0, fmt.Errorf("--project not specified and no .sprintos file found — run `sprintos init`")
	}
	return cfg.ProjectID, nil
}

func ResolveStateByName(stateSvc *StateService, projectID uint, nameOrID string) (uint, error) {
	var id uint
	if _, err := fmt.Sscanf(nameOrID, "%d", &id); err == nil && id > 0 {
		return id, nil
	}

	states, err := stateSvc.ListByProject(projectID)
	if err != nil {
		return 0, fmt.Errorf("could not load states: %w", err)
	}

	for _, s := range states {
		if equalFold(s.Name, nameOrID) {
			return s.ID, nil
		}
	}

	return 0, fmt.Errorf("state %q not found in project %d", nameOrID, projectID)
}

func equalFold(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		ca, cb := a[i], b[i]
		if ca >= 'A' && ca <= 'Z' {
			ca += 32
		}
		if cb >= 'A' && cb <= 'Z' {
			cb += 32
		}
		if ca != cb {
			return false
		}
	}
	return true
}

type ProjectInfo struct {
	ID   uint
	Name string
}
