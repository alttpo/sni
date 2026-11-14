package fxpakpro

import (
	"context"
	"os"
	"strings"
	"testing"
)

// TestCleanupDirs removes everything under the directories named in
// SNI_TEST_CLEANUP_DIRS (comma separated), then the directories themselves.
// Aborted test runs leave files behind, and this card has very little free
// space, so they need clearing between runs.
func TestCleanupDirs(t *testing.T) {
	list := os.Getenv("SNI_TEST_CLEANUP_DIRS")
	if list == "" {
		t.Skip("set SNI_TEST_CLEANUP_DIRS to a comma separated list of directories")
	}

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	var rmDir func(dir string, depth int)
	rmDir = func(dir string, depth int) {
		files, err := d.listFiles(ctx, dir)
		if err != nil {
			t.Logf("ls(%s): %v", dir, err)
			return
		}
		for _, f := range files {
			if f.Name == "." || f.Name == ".." {
				continue
			}
			child := dir + "/" + f.Name
			if f.Type == 0 && depth < 4 {
				rmDir(child, depth+1)
				continue
			}
			if err := d.rm(ctx, child); err != nil {
				t.Logf("rm(%s): %v", child, err)
			}
		}
		if err := d.rm(ctx, dir); err != nil {
			t.Logf("rm(%s): %v", dir, err)
		} else {
			t.Logf("removed %s", dir)
		}
	}

	for _, dir := range strings.Split(list, ",") {
		dir = strings.TrimSpace(dir)
		if dir == "" {
			continue
		}
		rmDir(dir, 0)
	}

	// show what is left at the root
	files, err := d.listFiles(ctx, "")
	if err != nil {
		t.Logf("ls(root): %v", err)
		return
	}
	for _, f := range files {
		if f.Name == "." || f.Name == ".." {
			continue
		}
		t.Logf("root: %q (type=%v)", f.Name, f.Type)
	}
}
