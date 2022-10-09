package util

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"runtime"
)

var MaxStackDepth = 50

type StackTrace struct {
	stack []uintptr
}

func NewStackTrace(skip int) *StackTrace {
	stack := make([]uintptr, MaxStackDepth)
	length := runtime.Callers(skip, stack[:])
	return &StackTrace{
		stack: stack[:length],
	}
}

func (s *StackTrace) String() string {
	buf := bytes.Buffer{}

	callers := runtime.CallersFrames(s.stack)

	for frame, more := callers.Next(); more; frame, more = callers.Next() {
		if frame.Func == nil {
			// Ignore fully inlined functions
			continue
		}

		name := frame.Func.Name()
		str := fmt.Sprintf("\t%s: %s:%d (0x%x)\n", name, frame.File, frame.Line, frame.PC)

		buf.WriteString(str)
	}

	return buf.String()
}

func Recover() {
	r := recover()
	if r == nil {
		// no panic
		return
	}

	s := NewStackTrace(3)
	log.Printf("panic: %v\n%s", r, s.String())
	os.Exit(254)
}
