package fxpakpro

import (
	"context"
	"os"
	"testing"
	"time"
)

// TestDevice_bootPath boots the ROM named by SNI_TEST_BOOT_PATH and reports
// what the device says it is running afterwards. The firmware polls USB very
// differently depending on whether it sits in the system menu or runs a ROM:
// menu_main_loop() sleeps 20ms per iteration before calling usbint_handler(),
// whereas the in-game loop in main.c calls it every iteration with no sleep.
// An MSU-1 ROM is different again -- it runs while(!msu1_loop()), which reaches
// usbint_handler() only when it has no command of its own and is streaming
// audio off the same SD card.
func TestDevice_bootPath(t *testing.T) {
	path := os.Getenv("SNI_TEST_BOOT_PATH")
	if path == "" {
		t.Skip("set SNI_TEST_BOOT_PATH to the ROM to boot")
	}

	d := openExactDevice(t)
	defer d.Close()
	ctx := context.Background()

	if _, _, rom, err := d.info(ctx); err != nil {
		t.Fatalf("info() before boot: %v", err)
	} else {
		t.Logf("before boot, running: %q", rom)
	}

	start := time.Now()
	if err := d.boot(ctx, path); err != nil {
		t.Fatalf("boot(%s): %v", path, err)
	}
	t.Logf("boot(%s) returned in %v", path, time.Since(start))

	// give the SNES a moment to actually start the ROM. Each probe gets its own
	// short deadline, with a pause between them, so a device that has stopped
	// answering fails quickly instead of spinning.
	for attempt := 1; attempt <= 10; attempt++ {
		time.Sleep(2 * time.Second)

		pctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		_, _, rom, err := d.info(pctx)
		cancel()
		if err != nil {
			t.Logf("probe %d: info() -> %v", attempt, err)
			continue
		}
		t.Logf("probe %d: running %q", attempt, rom)
		if rom != "" && rom != "/sd2snes/menu.bin" {
			t.Logf("boot confirmed after %v", time.Since(start))
			return
		}
	}
	t.Errorf("device never reported a booted ROM within 10 probes")
}

// TestDevice_info reports what the device is currently running. Useful as a
// liveness and state probe between tests, particularly to confirm whether the
// firmware is sitting in the system menu or executing a ROM, which changes how
// often the main loop calls usbint_handler().
func TestDevice_info(t *testing.T) {
	d := openExactDevice(t)
	defer d.Close()

	version, device, rom, err := d.info(context.Background())
	if err != nil {
		t.Fatalf("info(): %v", err)
	}
	t.Logf("version=%q device=%q rom=%q", version, device, rom)
	if rom == "/sd2snes/menu.bin" || rom == "/sd2snes/m3nu.bin" {
		t.Logf("state: SYSTEM MENU (menu_main_loop, sleep_ms(20) per usbint_handler call)")
	} else {
		t.Logf("state: IN-GAME (main.c loop, usbint_handler every iteration)")
	}
}
