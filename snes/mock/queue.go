package mock

import (
	"fmt"
	"sni/snes"
	"time"
)

type Queue struct {
	snes.BaseQueue

	closed chan struct{}

	WRAM    [0x20000]byte
	nothing [0x100]byte

	frameTicker *time.Ticker
}

func (q *Queue) IsTerminalError(err error) bool {
	return false
}

func (q *Queue) Closed() <-chan struct{} {
	return q.closed
}

func (q *Queue) Close() error {
	q.frameTicker.Stop()
	q.frameTicker = nil
	close(q.closed)
	return nil
}

func (q *Queue) Init() {
	q.closed = make(chan struct{})
	q.frameTicker = time.NewTicker(16_639_265 * time.Nanosecond)
	go func() {
		// 5,369,317.5/89,341.5 ~= 60.0988 frames / sec ~= 16,639,265.605 ns / frame
		for range q.frameTicker.C {
			// increment frame timer:
			q.WRAM[0x1A]++
		}
	}()
}

func (q *Queue) MakeReadCommands(reqs []snes.Read, batchComplete snes.Completion) snes.CommandSequence {
	seq := make(snes.CommandSequence, 0, len(reqs))
	for _, req := range reqs {
		seq = append(seq, snes.CommandWithCompletion{
			Command:    &readCommand{req},
			Completion: batchComplete,
		})
	}
	return seq
}

func (q *Queue) MakeWriteCommands(reqs []snes.Write, batchComplete snes.Completion) snes.CommandSequence {
	seq := make(snes.CommandSequence, 0, len(reqs))
	for _, req := range reqs {
		seq = append(seq, snes.CommandWithCompletion{
			Command:    &writeCommand{req},
			Completion: batchComplete,
		})
	}
	return seq
}

type readCommand struct {
	Request snes.Read
}

func (r *readCommand) Execute(queue snes.Queue, keepAlive snes.KeepAlive) error {
	q, ok := queue.(*Queue)
	if !ok {
		return fmt.Errorf("queue is not of expected internal type")
	}

	// wait 1ms before returning response to simulate the delay of FX Pak Pro device:
	<-time.After(time.Millisecond * 1)

	completed := r.Request.Completion
	if completed == nil {
		return nil
	}

	var data []byte
	if r.Request.Address >= 0xF50000 && r.Request.Address < 0xF70000 {
		// read from wram:
		o := r.Request.Address - 0xF50000
		data = q.WRAM[o : o+uint32(r.Request.Size)]
	} else {
		// read from nothing:
		data = q.nothing[0:r.Request.Size]
	}

	completed(snes.Response{
		IsWrite: false,
		Address: r.Request.Address,
		Size:    r.Request.Size,
		Extra:   r.Request.Extra,
		Data:    data,
	})

	return nil
}

type writeCommand struct {
	Request snes.Write
}

func (r *writeCommand) Execute(_ snes.Queue, keepAlive snes.KeepAlive) error {
	<-time.After(time.Millisecond * 1)

	completed := r.Request.Completion
	if completed != nil {
		completed(snes.Response{
			IsWrite: true,
			Address: r.Request.Address,
			Size:    r.Request.Size,
			Extra:   r.Request.Extra,
			Data:    r.Request.Data,
		})
	}
	return nil
}
