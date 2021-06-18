package main

import (
	"context"
	"flag"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sni/protos/sni"
	"sni/util/env"
	"strconv"
	"strings"
	"time"
)

import _ "net/http/pprof"

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

var (
	cpuprofile = flag.String("cpuprofile", "", "start pprof profiler on addr:port")
	logTiming  = flag.Bool("logtiming", false, "log gRPC method timings")
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
	flag.Parse()
	if *cpuprofile != "" {
		go func() {
			// "localhost:6060"
			log.Println(http.ListenAndServe(*cpuprofile, nil))
		}()
	}

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
	var serverOptions []grpc.ServerOption
	if *logTiming {
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(unaryInterceptor))
	}
	s := grpc.NewServer(serverOptions...)
	sni.RegisterDevicesServer(s, &devicesService{})
	sni.RegisterDeviceMemoryServer(s, &deviceMemoryService{})
	sni.RegisterDeviceControlServer(s, &deviceControlService{})
	reflection.Register(s)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	// start up a systray handler if possible:
	createSystray()
}

type methodRequestStringer interface {
	MethodRequestString(method string, req interface{}) string
}

type methodResponseStringer interface {
	MethodResponseString(method string, rsp interface{}) string
}

func unaryInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (rsp interface{}, err error) {
	// measure time taken for the call:
	tStart := time.Now()

	// report time taken:
	defer func() {
		// stop timer:
		tEnd := time.Now()
		// format request message as string:
		var reqStr, rspStr string
		if reqStringer, ok := info.Server.(methodRequestStringer); ok {
			reqStr = reqStringer.MethodRequestString(info.FullMethod, req)
		} else {
			reqStr = fmt.Sprintf("%+v", req)
		}
		// format response message as string:
		if rspStringer, ok := info.Server.(methodResponseStringer); ok {
			rspStr = rspStringer.MethodResponseString(info.FullMethod, rsp)
		} else {
			rspStr = fmt.Sprintf("%+v", rsp)
		}
		// log method, time taken, request, and response:
		log.Printf("%26s: %10d ns: req=`%s`, rsp=`%s`", info.FullMethod, tEnd.Sub(tStart).Nanoseconds(), reqStr, rspStr)
	}()

	// invoke the method handler:
	return handler(ctx, req)
}
