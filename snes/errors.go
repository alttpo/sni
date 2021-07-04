package snes

import (
	"fmt"
	"google.golang.org/grpc/codes"
)

type ErrDeviceDisconnected struct {
	wrapped error
}

func (e ErrDeviceDisconnected) Unwrap() error { return e.wrapped }
func (e ErrDeviceDisconnected) Error() string {
	return fmt.Sprintf("snes device disconnected: %v", e.wrapped)
}

type CodedError struct {
	codes.Code
	Cause error
}

func (e *CodedError) Error() string { return e.Cause.Error() }
func (e *CodedError) Unwrap() error { return e.Cause }

func WithCode(code codes.Code, cause error) *CodedError { return &CodedError{code, cause} }
