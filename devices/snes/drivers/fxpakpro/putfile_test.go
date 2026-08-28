package fxpakpro

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"testing"
	"time"
)

type patternReader struct {
	size uint32
	offs uint32
}

func (r *patternReader) Read(d []byte) (n int, err error) {
	n = 0
	for i := range d {
		if r.offs >= r.size {
			err = io.EOF
			return
		}
		d[i] = byte(r.offs)

		r.offs++
		n++
		if n >= 63 {
			return
		}
	}

	return
}

func TestDevice_putFile(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	ctx := context.Background()
	type args struct {
		path string
		size uint32
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "511.sfc",
			args: args{
				path: "unittest/test1.sfc",
				size: 511,
			},
			wantErr: false,
		},
		{
			name: "513.sfc",
			args: args{
				path: "unittest/test2.sfc",
				size: 513,
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &patternReader{size: tt.args.size}
			n, err := d.putFile(ctx, tt.args.path, tt.args.size, r, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("putFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			_ = n

			// make sure a second command works immediately after transfer:
			_, _, _, err = d.info(ctx)
			if err != nil {
				t.Errorf("info() after putFile() failed: %v", err)
			}
		})
	}
}

// The tests below investigate an end-user report: uploading a large (4 MiB)
// file over the gRPC PutFile API fails under normal use but succeeds when
// SNI_DEBUG=1 is set. The working theory was that debug logging slows the host
// down enough to stop it from outrunning the device.

const largeTestDir = "unittest-large"
const largeTestSize = 4 * 1024 * 1024

// offsetPattern fills a buffer with a deterministic pattern where each 4-byte
// little-endian word holds its own byte offset within the file. If the stream
// desyncs, the value read back at a given offset reveals exactly how far the
// data shifted, which separates a dropped or duplicated run of bytes from
// bit-level corruption.
func offsetPattern(size uint32) []byte {
	d := make([]byte, size)
	for i := uint32(0); i+4 <= size; i += 4 {
		binary.LittleEndian.PutUint32(d[i:], i)
	}
	return d
}

// describeMismatch locates the first differing byte and reports the words on
// either side of it.
func describeMismatch(actual, expected []byte) string {
	for i := range expected {
		if i >= len(actual) {
			return fmt.Sprintf("contents truncated at offset %d ($%06x)", i, i)
		}
		if actual[i] == expected[i] {
			continue
		}
		lo := i &^ 3
		got := binary.LittleEndian.Uint32(actual[lo:])
		want := binary.LittleEndian.Uint32(expected[lo:])
		return fmt.Sprintf(
			"contents differ starting at offset %d ($%06x): got word $%08x, want $%08x (shifted by %d bytes)",
			i, i, got, want, int64(got)-int64(want),
		)
	}
	return "contents match"
}

// openLargeTestDir opens the device and ensures the scratch folder exists off
// the root of the SD card, cleaning up its contents when the test ends. The
// device is opened once and reused for every iteration: the fxpakpro firmware
// cannot tear down and re-establish a USB session, so reconnecting per
// iteration would wedge it before we reached the condition we are hunting for.
func openLargeTestDir(t *testing.T) (*Device, context.Context) {
	d := openExactDevice(t)
	// runs last, after the cleanup registered below:
	t.Cleanup(func() { d.Close() })

	ctx := context.Background()

	// mkdir fails with error code 1 if the folder already exists, so fall back
	// to listing it to tell "already there" apart from a real failure:
	if err := d.mkdir(ctx, largeTestDir); err != nil {
		if _, lserr := d.listFiles(ctx, largeTestDir); lserr != nil {
			t.Fatalf("mkdir(%s): %v (and it does not already exist: %v)", largeTestDir, err, lserr)
		}
	}

	t.Cleanup(func() {
		// best-effort; the device may be wedged by the time we get here:
		files, err := d.listFiles(context.Background(), largeTestDir)
		if err != nil {
			t.Logf("cleanup: ls(%s): %v", largeTestDir, err)
			return
		}
		for _, f := range files {
			if f.Name == "." || f.Name == ".." {
				continue
			}
			p := largeTestDir + "/" + f.Name
			if err := d.rm(context.Background(), p); err != nil {
				t.Logf("cleanup: rm(%s): %v", p, err)
			}
		}
		if err := d.rm(context.Background(), largeTestDir); err != nil {
			t.Logf("cleanup: rm(%s): %v", largeTestDir, err)
		}
	})

	return d, ctx
}

// verifyFile reads path back off the device and compares it to expected.
func verifyFile(t *testing.T, d *Device, ctx context.Context, path string, expected []byte) {
	t.Helper()

	// the device must still be responsive on the very next command:
	if _, _, _, err := d.info(ctx); err != nil {
		t.Fatalf("info() after putFile(): %v", err)
	}

	var w bytes.Buffer
	w.Grow(len(expected))
	received, err := d.getFile(ctx, path, &w, nil, nil)
	if err != nil {
		t.Fatalf("getFile(%s): %v", path, err)
	}
	if received != uint32(len(expected)) {
		t.Fatalf("getFile() received %d bytes, want %d", received, len(expected))
	}
	if actual := w.Bytes(); !bytes.Equal(actual, expected) {
		t.Fatalf("%s", describeMismatch(actual, expected))
	}
}

// TestDevice_putFile_large uploads a 4 MiB file several times over a single
// connection, verifying each upload byte-for-byte, to look for an intermittent
// transfer failure.
func TestDevice_putFile_large(t *testing.T) {
	const iterations = 5

	d, ctx := openLargeTestDir(t)
	expected := offsetPattern(largeTestSize)

	for i := 0; i < iterations; i++ {
		path := fmt.Sprintf("%s/large%d.bin", largeTestDir, i)

		// match what services/grpcimpl passes down from a gRPC PutFile: the
		// whole payload in memory as a bytes.Reader and no progress callback.
		start := time.Now()
		n, err := d.putFile(ctx, path, largeTestSize, bytes.NewReader(expected), nil)
		elapsed := time.Since(start)
		if err != nil {
			t.Fatalf("[%d] putFile(%s, %d): %v", i, path, largeTestSize, err)
		}
		if n != largeTestSize {
			t.Fatalf("[%d] putFile() sent %d bytes, want %d", i, n, largeTestSize)
		}
		t.Logf("[%d] putFile: %d bytes in %v (%.1f KiB/s)", i, n, elapsed, float64(n)/1024.0/elapsed.Seconds())

		verifyFile(t, d, ctx, path, expected)
	}
}

// putFileWriteSize performs the same PUT as Device.putFile but writes the
// payload in writeSize-byte calls instead of the 512-byte calls that
// sendSerialProgress uses. The bytes on the wire are identical; only the gap
// between successive write() syscalls changes. A large writeSize lets the host
// driver stream USB packets back-to-back with no user-space gap, which is the
// opposite of what SNI_DEBUG=1 does when it interposes a hex dump between every
// 512-byte chunk.
func putFileWriteSize(t *testing.T, d *Device, ctx context.Context, path string, payload []byte, writeSize int) {
	t.Helper()

	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)
	copy(sb[256:512], []byte(path))
	binary.BigEndian.PutUint32(sb[252:], uint32(len(payload)))

	d.lock.Lock()
	defer d.lock.Unlock()

	if err := sendSerialChunked(ctx, d.f, 512, sb); err != nil {
		t.Fatalf("send PUT command: %v", err)
	}
	if err := recvSerial(ctx, d.f, sb, 512); err != nil {
		t.Fatalf("recv PUT response: %v", err)
	}
	if sb[0] != 'U' || sb[1] != 'S' || sb[2] != 'B' || sb[3] != 'A' {
		t.Fatalf("PUT response missing USBA header: %x", sb[:8])
	}
	if ec := sb[5]; ec != 0 {
		t.Fatalf("PUT response error: %v", fxpakproError(ec))
	}

	for off := 0; off < len(payload); off += writeSize {
		end := off + writeSize
		if end > len(payload) {
			end = len(payload)
		}
		p := payload[off:end]
		for len(p) > 0 {
			n, err := d.f.Write(p)
			if err != nil {
				t.Fatalf("write at offset %d: %v", off, err)
			}
			p = p[n:]
		}
	}
}

