package fxpakpro

import (
	"bytes"
	"context"
	"os"
	"testing"
	"time"
)

// TestDevice_errorPathStaysAligned checks that a command which fails with a
// device error code leaves the protocol stream aligned, by issuing a normal
// command afterwards and seeing whether it still gets a sane reply.
//
// This matters because the firmware commits to a data phase before it knows
// whether the command succeeded:
//
//   - LS sets HANDLE_DAT regardless of f_opendir, and still emits a block
//     holding the 0xFF terminator on the error path.
//   - GET sets HANDLE_DAT regardless of f_stat/f_open, with server_info.size
//     taken from a FILINFO that a failed f_stat never wrote.
//   - PUT sets cmdDat=1 in usbint_recv_block before f_open is even attempted,
//     then waits in HANDLE_LOCK for the whole payload; server_state only leaves
//     HANDLE_LOCK once count >= server_info.size in usbint_recv_block.
//
// SNI returns as soon as it sees the error code in every case, so anything the
// device still expects to send or receive is left in the pipe.
//
// Select which path to exercise with SNI_TEST_ERROR_PATH=ls|get|put. One per
// run, because a failure here can wedge the device.
func TestDevice_errorPathStaysAligned(t *testing.T) {
	which := os.Getenv("SNI_TEST_ERROR_PATH")
	if which == "" {
		t.Skip("set SNI_TEST_ERROR_PATH to ls, get or put")
	}

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	if _, _, rom, err := d.info(ctx); err != nil {
		t.Fatalf("baseline info(): %v", err)
	} else {
		t.Logf("baseline ok, rom=%q", rom)
	}

	const missingDir = "unittest-does-not-exist-xyz"
	const missingFile = missingDir + "/nope.bin"

	start := time.Now()
	var opErr error
	switch which {
	case "ls":
		_, opErr = d.listFiles(ctx, missingDir)
	case "get":
		var w bytes.Buffer
		_, opErr = d.getFile(ctx, missingFile, &w, nil, nil)
	case "put":
		// a path whose parent does not exist, so f_open fails on the device
		payload := make([]byte, 4096)
		_, opErr = d.putFile(ctx, missingFile, uint32(len(payload)), bytes.NewReader(payload), nil)
	default:
		t.Fatalf("SNI_TEST_ERROR_PATH=%q: want ls, get or put", which)
	}

	if opErr == nil {
		t.Fatalf("%s on a missing path unexpectedly succeeded", which)
	}
	t.Logf("%s failed as expected in %v: %v", which, time.Since(start), opErr)

	// The real question: is the stream still aligned?
	probeStart := time.Now()
	version, _, _, err := d.info(ctx)
	probeElapsed := time.Since(probeStart)
	if err != nil {
		t.Fatalf("DESYNCED: info() after a failed %s took %v and failed: %v",
			which, probeElapsed, err)
	}
	if version == "" {
		t.Fatalf("DESYNCED: info() after a failed %s returned an empty version "+
			"(read someone else's block)", which)
	}
	t.Logf("stream still aligned after a failed %s: version=%q (%v)", which, version, probeElapsed)

	// and a second one, in case the damage is one command further along
	if v2, _, _, err := d.info(ctx); err != nil || v2 == "" {
		t.Fatalf("DESYNCED on the second command after a failed %s: version=%q err=%v",
			which, v2, err)
	}
	t.Logf("second command also fine")
}

// TestDevice_putErrorRecovers verifies the fix for the failed-PUT desync
// through devices.AutoCloseableDevice, the layer services/grpcimpl uses. A
// command error from PUT is reported as fatal, so ensureOpened closes and
// deletes the device before anything else is written; the firmware's
// usbint_check_connect() then resets server_state and cmdDat on the disconnect,
// and the following request reopens onto a clean device.
//
// Before the fix this same sequence bricked the pak: the error was non-fatal,
// SNI carried on, and the next command's 512 bytes were consumed as file data.
func TestDevice_putErrorRecovers(t *testing.T) {
	d := openAutoCloseableDevice(t)
	defer d.Close()

	ctx := context.Background()

	if v, err := d.FetchFields(ctx, 0); err != nil {
		t.Fatalf("baseline FetchFields(): %v", err)
	} else {
		t.Logf("baseline ok: %v", v)
	}

	const missingFile = "unittest-does-not-exist-xyz/nope.bin"
	payload := make([]byte, 4096)

	start := time.Now()
	_, err := d.PutFile(ctx, missingFile, uint32(len(payload)), bytes.NewReader(payload), nil)
	if err == nil {
		t.Fatalf("PutFile to a missing directory unexpectedly succeeded")
	}
	t.Logf("PutFile failed as expected in %v: %v", time.Since(start), err)

	// The device should have been closed and dropped on that fatal error, so
	// this call reopens onto a device the disconnect has reset.
	probe := time.Now()
	v, err := d.FetchFields(ctx, 0)
	if err != nil {
		t.Fatalf("NOT RECOVERED: FetchFields() after a failed PutFile took %v: %v",
			time.Since(probe), err)
	}
	t.Logf("recovered in %v: %v", time.Since(probe), v)

	// exercise it a bit further to be sure the stream is genuinely aligned
	for i := 0; i < 3; i++ {
		if _, err := d.ReadDirectory(ctx, ""); err != nil {
			t.Fatalf("ReadDirectory() %d after recovery: %v", i, err)
		}
	}
	t.Logf("device fully functional after recovery")
}
