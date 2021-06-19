package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"sni/examples/golang/sni"
)

func main() {
	var err error

	var conn *grpc.ClientConn
	conn, err = grpc.Dial("localhost:8191", grpc.WithInsecure())
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}
	defer conn.Close()

	client := sni.NewDevicesClient(conn)
	var rsp *sni.DevicesResponse
	rsp, err = client.ListDevices(context.Background(), &sni.DevicesRequest{})
	if err != nil {
		log.Fatalf("fail to list devices: %v", err)
	}

	for i, device := range rsp.Devices {
		fmt.Printf("[%d]: %+v\n", i, device)
	}
}
