//+build !windows

package fxpakpro

import (
	"log"
	"syscall"
)

func (q *Queue) IsTerminalError(err error) bool {
	if err == nil {
		return false
	}

	if serr, ok := err.(syscall.Errno); ok {
		// temporary errors don't count:
		if serr.Temporary() {
			return false
		}
		// "device not configured" on mac:
		if serr == syscall.ENXIO {
			return true
		}
		log.Printf("fxpakpro: isTerminalError(%d) = true; %v\n", uintptr(serr), serr)
		return true
	}

	return false
}
