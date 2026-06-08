//go:build !cgo

package tray

import (
	"fmt"

	"gorm.io/gorm"
)

func Run(_ *gorm.DB) error {
	return fmt.Errorf("tray is not supported in this build (requires CGO)")
}
