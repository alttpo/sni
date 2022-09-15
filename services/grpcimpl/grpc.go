package grpcimpl

import (
	"context"
	"fmt"
	"github.com/improbable-eng/grpc-web/go/grpcweb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"sni/cmd/sni/config"
	"sni/protos/sni"
	"sni/util"
	"sni/util/env"
	"strconv"
	"time"
)

const fullMethodFormatter = "%32s"

var (
	ListenHost string
	GrpcServer *grpc.Server
)

func StartGrpcServer() {
	// Parse env vars:
	ListenHost = env.GetOrDefault("SNI_GRPC_LISTEN_HOST", "0.0.0.0")

	const maxMessageSize = 100 * 1024 * 1024 // 100 MB

	// create gRPC server:
	GrpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(logTimingInterceptor),
		grpc.ChainStreamInterceptor(reportErrorStreamInterceptor),
		grpc.MaxMsgSize(maxMessageSize),
	)
	sni.RegisterDevicesServer(GrpcServer, &DevicesService{})
	sni.RegisterDeviceMemoryServer(GrpcServer, &DeviceMemoryService{})
	sni.RegisterDeviceControlServer(GrpcServer, &DeviceControlService{})
	sni.RegisterDeviceFilesystemServer(GrpcServer, &DeviceFilesystem{})
	sni.RegisterDeviceInfoServer(GrpcServer, &DeviceInfoService{})
	sni.RegisterDeviceNWAServer(GrpcServer, &DeviceNWAService{})
	reflection.Register(GrpcServer)

	go serveGrpc()
	go serveGrpcWeb()
}

func serveGrpc() {
	defer util.Recover()

	var err error

	var listenPort int
	listenPort, err = strconv.Atoi(env.GetOrDefault("SNI_GRPC_LISTEN_PORT", "8191"))
	if err != nil {
		listenPort = 8191
	}
	if listenPort <= 0 {
		listenPort = 8191
	}

	listenAddr := net.JoinHostPort(ListenHost, strconv.Itoa(listenPort))

	for {
		listenGrpc(listenAddr)
	}
}

func listenGrpc(listenAddr string) {
	defer func() {
		if pnk := recover(); pnk != nil {
			log.Printf("grpc: panic: %v\n", pnk)
		}
	}()

	lc := &net.ListenConfig{Control: util.ReusePortControl}

	var lis net.Listener
	var err error
	lis, err = lc.Listen(context.Background(), "tcp", listenAddr)
	if err != nil {
		log.Fatalf("grpc: failed to listen: %v", err)
	}

	log.Printf("grpc: listening on %s\n", listenAddr)
	if err := GrpcServer.Serve(lis); err != nil {
		log.Fatalf("grpc: failed to serve: %v", err)
	}
	log.Println("grpc: exit")
}

func serveGrpcWeb() {
	defer util.Recover()

	// wrap the GrpcServer with a GrpcWebServer:
	wrappedGrpc := grpcweb.WrapServer(
		GrpcServer,
		grpcweb.WithWebsockets(true),
		grpcweb.WithOriginFunc(func(origin string) bool { return true }),
		grpcweb.WithWebsocketOriginFunc(func(req *http.Request) bool { return true }),
	)

	//corsWrapper := wrappedGrpc
	corsWrapper := http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		// no CORS checking:
		if wrappedGrpc.IsGrpcWebSocketRequest(req) {
			wrappedGrpc.HandleGrpcWebsocketRequest(rw, req)
			return
		}
		if wrappedGrpc.IsGrpcWebRequest(req) {
			wrappedGrpc.HandleGrpcWebRequest(rw, req)
			return
		}

		// Likely an OPTIONS request:
		rw.Header().Add("Access-Control-Allow-Origin", "*")
		rw.Header().Add("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		rw.Header().Add("Access-Control-Allow-Headers", "*")

		rw.WriteHeader(http.StatusOK)
		_, _ = rw.Write(make([]byte, 0))
	})

	webListenPort := env.GetOrDefault("SNI_GRPCWEB_LISTEN_PORT", "8190")
	webListenAddr := net.JoinHostPort(ListenHost, webListenPort)

	for {
		listenGrpcWeb(webListenAddr, corsWrapper)
	}
}

func listenGrpcWeb(webListenAddr string, corsWrapper http.HandlerFunc) {
	defer func() {
		if pnk := recover(); pnk != nil {
			log.Printf("grpcweb: panic: %v\n", pnk)
		}
	}()

	var err error
	var lis net.Listener

	// attempt to start the usb2snes server:
	count := 0
	for {
		//lc := &net.ListenConfig{Control: util.ReusePortControl}
		lc := &net.ListenConfig{}
		lis, err = lc.Listen(context.Background(), "tcp", webListenAddr)
		if err == nil {
			break
		}

		if count == 0 {
			log.Printf("grpcweb: failed to listen on %s: %v\n", webListenAddr, err)
		}
		count++
		if count >= 30 {
			count = 0
		}

		time.Sleep(time.Second)
	}

	log.Printf("grpcweb: listening on %s\n", webListenAddr)
	err = http.Serve(lis, corsWrapper)
	log.Printf("grpcweb: exit listenHttp: %v\n", err)
}

type methodRequestStringer interface {
	MethodRequestString(method string, req interface{}) string
}

type methodResponseStringer interface {
	MethodResponseString(method string, rsp interface{}) string
}

func logTimingInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (rsp interface{}, err error) {
	// measure time taken for the call:
	tStart := time.Now()

	// invoke the method handler:
	rsp, err = handler(ctx, req)

	// stop timer:
	tEnd := time.Now()

	var reqStr, rspStr string
	if err != nil || config.VerboseLogging {
		// format request message as string:
		if reqStringer, ok := info.Server.(methodRequestStringer); ok {
			reqStr = reqStringer.MethodRequestString(info.FullMethod, req)
		} else {
			reqStr = fmt.Sprintf("%+v", req)
		}
	}

	if err != nil {
		// log method, time taken, request, and error:
		log.Printf(fullMethodFormatter+": %10d ns: req=`%s`, err=`%v`\n", info.FullMethod, tEnd.Sub(tStart).Nanoseconds(), reqStr, err)
	} else if config.VerboseLogging {
		// only log normal requests+responses when verbose mode on:

		// format response message as string:
		if rspStringer, ok := info.Server.(methodResponseStringer); ok {
			rspStr = rspStringer.MethodResponseString(info.FullMethod, rsp)
		} else {
			rspStr = fmt.Sprintf("%+v", rsp)
		}

		// log method, time taken, request, and response:
		log.Printf(fullMethodFormatter+": %10d ns: req=`%s`, rsp=`%s`\n", info.FullMethod, tEnd.Sub(tStart).Nanoseconds(), reqStr, rspStr)
	}

	return
}

func reportErrorStreamInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerInfo,
	handler grpc.StreamHandler,
) (err error) {
	streamSource := "unknown peer"
	if p, ok := peer.FromContext(ss.Context()); ok {
		streamSource = p.Addr.String()
	}

	log.Printf(fullMethodFormatter+": start stream from %s\n", info.FullMethod, streamSource)
	err = handler(srv, ss)
	if err != nil {
		log.Printf(fullMethodFormatter+": end stream from %s; err=`%v`\n", info.FullMethod, streamSource, err)
	} else {
		log.Printf(fullMethodFormatter+": end stream from %s\n", info.FullMethod, streamSource)
	}

	return
}
