//go:build !windows
// +build !windows

package tray

func initConsole() (err error) {
	return nil
}

func consoleVisible(show bool) (err error) {
	return nil
}

func consoleIsDynamic() bool {
	return false
}
