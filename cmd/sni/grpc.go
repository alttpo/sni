package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"sni/cmd/sni/service"
	"sni/protos/sni"
	"sni/util/env"
	"strconv"
	"time"
)

const fullMethodFormatter = "%32s"

var grpcServer *grpc.Server

func StartGrpcServer() {
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
	grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(logTimingInterceptor),
		grpc.ChainStreamInterceptor(reportErrorStreamInterceptor),
	)
	sni.RegisterDevicesServer(grpcServer, &service.DevicesService{})
	sni.RegisterDeviceMemoryServer(grpcServer, &service.DeviceMemoryService{})
	sni.RegisterDeviceControlServer(grpcServer, &service.DeviceControlService{})
	sni.RegisterDeviceFilesystemServer(grpcServer, &service.DeviceFilesystem{})
	reflection.Register(grpcServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
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
	if err != nil || verboseLogging {
		// format request message as string:
		if reqStringer, ok := info.Server.(methodRequestStringer); ok {
			reqStr = reqStringer.MethodRequestString(info.FullMethod, req)
		} else {
			reqStr = fmt.Sprintf("%+v", req)
		}
	}

	if err != nil {
		// log method, time taken, request, and error:
		log.Printf(fullMethodFormatter + ": %10d ns: req=`%s`, err=`%v`\n", info.FullMethod, tEnd.Sub(tStart).Nanoseconds(), reqStr, err)
	} else if verboseLogging {
		// only log normal requests+responses when verbose mode on:

		// format response message as string:
		if rspStringer, ok := info.Server.(methodResponseStringer); ok {
			rspStr = rspStringer.MethodResponseString(info.FullMethod, rsp)
		} else {
			rspStr = fmt.Sprintf("%+v", rsp)
		}

		// log method, time taken, request, and response:
		log.Printf(fullMethodFormatter + ": %10d ns: req=`%s`, rsp=`%s`\n", info.FullMethod, tEnd.Sub(tStart).Nanoseconds(), reqStr, rspStr)
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

	log.Printf(fullMethodFormatter + ": start stream from %s\n", info.FullMethod, streamSource)
	err = handler(srv, ss)
	if err != nil {
		log.Printf(fullMethodFormatter + ": end stream from %s; err=`%v`\n", info.FullMethod, streamSource, err)
	} else {
		log.Printf(fullMethodFormatter + ": end stream from %s\n", info.FullMethod, streamSource)
	}

	return
}