// TestDevice_putFile_writeSizes uploads the same 4 MiB payload at host write
// sizes from the 512 bytes SNI uses today up to 64 KiB. If the reported failure
// were caused by the host outrunning the device, the larger write sizes are
// where it should first appear, and throughput should rise as the gaps between
// writes shrink. Identical throughput across all sizes instead means the
// transfer is device-bound: USB NAK flow control is pacing the host, and
// host-side pacing changes nothing on the wire.
func TestDevice_putFile_writeSizes(t *testing.T) {
	d, ctx := openLargeTestDir(t)
	expected := offsetPattern(largeTestSize)

	for _, writeSize := range []int{512, 4096, 16384, 65536} {
		t.Run(fmt.Sprintf("write%d", writeSize), func(t *testing.T) {
			path := fmt.Sprintf("%s/ws%d.bin", largeTestDir, writeSize)

			start := time.Now()
			putFileWriteSize(t, d, ctx, path, expected, writeSize)
			elapsed := time.Since(start)
			t.Logf("putFile(writeSize=%d): %d bytes in %v (%.1f KiB/s)",
				writeSize, largeTestSize, elapsed, float64(largeTestSize)/1024.0/elapsed.Seconds())

			verifyFile(t, d, ctx, path, expected)
		})
	}
}

