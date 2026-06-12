//go:build !darwin

package tray

import "fmt"

func IsInstalled() bool      { return false }
func EnsureInstalled() error { return nil }

func Install() error {
	return fmt.Errorf("menu bar app is only supported on macOS")
}

func Unload() error {
	return fmt.Errorf("menu bar app is only supported on macOS")
}

func Uninstall() error {
	return fmt.Errorf("menu bar app is only supported on macOS")
}
