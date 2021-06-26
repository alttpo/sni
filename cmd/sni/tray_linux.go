package main

import (
	"os"
)

func createSystray() {
	// Nothing to do here
	<-make(chan struct{})
}

func quitSystray() {
	os.Exit(0)
}
