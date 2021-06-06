package main

import (
	"os"
)

func createSystray() {
	// just open the browser UI on startup:
	openWebUI()
}

func quitSystray() {
	os.Exit(0)
}
