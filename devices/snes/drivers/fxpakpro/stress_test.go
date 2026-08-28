package fxpakpro

import (
	"bytes"
	"context"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"testing"
	"time"

	"sni/devices"
	"sni/protos/sni"
)

// filePattern builds deterministic content that encodes both the byte offset
// and a per-file tag. A stream desync shows up as a shifted offset; data from
// the wrong file shows up as a wrong tag. Both are recoverable from a single
// mismatched word.
func filePattern(tag, size uint32) []byte {
	d := make([]byte, size)
	i := uint32(0)
	for ; i+4 <= size; i += 4 {
		binary.LittleEndian.PutUint32(d[i:], i^tag)
	}
	for ; i < size; i++ {
		d[i] = byte(i ^ tag)
	}
	return d
}

func describeStressMismatch(actual, expected []byte, tag uint32) string {
	for i := range expected {
		if i >= len(actual) {
			return fmt.Sprintf("truncated at offset %d ($%06x)", i, i)
		}
		if actual[i] == expected[i] {
			continue
		}
		lo := i &^ 3
		var got, want uint32
		if lo+4 <= len(actual) {
			got = binary.LittleEndian.Uint32(actual[lo:])
		}
		if lo+4 <= len(expected) {
			want = binary.LittleEndian.Uint32(expected[lo:])
		}
		return fmt.Sprintf(
			"differs at offset %d ($%06x): got word $%08x, want $%08x; "+
				"decoded offset %d (shift %d), decoded tag $%08x (want $%08x)",
			i, i, got, want, got^tag, int64(got^tag)-int64(lo), got^uint32(lo), tag)
	}
	return "contents match"
}

// awkwardSizes clusters around the protocol's 64- and 512-byte block
// boundaries, where the padding and remainder paths in sendSerialProgress and
// getFile behave differently.
var awkwardSizes = []uint32{
	1, 2, 3, 63, 64, 65, 127, 128, 129, 255, 256, 257,
	511, 512, 513, 767, 1023, 1024, 1025, 1535, 2047, 2048, 2049,
	4095, 4096, 4097, 8191, 8193, 12289, 65535, 65537, 100003, 131071,
}

type stressFile struct {
	path string
	tag  uint32
	size uint32
}

