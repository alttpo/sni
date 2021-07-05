package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sni/cmd/sni/appversion"
	"strings"
	"time"
)

var (
	Dir  string
	Path string
)

func Init() {
	// decide on a logging output directory:
	if runtime.GOOS == "windows" {
		Dir = filepath.Join(os.Getenv("LOCALAPPDATA"), "sni")
	} else {
		var err error
		Dir, err = os.UserHomeDir()
		if err != nil {
			log.Printf("could not retrieve home directory: %s\n", err)
			return
		}
		Dir = filepath.Join(Dir, ".sni")
	}
	// make the directory if it doesn't exist:
	_ = os.MkdirAll(Dir, 0755|os.ModeDir)

	// include microseconds in timestamp and use UTC:
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)

	// create the log file:
	ts := time.Now().Format("2006-01-02T15:04:05.000Z")
	ts = strings.ReplaceAll(ts, ":", "-")
	ts = strings.ReplaceAll(ts, ".", "-")
	Path = filepath.Join(Dir, fmt.Sprintf("sni-%s.log", ts))

	// open the log file for writing:
	logFile, err := os.OpenFile(Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("could not open log file '%s' for writing\n", Path)
	}
	// log to both stderr and log file:
	log.SetOutput(io.MultiWriter(os.Stderr, logFile))

	// first line should be the app version:
	log.Printf(
		"sni %s %s built on %s by %s",
		appversion.Version,
		appversion.Commit,
		appversion.Date,
		appversion.BuiltBy,
	)
	log.Printf("logging to '%s'\n", Path)
}
