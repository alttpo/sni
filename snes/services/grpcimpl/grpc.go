package grpcimpl

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"sni/cmd/sni/tray"
	"sni/protos/sni"
	"sni/util/env"
	"strconv"
	"time"
)

const fullMethodFormatter = "%32s"

var (
	ListenHost string
	ListenPort int
	ListenAddr string
	GrpcServer *grpc.Server
)

func StartGrpcServer() {
	var err error

	// Parse env vars:
	ListenHost = env.GetOrDefault("SNI_GRPC_LISTEN_HOST", "0.0.0.0")

	ListenPort, err = strconv.Atoi(env.GetOrDefault("SNI_GRPC_LISTEN_PORT", "8191"))
	if err != nil {
		ListenPort = 8191
	}
	if ListenPort <= 0 {
		ListenPort = 8191
	}

	ListenAddr = net.JoinHostPort(ListenHost, strconv.Itoa(ListenPort))
	lis, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// start gRPC server:
	GrpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(logTimingInterceptor),
		grpc.ChainStreamInterceptor(reportErrorStreamInterceptor),
	)
	sni.RegisterDevicesServer(GrpcServer, &DevicesService{})
	sni.RegisterDeviceMemoryServer(GrpcServer, &DeviceMemoryService{})
	sni.RegisterDeviceControlServer(GrpcServer, &DeviceControlService{})
	sni.RegisterDeviceFilesystemServer(GrpcServer, &DeviceFilesystem{})
	reflection.Register(GrpcServer)

	go func() {
		if err := GrpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
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
	if err != nil || tray.VerboseLogging {
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
	} else if tray.VerboseLogging {
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
