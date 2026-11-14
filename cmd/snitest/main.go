// Command snitest drives SNI over gRPC the way a file-transfer client does, to
// reproduce transfer failures through the whole daemon stack rather than by
// calling the driver directly.
//
// Every hardware test in this repo talks to the fxpakpro driver in-process,
// which skips the gRPC server, autoCloseableDevice, and the daemon's goroutine
// scheduling. The reported freeze happened through that stack, so this exists
// to exercise it: ListDevices, then MakeDirectory for each path component,
// PutFile, then ReadDirectory -- the sequence SNFM issues.
//
// PutFile is unary and blocking with no progress reporting, so a stalled
// transfer shows up here the same way it does for a user: the call simply does
// not return. Each call is therefore timed and reported.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"sni/protos/sni"
)

func main() {
	addr := flag.String("addr", "localhost:8191", "SNI gRPC address")
	dir := flag.String("dir", "unittest-grpc/sub", "directory to upload into")
	size := flag.Int("size", 4*1024*1024, "payload size in bytes")
	iterations := flag.Int("n", 20, "iterations")
	timeout := flag.Duration("timeout", 0, "per-call deadline; 0 means none")
	skipGet := flag.Bool("skipget", false, "skip GetFile verification (isolate the upload path)")
	skipLs := flag.Bool("skipls", false, "skip ReadDirectory")
	lsMissing := flag.Bool("lsmissing", false,
		"ReadDirectory a path that does not exist before each PutFile")
	flag.Parse()

	// A 4 MiB file plus protobuf framing exceeds gRPC's default 4MB receive
	// cap, so GetFile fails with ResourceExhausted unless this is raised. Worth
	// knowing for any client that reads whole ROMs back.
	const maxMsg = 64 * 1024 * 1024
	conn, err := grpc.NewClient(*addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(
			grpc.MaxCallRecvMsgSize(maxMsg),
			grpc.MaxCallSendMsgSize(maxMsg),
		))
	if err != nil {
		log.Fatalf("dial %s: %v", *addr, err)
	}
	defer conn.Close()

	devices := sni.NewDevicesClient(conn)
	fsys := sni.NewDeviceFilesystemClient(conn)

	// find a device
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	list, err := devices.ListDevices(ctx, &sni.DevicesRequest{})
	cancel()
	if err != nil {
		log.Fatalf("ListDevices: %v", err)
	}
	if len(list.GetDevices()) == 0 {
		log.Fatalf("no devices found; is a device connected and SNI running?")
	}
	uri := list.GetDevices()[0].GetUri()
	log.Printf("device: %s (%s)", uri, list.GetDevices()[0].GetDisplayName())

	// deterministic payload: each 4-byte word holds its own offset, so a
	// mismatch reports how far the data shifted
	payload := make([]byte, *size)
	for i := 0; i+4 <= len(payload); i += 4 {
		payload[i] = byte(i)
		payload[i+1] = byte(i >> 8)
		payload[i+2] = byte(i >> 16)
		payload[i+3] = byte(i >> 24)
	}

	call := func(name string, fn func(context.Context) error) time.Duration {
		ctx := context.Background()
		var cancel context.CancelFunc
		if *timeout > 0 {
			ctx, cancel = context.WithTimeout(ctx, *timeout)
			defer cancel()
		}
		start := time.Now()
		err := fn(ctx)
		d := time.Since(start)
		if err != nil {
			log.Printf("  %-16s FAILED after %v: %v", name, d.Round(time.Millisecond), err)
			os.Exit(1)
		}
		return d
	}

	// MakeDirectory for each path component, as SNFM does
	parts := strings.Split(*dir, "/")
	for i := range parts {
		p := strings.Join(parts[:i+1], "/")
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		_, err := fsys.MakeDirectory(ctx, &sni.MakeDirectoryRequest{Uri: uri, Path: p})
		cancel()
		if err != nil {
			log.Printf("  mkdir %-20s %v (continuing; it may already exist)", p, err)
		}
	}

	path := *dir + "/grpc-test.bin"
	log.Printf("uploading %d bytes to %s, %d iterations", *size, path, *iterations)

	var worst time.Duration
	for i := 1; i <= *iterations; i++ {
		// A client that checks whether a folder exists before creating it will
		// LS a missing path and get an error back. The firmware commits to a
		// data phase for LS regardless, and still emits the block holding the
		// 0xFF terminator on the error path, so a client that returns on the
		// error code leaves it unread and every command afterwards is one block
		// out of step. This reproduces that sequence deliberately.
		if *lsMissing {
			lctx, lcancel := context.WithTimeout(context.Background(), 20*time.Second)
			_, lerr := fsys.ReadDirectory(lctx, &sni.ReadDirectoryRequest{
				Uri: uri, Path: "unittest-no-such-dir-" + strconv.Itoa(i)})
			lcancel()
			if lerr == nil {
				log.Printf("[%2d] ReadDirectory of a missing path unexpectedly succeeded", i)
			}
		}

		put := call("PutFile", func(ctx context.Context) error {
			_, err := fsys.PutFile(ctx, &sni.PutFileRequest{Uri: uri, Path: path, Data: payload})
			return err
		})
		var ls time.Duration
		if !*skipLs {
			ls = call("ReadDirectory", func(ctx context.Context) error {
				_, err := fsys.ReadDirectory(ctx, &sni.ReadDirectoryRequest{Uri: uri, Path: *dir})
				return err
			})
		}
		var get time.Duration
		if !*skipGet {
			get = call("GetFile", func(ctx context.Context) error {
				rsp, err := fsys.GetFile(ctx, &sni.GetFileRequest{Uri: uri, Path: path})
				if err != nil {
					return err
				}
				data := rsp.GetData()
				if len(data) != len(payload) {
					return fmt.Errorf("read back %d bytes, sent %d", len(data), len(payload))
				}
				for j := range payload {
					if data[j] != payload[j] {
						return fmt.Errorf("contents differ at offset %d: got %02x want %02x",
							j, data[j], payload[j])
					}
				}
				return nil
			})
		}
		if put > worst {
			worst = put
		}
		log.Printf("[%2d] PutFile %v  ReadDirectory %v  GetFile+verify %v",
			i, put.Round(time.Millisecond), ls.Round(time.Millisecond), get.Round(time.Millisecond))
	}
	log.Printf("completed %d iterations; slowest PutFile %v", *iterations, worst.Round(time.Millisecond))
}
