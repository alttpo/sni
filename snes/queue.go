package snes

import (
	"io"
)

// Represents an asynchronous communication interface to either a physical or emulated SNES system.
// Communication with a physical SNES console is done via a flash cart with a USB connection.
// Both read and write requests are both enqueued into the same request queue and are processed in the order received.
// For reads, the read data is sent via the Completion callback specified in the Read struct.
// Depending on the implementation, reads and writes may be broken up into fixed-size batches.
// Read requests can read from ROM, SRAM, and WRAM. Flash carts can listen to the SNES address and data buses in order
// to shadow WRAM for reading.
// Write requests can only write to ROM and SRAM. WRAM cannot be written to from flash carts on real hardware; this is a
// hard limitation due to the design of the SNES and is not specific to any flash cart.
type Queue interface {
	io.Closer

	// This channel is closed when the underlying connection is closed
	Closed() <-chan struct{}

	// Enqueues a command with an optional completion callback
	Enqueue(cmd CommandWithCompletion) error

	// Creates a sequence of Commands which submit possibly multiple batches of read requests to the device
	MakeReadCommands(reqs []Read, batchComplete Completion) CommandSequence

	// Creates a sequence of Commands which submit possibly multiple batches of write requests to the device
	MakeWriteCommands(reqs []Write, batchComplete Completion) CommandSequence

	// IsTerminalError determines if the given error should cause the underlying device to be closed
	IsTerminalError(err error) bool
}
