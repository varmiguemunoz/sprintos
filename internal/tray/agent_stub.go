//go:build !darwin

package tray

func IsInstalled() bool      { return false }
func Install() error         { return nil }
func EnsureInstalled() error { return nil }
