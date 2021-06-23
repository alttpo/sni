package main

import (
	"context"
	"fmt"
	"google.golang.org/grpc"
	"log"
	"sni/examples/golang/sni"
)

func main() {
	log.SetFlags(log.Lmicroseconds | log.LUTC)

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

	memory := sni.NewDeviceMemoryClient(conn)
	read, err := memory.StreamRead(context.Background())
	if err != nil {
		return
	}

	mr := &sni.MultiReadMemoryRequest{
		Uri: rsp.Devices[0].Uri,
		Requests: []*sni.ReadMemoryRequest{
			{
				RequestAddress:       0x7e0010,
				RequestAddressSpace:  sni.AddressSpace_SnesABus,
				RequestMemoryMapping: sni.MemoryMapping_LoROM,
				Size:                 1,
			},
		},
	}

	for i := 0; i < 60*60; i++ {
		err = read.Send(mr)
		if err != nil {
			log.Fatal(err)
		}
		var mrsp *sni.MultiReadMemoryResponse
		mrsp, err = read.Recv()
		if err != nil {
			log.Fatal(err)
		}
		log.Println(mrsp.String())
	}
}
