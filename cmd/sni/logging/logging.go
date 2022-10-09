package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sni/cmd/sni/appversion"
	"sni/cmd/sni/config"
	"strings"
	"time"
)

var (
	Path string
)

func Init() {
	// include microseconds in timestamp and use UTC:
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)

	// create the log file:
	ts := time.Now().Format("2006-01-02T15:04:05.000Z")
	ts = strings.ReplaceAll(ts, ":", "-")
	ts = strings.ReplaceAll(ts, ".", "-")
	Path = filepath.Join(config.Dir, fmt.Sprintf("sni-%s-%s-%s-%s.log", runtime.GOOS, runtime.GOARCH, appversion.Version, ts))

	// open the log file for writing:
	logFile, err := os.OpenFile(Path, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("could not open log file '%s' for writing\n", Path)
	}
	// log to both stderr and log file:
	log.SetOutput(io.MultiWriter(os.Stderr, logFile))

	// first line should be the app version:
	log.Printf(
		"sni-%s-%s %s %s built on %s by %s",
		runtime.GOOS,
		runtime.GOARCH,
		appversion.Version,
		appversion.Commit,
		appversion.Date,
		appversion.BuiltBy)
	log.Printf("logging to '%s'\n", Path)
}
