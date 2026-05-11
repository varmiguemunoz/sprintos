package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/varmiguemunoz/command_pm_app/internal/domain"
)

type Session struct {
	UserID     uint   `json:"user_id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	Provider   string `json:"provider"`
	ProviderID string `json:"provider_id"`
}

func sessionPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("could not get user home directory: %w", err)
	}
	return filepath.Join(home, ".commandpm", "session.json"), nil
}

func SaveSession(user *domain.User) error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("could not create session directory: %w", err)
	}

	session := Session{
		UserID:     user.ID,
		Name:       user.Name,
		Email:      user.Email,
		Provider:   user.Provider,
		ProviderID: user.ProviderID,
	}

	data, err := json.Marshal(session)
	if err != nil {
		return fmt.Errorf("could not marshal session data: %w", err)
	}

	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("could not write session file: %w", err)
	}

	return nil
}

func LoadSession() (*Session, error) {
	path, err := sessionPath()
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("no active session found")
		}
		return nil, fmt.Errorf("could not read session file: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return nil, fmt.Errorf("could not unmarshal session data: %w", err)
	}

	return &session, nil
}

func ClearSession() error {
	path, err := sessionPath()
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("could not clear session: %w", err)
	}

	return nil
}
