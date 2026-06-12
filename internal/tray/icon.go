//go:build cgo

package tray

func iconBytes() []byte {
	return trayIconBytes
}
