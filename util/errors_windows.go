package util

import (
	"errors"
	"syscall"
)

func IsConnectionRefused(err error) bool {
	var serr syscall.Errno
	if errors.As(err, &serr) {
		if serr == syscall.ECONNREFUSED {
			return true
		}
		// `connectex: No connection could be made because the target machine actively refused it.`
		if serr == 0x274d {
			return true
		}
	}

	return false
}