// TestDevice_stressMixed interleaves PUT, GET, LS, MKDIR, RM and INFO in a
// seeded random order, using file sizes that are deliberately not multiples of
// the 512-byte block size. Every operation is logged so a failure can be
// replayed with SNI_TEST_SEED.
func TestDevice_stressMixed(t *testing.T) {
	seed := time.Now().UnixNano()
	if v := os.Getenv("SNI_TEST_SEED"); v != "" {
		s, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			t.Fatalf("SNI_TEST_SEED=%q: %v", v, err)
		}
		seed = s
	}
	ops := 300
	if v := os.Getenv("SNI_TEST_OPS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil {
			t.Fatalf("SNI_TEST_OPS=%q: %v", v, err)
		}
		ops = n
	}
	// the card is nearly full, so keep live data well inside free space:
	maxLive := uint32(3 * 1024 * 1024)
	if v := os.Getenv("SNI_TEST_MAX_LIVE"); v != "" {
		n, err := strconv.ParseUint(v, 0, 32)
		if err != nil {
			t.Fatalf("SNI_TEST_MAX_LIVE=%q: %v", v, err)
		}
		maxLive = uint32(n)
	}

	t.Logf("seed=%d ops=%d maxLive=%d (replay with SNI_TEST_SEED=%d)", seed, ops, maxLive, seed)
	rng := rand.New(rand.NewSource(seed))

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	const root = "unittest-stress"
	dirs := []string{root, root + "/a", root + "/b", root + "/a/c"}
	for _, dir := range dirs {
		if err := d.mkdir(ctx, dir); err != nil {
			if _, lserr := d.listFiles(ctx, dir); lserr != nil {
				t.Fatalf("setup mkdir(%s): %v (and does not exist: %v)", dir, err, lserr)
			}
		}
	}

	live := make([]stressFile, 0, 64)
	var liveBytes uint32
	seq := 0

	// history keeps the recent operation log so a failure can report what led
	// up to it rather than just the failing call.
	history := make([]string, 0, 32)
	record := func(format string, args ...interface{}) {
		s := fmt.Sprintf(format, args...)
		history = append(history, s)
		if len(history) > 24 {
			history = history[1:]
		}
	}
	// probeRecovery distinguishes the possible failure modes after a command
	// gets no answer within readExact's 9x1s budget:
	//   - a late response still arriving  => the device was merely slow
	//   - nothing, then INFO works        => the command itself was lost
	//   - nothing, and INFO keeps failing => the device is wedged
	probeRecovery := func() {
		t.Errorf("--- recovery probe ---")

		// is a late response still on its way?
		buf := make([]byte, 512)
		dctx, dcancel := context.WithTimeout(context.Background(), 30*time.Second)
		start := time.Now()
		n, derr := readExact(dctx, d.f, 512, buf)
		dcancel()
		if n > 0 {
			t.Errorf("  LATE DATA: %d bytes arrived %v after the timeout: %v",
				n, time.Since(start), derr)
			t.Errorf("  first 32 bytes: %s", hex.EncodeToString(buf[:min(int(n), 32)]))
		} else {
			t.Errorf("  no late data within %v (%v)", time.Since(start), derr)
		}

		// does the device answer a fresh command?
		for attempt := 1; attempt <= 3; attempt++ {
			ps := time.Now()
			_, _, _, ierr := d.info(context.Background())
			t.Errorf("  probe %d: info() -> %v [%v]", attempt, ierr, time.Since(ps))
			if ierr == nil {
				t.Errorf("  device answered again on probe %d", attempt)
				return
			}
		}
		t.Errorf("  device still unresponsive after 3 probes")
	}

	fail := func(format string, args ...interface{}) {
		t.Errorf("preceding operations (seed=%d):", seed)
		for _, h := range history {
			t.Errorf("  %s", h)
		}
		probeRecovery()
		t.Fatalf(format, args...)
	}

	rmFile := func(i int) {
		f := live[i]
		start := time.Now()
		err := d.rm(ctx, f.path)
		record("rm(%s) -> %v [%v]", f.path, err, time.Since(start))
		if err != nil {
			fail("rm(%s) failed: %v", f.path, err)
		}
		liveBytes -= f.size
		live = append(live[:i], live[i+1:]...)
	}

	doPut := func() {
		size := awkwardSizes[rng.Intn(len(awkwardSizes))]
		if rng.Intn(3) == 0 {
			// mix in wholly arbitrary sizes too:
			size = uint32(rng.Intn(200*1024)) + 1
		}
		for liveBytes+size > maxLive && len(live) > 0 {
			rmFile(rng.Intn(len(live)))
		}
		seq++
		f := stressFile{
			path: fmt.Sprintf("%s/f%04d.bin", dirs[rng.Intn(len(dirs))], seq),
			tag:  rng.Uint32(),
			size: size,
		}
		start := time.Now()
		n, err := d.putFile(ctx, f.path, f.size, bytes.NewReader(filePattern(f.tag, f.size)), nil)
		record("putFile(%s, size=%d, tag=$%08x) -> n=%d, %v [%v]",
			f.path, f.size, f.tag, n, err, time.Since(start))
		if err != nil {
			fail("putFile(%s, %d) failed: %v", f.path, f.size, err)
		}
		live = append(live, f)
		liveBytes += f.size
	}

	doGet := func() {
		if len(live) == 0 {
			doPut()
			return
		}
		f := live[rng.Intn(len(live))]
		var w bytes.Buffer
		w.Grow(int(f.size))
		start := time.Now()
		received, err := d.getFile(ctx, f.path, &w, nil, nil)
		record("getFile(%s, size=%d, tag=$%08x) -> received=%d, %v [%v]",
			f.path, f.size, f.tag, received, err, time.Since(start))
		if err != nil {
			fail("getFile(%s) failed: %v", f.path, err)
		}
		if received != f.size {
			fail("getFile(%s) received %d bytes, want %d", f.path, received, f.size)
		}
		expected := filePattern(f.tag, f.size)
		if !bytes.Equal(w.Bytes(), expected) {
			fail("getFile(%s) content mismatch: %s",
				f.path, describeStressMismatch(w.Bytes(), expected, f.tag))
		}
	}

	doLs := func() {
		dir := dirs[rng.Intn(len(dirs))]
		start := time.Now()
		files, err := d.listFiles(ctx, dir)
		record("ls(%s) -> %d entries, %v [%v]", dir, len(files), err, time.Since(start))
		if err != nil {
			fail("ls(%s) failed: %v", dir, err)
		}
	}

	doMkdir := func() {
		dir := fmt.Sprintf("%s/d%04d", dirs[rng.Intn(len(dirs))], rng.Intn(8))
		start := time.Now()
		err := d.mkdir(ctx, dir)
		record("mkdir(%s) -> %v [%v]", dir, err, time.Since(start))
		// error code 1 just means it already exists; only a transport failure
		// matters here, and that surfaces as a read timeout instead.
	}

	doInfo := func() {
		start := time.Now()
		_, _, _, err := d.info(ctx)
		record("info() -> %v [%v]", err, time.Since(start))
		if err != nil {
			fail("info() failed: %v", err)
		}
	}

	// doMemRead issues a VGET. Memory reads use a 64-byte command block with
	// FlagDATA64B|FlagNORESP, whereas every filesystem command uses 512 bytes.
	// The firmware re-evaluates server_info.cmd_size only when
	// recv_buffer_offset crosses 64 from below, so interleaving the two sizes
	// is where a framing desync between commands would show up.
	doMemRead := func() {
		// WRAM and SRAM, sizes chosen to span the 64-byte packet boundary:
		addrs := []uint32{0xF50010, 0xF50100, 0xE00000, 0xF5F340}
		sizes := []int{1, 2, 63, 64, 65, 100}
		req := devices.MemoryReadRequest{
			RequestAddress: devices.AddressTuple{
				Address:       addrs[rng.Intn(len(addrs))],
				AddressSpace:  sni.AddressSpace_FxPakPro,
				MemoryMapping: sni.MemoryMapping_LoROM,
			},
			Size: sizes[rng.Intn(len(sizes))],
		}
		start := time.Now()
		rsp, err := d.MultiReadMemory(ctx, req)
		record("MultiReadMemory(addr=$%06x, size=%d) -> %d rsp, %v [%v]",
			req.RequestAddress.Address, req.Size, len(rsp), err, time.Since(start))
		if err != nil {
			fail("MultiReadMemory(addr=$%06x, size=%d) failed: %v",
				req.RequestAddress.Address, req.Size, err)
		}
		if len(rsp) == 1 && len(rsp[0].Data) != req.Size {
			fail("MultiReadMemory(addr=$%06x) returned %d bytes, want %d",
				req.RequestAddress.Address, len(rsp[0].Data), req.Size)
		}
	}

	doRm := func() {
		if len(live) == 0 {
			doPut()
			return
		}
		rmFile(rng.Intn(len(live)))
	}

	start := time.Now()
	for i := 0; i < ops; i++ {
		switch n := rng.Intn(100); {
		case n < 26:
			doPut()
		case n < 52:
			doGet()
		case n < 72:
			doLs()
		case n < 82:
			doMkdir()
		case n < 88:
			doRm()
		case n < 96:
			doMemRead()
		default:
			doInfo()
		}
		if (i+1)%25 == 0 {
			t.Logf("%d/%d ops, %d live files, %d live bytes, %v elapsed",
				i+1, ops, len(live), liveBytes, time.Since(start))
		}
	}
	t.Logf("completed %d ops in %v", ops, time.Since(start))

	// leave the card as we found it:
	for len(live) > 0 {
		rmFile(len(live) - 1)
	}
}
