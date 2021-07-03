package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sni/cmd/sni/tray"
	"strings"
	"time"
)

import _ "net/http/pprof"

// include these SNES drivers:
import (
	"sni/snes/drivers/fxpakpro"
	"sni/snes/drivers/luabridge"
	"sni/snes/drivers/mock"
	"sni/snes/drivers/retroarch"
)

// build variables set via ldflags by `go build -ldflags="-X 'main.version=v1.0.0'"`:
var (
	version string = "v0.0.0"
	commit  string = "dirty"
	date    string = "2021-05-03T00:17:00Z"
	builtBy string = "go"
)

var (
	listenHost string // hostname/ip to listen on for webserver
	listenPort int    // port number to listen on for webserver
	logPath    string
)

var (
	cpuprofile = flag.String("cpuprofile", "", "start pprof profiler on addr:port")
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)

	ts := time.Now().Format("2006-01-02T15:04:05.000Z")
	ts = strings.ReplaceAll(ts, ":", "-")
	ts = strings.ReplaceAll(ts, ".", "-")
	logPath = filepath.Join(os.TempDir(), fmt.Sprintf("sni-%s.log", ts))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("could not open log file '%s' for writing\n", logPath)
	}

	log.Printf("sni %s %s built on %s by %s", version, commit, date, builtBy)
	log.Printf("logging to '%s'\n", logPath)
	log.SetOutput(io.MultiWriter(os.Stderr, logFile))
}

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		go func() {
			// "localhost:6060"
			log.Println(http.ListenAndServe(*cpuprofile, nil))
		}()
	}

	// explicitly initialize all the drivers:
	fxpakpro.DriverInit()
	luabridge.DriverInit()
	retroarch.DriverInit()
	mock.DriverInit()

	StartGrpcServer()
	StartHttpServer()

	// start up a systray:
	tray.CreateSystray(version, commit, date, builtBy)
}
