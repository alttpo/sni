//go:build windows
// +build windows

package tray

import (
	"fmt"
	"os"
	"syscall"
)

var (
	dllKernel32      *syscall.LazyDLL
	dllUser32        *syscall.LazyDLL
	procAllocConsole *syscall.LazyProc
	procGetWin       *syscall.LazyProc
	procShowWin      *syscall.LazyProc
)

func initConsole() (err error) {
	dllKernel32 = syscall.NewLazyDLL("kernel32.dll")
	dllUser32 = syscall.NewLazyDLL("user32.dll")

	procAllocConsole = dllKernel32.NewProc("AllocConsole")
	procGetWin = dllKernel32.NewProc("GetConsoleWindow")
	procShowWin = dllUser32.NewProc("ShowWindow")

	var r0 uintptr
	r0, _, err = syscall.Syscall(procAllocConsole.Addr(), 0, 0, 0, 0)
	if r0 == 0 {
		err = fmt.Errorf("AllocConsole(): %w", err)
		return
	}

	var hin, hout, herr syscall.Handle
	hin, err = syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
	if err != nil {
		err = fmt.Errorf("GetStdHandle(stdin): %w", err)
		return
	}
	hout, err = syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	if err != nil {
		err = fmt.Errorf("GetStdHandle(stdout): %w", err)
		return
	}
	herr, err = syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
	if err != nil {
		err = fmt.Errorf("GetStdHandle(stderr): %w", err)
		return
	}

	newStdin := os.NewFile(uintptr(hin), "/dev/stdin")
	newStdout := os.NewFile(uintptr(hout), "/dev/stdout")
	newStderr := os.NewFile(uintptr(herr), "/dev/stderr")

	//// Set handles for standard input, output and error devices.
	//err = windows.SetStdHandle(windows.STD_INPUT_HANDLE, windows.Handle(newStdin.Fd()))
	//if err != nil {
	//	return fmt.Errorf("failed to set standard input handler: %v", err)
	//}
	//err = windows.SetStdHandle(windows.STD_OUTPUT_HANDLE, windows.Handle(newStdout.Fd()))
	//if err != nil {
	//	return fmt.Errorf("failed to set standard output handler: %v", err)
	//}
	//err = windows.SetStdHandle(windows.STD_ERROR_HANDLE, windows.Handle(newStderr.Fd()))
	//if err != nil {
	//	return fmt.Errorf("failed to set standard error handler: %v", err)
	//}

	os.Stdin = newStdin
	os.Stdout = newStdout
	os.Stderr = newStderr

	err = consoleVisible(false)
	return
}

func consoleVisible(show bool) (err error) {
	hwnd, _, _ := procGetWin.Call()
	if hwnd == 0 {
		return
	}

	var r1, r2 uintptr
	if show {
		var SW_RESTORE uintptr = 9
		r1, r2, err = procShowWin.Call(hwnd, SW_RESTORE)
		//log.Printf("ShowWindow(SW_RESTORE) -> (%v, %v, %v)\n", r1, r2, err)
		_, _, _ = r1, r2, err
	} else {
		var SW_HIDE uintptr = 0
		r1, r2, err = procShowWin.Call(hwnd, SW_HIDE)
		//log.Printf("ShowWindow(SW_HIDE) -> (%v, %v, %v)\n", r1, r2, err)
		_, _, _ = r1, r2, err
	}

	err = nil
	return
}

func consoleIsDynamic() bool {
	return true
}
