package retroarch

import (
	"fmt"
	"sni/snes"
	"sync"
)

type Queue struct {
	snes.BaseQueue

	closed chan struct{}

	c    *RAClient
	lock sync.Mutex
}

var (
	ErrClosed = fmt.Errorf("connection is closed")
)

func (q *Queue) IsTerminalError(err error) bool {
	//if errors.Is(err, udpclient.ErrTimeout) {
	//	return true
	//}
	//if errors.Is(err, ErrClosed) {
	//	return true
	//}
	return false
}

func (q *Queue) IsClosed() bool {
	return q.c == nil
}


func (q *Queue) Closed() <-chan struct{} {
	return q.closed
}

func (q *Queue) Close() error {
	defer q.lock.Unlock()
	q.lock.Lock()

	if q.c == nil {
		return nil
	}

	q.c.Close()
	q.c = nil
	close(q.closed)

	return nil
}

func (q *Queue) Init() {
	q.closed = make(chan struct{})
}

func (q *Queue) MakeReadCommands(reqs []snes.Read, batchComplete snes.Completion) (cmds snes.CommandSequence) {
	cmds = make(snes.CommandSequence, 0, len(reqs)/8+1)

	for len(reqs) >= 8 {
		// queue up a batch read command:
		batch := reqs[:8]
		cmds = append(cmds, snes.CommandWithCompletion{
			Command:    &readCommand{batch},
			Completion: batchComplete,
		})

		// move to next batch:
		reqs = reqs[8:]
	}

	if len(reqs) > 0 && len(reqs) <= 8 {
		cmds = append(cmds, snes.CommandWithCompletion{
			Command:    &readCommand{reqs},
			Completion: batchComplete,
		})
	}

	return cmds
}

func (q *Queue) MakeWriteCommands(reqs []snes.Write, batchComplete snes.Completion) (cmds snes.CommandSequence) {
	cmds = make(snes.CommandSequence, 0, len(reqs)/8+1)

	for len(reqs) >= 8 {
		// queue up a batch read command:
		batch := reqs[:8]
		cmds = append(cmds, snes.CommandWithCompletion{
			Command:    &writeCommand{batch},
			Completion: batchComplete,
		})

		// move to next batch:
		reqs = reqs[8:]
	}

	if len(reqs) > 0 && len(reqs) <= 8 {
		cmds = append(cmds, snes.CommandWithCompletion{
			Command:    &writeCommand{reqs},
			Completion: batchComplete,
		})
	}

	return cmds
}

type readCommand struct {
	Batch []snes.Read
}

func (cmd *readCommand) Execute(queue snes.Queue, keepAlive snes.KeepAlive) (err error) {
	q, ok := queue.(*Queue)
	if !ok {
		return fmt.Errorf("queue is not of expected internal type")
	}

	q.lock.Lock()
	c := q.c
	q.lock.Unlock()
	if c == nil {
		return fmt.Errorf("retroarch: read: %w", ErrClosed)
	}
	keepAlive <- struct{}{}

	err = c.ReadMemoryBatch(cmd.Batch, keepAlive)
	//if err != nil {
	//	_ = q.Close()
	//}

	return
}

type writeCommand struct {
	Batch []snes.Write
}

const hextable = "0123456789abcdef"

func (cmd *writeCommand) Execute(queue snes.Queue, keepAlive snes.KeepAlive) (err error) {
	q, ok := queue.(*Queue)
	if !ok {
		return fmt.Errorf("queue is not of expected internal type")
	}

	q.lock.Lock()
	c := q.c
	q.lock.Unlock()
	if c == nil {
		return fmt.Errorf("retroarch: write: %w", ErrClosed)
	}
	keepAlive <- struct{}{}

	err = c.WriteMemoryBatch(cmd.Batch, keepAlive)
	//if err != nil {
	//	_ = q.Close()
	//}

	return
}