// TestDevice_putFile_overwrite uploads a 4 MiB file to the same path several
// times. Every upload after the first forces the firmware to truncate the
// previous 4 MiB file inside f_open(FA_WRITE|FA_CREATE_ALWAYS) before it can
// reply, which is the slowest device-side step in a PUT and the one most likely
// to trip the host's response timeout. SNI reads that reply with readExact,
// which gives up after 9 consecutive one-second zero-byte reads, or at the
// context deadline if the caller set one.
func TestDevice_putFile_overwrite(t *testing.T) {
	const iterations = 4
	const path = largeTestDir + "/overwrite.bin"

	d, ctx := openLargeTestDir(t)
	expected := offsetPattern(largeTestSize)

	for i := 0; i < iterations; i++ {
		sb := make([]byte, 512)
		sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
		sb[4] = byte(OpPUT)
		sb[5] = byte(SpaceFILE)
		sb[6] = byte(FlagNONE)
		copy(sb[256:512], []byte(path))
		binary.BigEndian.PutUint32(sb[252:], largeTestSize)

		d.lock.Lock()
		if err := sendSerialChunked(ctx, d.f, 512, sb); err != nil {
			d.lock.Unlock()
			t.Fatalf("[%d] send PUT command: %v", i, err)
		}

		// time the command->response round trip on its own; this is the f_open:
		start := time.Now()
		err := recvSerial(ctx, d.f, sb, 512)
		latency := time.Since(start)
		if err != nil {
			d.lock.Unlock()
			t.Fatalf("[%d] recv PUT response after %v: %v", i, latency, err)
		}
		if ec := sb[5]; ec != 0 {
			d.lock.Unlock()
			t.Fatalf("[%d] PUT response error: %v", i, fxpakproError(ec))
		}

		sent, err := sendSerialProgress(ctx, d.f, 512, largeTestSize, bytes.NewReader(expected), nil)
		d.lock.Unlock()
		if err != nil {
			t.Fatalf("[%d] sendSerialProgress: %v", i, err)
		}
		if sent != largeTestSize {
			t.Fatalf("[%d] sent %d bytes, want %d", i, sent, largeTestSize)
		}
		t.Logf("[%d] f_open latency: %v", i, latency)

		verifyFile(t, d, ctx, path, expected)
	}
}

// chunkTiming records how long one 512-byte write to the serial port took.
type chunkTiming struct {
	offset uint32
	dur    time.Duration
}

