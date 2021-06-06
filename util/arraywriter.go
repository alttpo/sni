package util

// ArrayWriter writes into a []byte slice without altering its length
type ArrayWriter struct {
	Buffer []byte
	Offset uint32
}

func (a *ArrayWriter) Write(p []byte) (n int, err error) {
	l := uint32(len(p))
	n = copy(a.Buffer[a.Offset:a.Offset+l], p)
	err = nil
	return
}
