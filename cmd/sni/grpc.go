package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"sni/protos/sni"
	"sni/util/env"
	"strconv"
	"time"
)

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
	var serverOptions []grpc.ServerOption
	if *logTiming {
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(logTimingInterceptor))
	} else {
		serverOptions = append(serverOptions, grpc.ChainUnaryInterceptor(reportErrorInterceptor))
	}
	serverOptions = append(serverOptions, grpc.ChainStreamInterceptor(reportErrorStreamInterceptor))

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

		if err == nil {
			// format response message as string:
			if rspStringer, ok := info.Server.(methodResponseStringer); ok {
				rspStr = rspStringer.MethodResponseString(info.FullMethod, rsp)
			} else {
				rspStr = fmt.Sprintf("%+v", rsp)
			}

			// log method, time taken, request, and response:
			log.Printf("%26s: %10d ns: req=`%s`, rsp=`%s`\n", info.FullMethod, tEnd.Sub(tStart).Nanoseconds(), reqStr, rspStr)
		} else {
			// log method, time taken, request, and error:
			log.Printf("%26s: %10d ns: req=`%s`, err=`%v`\n", info.FullMethod, tEnd.Sub(tStart).Nanoseconds(), reqStr, err)
		}
	}()

	// invoke the method handler:
	rsp, err = handler(ctx, req)
	return
}

func reportErrorInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (rsp interface{}, err error) {
	// invoke the method handler:
	rsp, err = handler(ctx, req)

	if err != nil {
		// format request message as string:
		var reqStr string
		if reqStringer, ok := info.Server.(methodRequestStringer); ok {
			reqStr = reqStringer.MethodRequestString(info.FullMethod, req)
		} else {
			reqStr = fmt.Sprintf("%+v", req)
		}

		// log method, time taken, request, and error:
		log.Printf("%26s: req=`%s`, err=`%v`\n", info.FullMethod, reqStr, err)
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

	log.Printf("%26s: start stream from %s\n", info.FullMethod, streamSource)
	err = handler(srv, ss)
	if err != nil {
		log.Printf("%26s: end stream from %s; err=`%v`\n", info.FullMethod, streamSource, err)
	} else {
		log.Printf("%26s: end stream from %s\n", info.FullMethod, streamSource)
	}

	return
}
