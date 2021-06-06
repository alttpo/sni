package fxpakpro

import (
	"golang.org/x/sys/windows"
	"log"
	"syscall"
)

func (q *Queue) IsTerminalError(err error) bool {
	if err == nil {
		return false
	}

	if sysErr, ok := err.(syscall.Errno); ok {
		// temporary errors don't count:
		if sysErr.Temporary() {
			return false
		}
		// "device not configured" on mac:
		if sysErr == syscall.ENXIO {
			return true
		}
		// "The device does not recognize the command." on windows:
		if sysErr == windows.ERROR_BAD_COMMAND {
			return true
		}
		log.Printf("fxpakpro: isTerminalError(%d) = true; %v\n", uintptr(sysErr), sysErr)
		return true
	}

	return false
}
