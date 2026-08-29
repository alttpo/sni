package fxpakpro

import (
	"bytes"
	"context"
	"strings"
	"testing"
)

// TestDevice_usbstat reads /sd2snes/usbstat.txt, written by the instrumented
// firmware when CDC_BulkOut drops an already-ACKed packet because the server
// was busy. The file only exists once a drop has happened: usbstat_dirty is set
// solely on the drop path, so absence means no drop has ever been recorded.
func TestDevice_usbstat(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	files, err := d.listFiles(ctx, "sd2snes")
	if err != nil {
		t.Fatalf("ls(sd2snes): %v", err)
	}
	var found bool
	for _, f := range files {
		if f.Name == "usbstat.txt" {
			found = true
		}
	}
	if !found {
		t.Logf("usbstat.txt does not exist: no packet drop has been recorded")
		return
	}

	var w bytes.Buffer
	n, err := d.getFile(ctx, "sd2snes/usbstat.txt", &w, nil, nil)
	if err != nil {
		t.Fatalf("getFile(usbstat.txt): %v", err)
	}
	t.Logf("usbstat.txt is %d bytes:", n)
	for _, line := range strings.Split(strings.TrimRight(w.String(), "\x00\n"), "\n") {
		if line != "" {
			t.Logf("  %s", line)
		}
	}
}
