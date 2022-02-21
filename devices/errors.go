package devices

import (
	"errors"
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

type DeviceError interface {
	error
	IsFatal() bool
}

func IsFatal(err error) bool {
	if e, ok := err.(DeviceError); ok {
		return e.IsFatal()
	}

	if unwrapped := errors.Unwrap(err); unwrapped != nil {
		return IsFatal(unwrapped)
	}

	return false
}

type DeviceErrorGenerator interface {
	FatalError(cause error) DeviceError
	NonFatalError(cause error) DeviceError
}

type Error struct {
	msg     string
	cause   error
	isFatal bool
}

func (e *Error) Error() string { return e.msg }
func (e *Error) Unwrap() error { return e.cause }
func (e *Error) IsFatal() bool { return e.isFatal }

func DeviceNonFatal(msg string, cause error) DeviceError { return &Error{msg, cause, false} }
func DeviceFatal(msg string, cause error) DeviceError    { return &Error{msg, cause, true} }
