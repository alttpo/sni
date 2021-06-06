package main

import (
	"fmt"
	"github.com/skratchdot/open-golang/open"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sni/util/env"
	"strconv"
	"strings"
	"time"
)

// include these SNES drivers:
import (
	_ "sni/snes/fxpakpro"
	_ "sni/snes/mock"
	_ "sni/snes/qusb2snes"
	_ "sni/snes/retroarch"
)

// build variables set via ldflags by goreleaser:
var (
	version string = "v0.0.0"
	commit  string = "dirty"
	date    string = "2021-05-03T00:17:00Z"
	builtBy string = "go"
)

var (
	listenHost  string // hostname/ip to listen on for webserver
	listenPort  int    // port number to listen on for webserver
	browserHost string // hostname to send as part of URL to browser to connect to webserver
	browserUrl  string // full URL that is sent to browser (composed of browserHost:listenPort)
	logPath     string
)

func orElse(a, b string) string {
	if a == "" {
		return b
	}
	return a
}

func init() {
	log.SetFlags(log.LstdFlags | log.Lmicroseconds | log.LUTC)

	ts := time.Now().Format("2006-01-02T15:04:05.000Z")
	ts = strings.ReplaceAll(ts, ":", "-")
	ts = strings.ReplaceAll(ts, ".", "-")
	logPath = filepath.Join(os.TempDir(), fmt.Sprintf("sni-%s.log", ts))
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		log.Printf("logging to '%s'\n", logPath)
		log.SetOutput(io.MultiWriter(os.Stderr, logFile))
	} else {
		log.Printf("could not open log file '%s' for writing\n", logPath)
	}

	log.Printf("sni %s %s built on %s by %s", version, commit, date, builtBy)
}

func main() {
	var err error

	// Parse env vars:
	listenHost = env.GetOrDefault("SNI_WEB_LISTEN_HOST", "0.0.0.0")

	listenPort, err = strconv.Atoi(env.GetOrDefault("SNI_WEB_LISTEN_PORT", "27637"))
	if err != nil {
		listenPort = 27637
	}
	if listenPort <= 0 {
		listenPort = 27637
	}
	listenAddr := net.JoinHostPort(listenHost, strconv.Itoa(listenPort))

	browserHost = env.GetOrDefault("SNI_WEB_BROWSER_HOST", "127.0.0.1")
	browserUrl = fmt.Sprintf("http://%s:%d/", browserHost, listenPort)

	// construct our web server:
	webServer := NewWebServer(listenAddr)

	// start the web server:
	go func() {
		log.Fatal(webServer.Serve())
	}()

	// start up a systray app (or just open web UI):
	createSystray()
}

func openWebUI() {
	err := open.Start(browserUrl)
	if err != nil {
		log.Println(err)
	}
}
