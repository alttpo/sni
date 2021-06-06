//+build !windows

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
	}

	return false
}
