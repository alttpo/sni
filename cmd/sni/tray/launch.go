package tray

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

var runDll32 string = ""

type appConfig struct {
	Name    string
	Tooltip string

	Os string

	Dir  string
	Path string
	Args []string

	Url string
}

func launch(app *appConfig) {
	path := app.Path
	path = os.ExpandEnv(path)
	cleanPath := filepath.Clean(path)

	args := app.Args[:]
	// expand environment variables like `$SNI_USB2SNES_LISTEN_HOST`:
	for j, arg := range args {
		args[j] = os.ExpandEnv(arg)
	}

	dir := app.Dir
	dir = os.ExpandEnv(dir)

	if app.Url != "" {
		log.Printf("open: %s\n", app.Url)

		var cmd *exec.Cmd
		if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", app.Url)
		} else if runtime.GOOS == "windows" {
			if runDll32 == "" {
				runDll32 = filepath.Join(os.Getenv("SYSTEMROOT"), "System32", "rundll32.exe")
			}
			cmd = exec.Command(runDll32, "url.dll,FileProtocolHandler", app.Url)
		} else {
			cmd = exec.Command("xdg-open", app.Url)
		}

		err := cmd.Start()
		if err != nil {
			log.Printf("open: %s\n", err)
			return
		}

		return
	}

	if runtime.GOOS == "darwin" {
		if filepath.Ext(cleanPath) == ".app" {
			// open app bundles with "open" command:
			if fi, err := os.Stat(cleanPath); err == nil && fi.IsDir() {
				args = append([]string{"-a", path}, args...)
				path = "open"
			}
		}
	}

	log.Printf("open: %s %s\n", path, args)
	cmd := exec.Command(path, args...)
	cmd.Dir = dir
	err := cmd.Start()
	if err != nil {
		log.Printf("open: %s\n", err)
		return
	}
}
