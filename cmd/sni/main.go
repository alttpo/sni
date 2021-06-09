package main

import (
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"sni/protos/sni"
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
	listenHost string // hostname/ip to listen on for webserver
	listenPort int    // port number to listen on for webserver
	logPath    string
)

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
	listenHost = env.GetOrDefault("SNI_GRPC_LISTEN_HOST", "0.0.0.0")

	listenPort, err = strconv.Atoi(env.GetOrDefault("SNI_GRPC_LISTEN_PORT", "8191"))
	if err != nil {
		listenPort = 8191
	}
	if listenPort <= 0 {
		listenPort = 8191
	}
	listenAddr := net.JoinHostPort(listenHost, strconv.Itoa(listenPort))

	lis, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// start gRPC server:
	_ = listenAddr
	s := grpc.NewServer()
	sni.RegisterDevicesServer(s, &devicesService{})
	sni.RegisterMemoryUnaryServer(s, &memoryUnaryService{})
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// start up a systray handler if possible:
	createSystray()
}
