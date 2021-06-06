package snes

import "fmt"

type ErrDeviceDisconnected struct {
	wrapped error
}

func (e ErrDeviceDisconnected) Unwrap() error { return e.wrapped }
func (e ErrDeviceDisconnected) Error() string {
	return fmt.Sprintf("snes device disconnected: %v", e.wrapped)
}