// putFileInstrumented performs a PUT and records the wall time of every
// 512-byte write to the serial port, plus the command->response latency (the
// device-side f_open). Because the transfer is device-bound -- the host blocks
// in write() while the fxpakpro NAKs -- a long stall inside the firmware's USB
// interrupt handler shows up directly as a slow write() here. FatFs does its
// cluster allocation inside that ISR, so a nearly-full or fragmented card makes
// create_chain scan the FAT, and that scan is what we expect to see.
func putFileInstrumented(t *testing.T, d *Device, ctx context.Context, path string, payload []byte) {
	t.Helper()

	size := uint32(len(payload))

	sb := make([]byte, 512)
	sb[0], sb[1], sb[2], sb[3] = byte('U'), byte('S'), byte('B'), byte('A')
	sb[4] = byte(OpPUT)
	sb[5] = byte(SpaceFILE)
	sb[6] = byte(FlagNONE)
	copy(sb[256:512], []byte(path))
	binary.BigEndian.PutUint32(sb[252:], size)

	d.lock.Lock()
	defer d.lock.Unlock()

	if err := sendSerialChunked(ctx, d.f, 512, sb); err != nil {
		t.Fatalf("send PUT command: %v", err)
	}

	openStart := time.Now()
	if err := recvSerial(ctx, d.f, sb, 512); err != nil {
		t.Fatalf("recv PUT response after %v: %v", time.Since(openStart), err)
	}
	openLatency := time.Since(openStart)
	if ec := sb[5]; ec != 0 {
		t.Fatalf("PUT response error: %v", fxpakproError(ec))
	}

	timings := make([]chunkTiming, 0, size/512)
	start := time.Now()
	for off := uint32(0); off < size; off += 512 {
		end := off + 512
		if end > size {
			end = size
		}
		p := payload[off:end]

		wStart := time.Now()
		for len(p) > 0 {
			n, err := d.f.Write(p)
			if err != nil {
				t.Fatalf("write at offset %d (%d chunks in, %v elapsed): %v",
					off, off/512, time.Since(start), err)
			}
			p = p[n:]
		}
		timings = append(timings, chunkTiming{offset: off, dur: time.Since(wStart)})
	}
	total := time.Since(start)

	// summarize: sort a copy by duration to find the worst stalls.
	sorted := make([]chunkTiming, len(timings))
	copy(sorted, timings)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].dur > sorted[j].dur })

	pct := func(p float64) time.Duration {
		if len(sorted) == 0 {
			return 0
		}
		i := int(float64(len(sorted)-1) * (1.0 - p/100.0))
		return sorted[i].dur
	}

	t.Logf("f_open latency: %v", openLatency)
	t.Logf("data phase: %d chunks in %v (%.1f KiB/s)",
		len(timings), total, float64(size)/1024.0/total.Seconds())
	t.Logf("per-chunk write latency: p50=%v p99=%v p99.9=%v max=%v",
		pct(50), pct(99), pct(99.9), sorted[0].dur)

	t.Logf("10 slowest chunks:")
	for i := 0; i < 10 && i < len(sorted); i++ {
		t.Logf("  offset %8d ($%06x, chunk %5d): %v",
			sorted[i].offset, sorted[i].offset, sorted[i].offset/512, sorted[i].dur)
	}

	// count how much total time went into stalls above a threshold:
	for _, thresh := range []time.Duration{10 * time.Millisecond, 50 * time.Millisecond, 200 * time.Millisecond} {
		var count int
		var sum time.Duration
		for _, ct := range timings {
			if ct.dur >= thresh {
				count++
				sum += ct.dur
			}
		}
		t.Logf("chunks >= %v: %d (%v total)", thresh, count, sum)
	}
}

// TestDevice_putFile_intoDir uploads a 4 MiB file into an existing folder on
// the SD card, chosen with SNI_TEST_PUTFILE_DIR, and reports where the device
// stalled. Point it at a folder holding many large files (an MSU-1 track set,
// say) to test the theory that FAT work inside the firmware's USB interrupt
// handler is what loses packets.
func TestDevice_putFile_intoDir(t *testing.T) {
	dir := os.Getenv("SNI_TEST_PUTFILE_DIR")
	if dir == "" {
		t.Skip("set SNI_TEST_PUTFILE_DIR to the folder to upload into")
	}

	// SNI_TEST_PUTFILE_SIZE overrides the payload size in bytes. Sizing the
	// upload close to the volume's remaining free space is what forces the
	// firmware's allocator to scan for each hole in a fragmented FAT.
	size := uint32(largeTestSize)
	if s := os.Getenv("SNI_TEST_PUTFILE_SIZE"); s != "" {
		v, err := strconv.ParseUint(s, 0, 32)
		if err != nil {
			t.Fatalf("SNI_TEST_PUTFILE_SIZE=%q: %v", s, err)
		}
		size = uint32(v)
	}

	d := openExactDevice(t)
	// registered first so it runs last, after the rm cleanup below:
	t.Cleanup(func() { d.Close() })
	ctx := context.Background()

	files, err := d.listFiles(ctx, dir)
	if err != nil {
		t.Fatalf("ls(%s): %v", dir, err)
	}
	t.Logf("target folder %q holds %d directory entries", dir, len(files))

	path := dir + "/sni-putfile-test.bin"
	t.Cleanup(func() {
		if err := d.rm(context.Background(), path); err != nil {
			t.Logf("cleanup: rm(%s): %v", path, err)
		}
	})

	t.Logf("uploading %d bytes (%.1f MiB)", size, float64(size)/1024.0/1024.0)
	expected := offsetPattern(size)
	putFileInstrumented(t, d, ctx, path, expected)
	verifyFile(t, d, ctx, path, expected)
}
