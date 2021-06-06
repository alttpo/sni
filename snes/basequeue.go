package snes

import (
	"fmt"
	"log"
	"time"
)

const chanSize = 8

type BaseQueue struct {
	// driver name
	name string

	// command execution queue:
	cq       chan CommandWithCompletion
	cqClosed bool

	// derived Queue struct:
	queue Queue
}

func (b *BaseQueue) BaseInit(name string, queue Queue) {
	if queue == nil {
		panic("queue must not be nil")
	}

	b.name = name
	b.cq = make(chan CommandWithCompletion, chanSize)
	b.queue = queue

	go b.handleQueue()
}

func (b *BaseQueue) Enqueue(cmd CommandWithCompletion) (err error) {
	// FIXME: no great way I can figure out how to avoid panic on closed channel send below.
	defer func() {
		if recover() != nil {
			err = fmt.Errorf("%s: device connection is closed", b.name)
		}
	}()

	if b.cqClosed {
		err = fmt.Errorf("%s: device connection is closed", b.name)
		return
	}

	// don't need a timeout here since the queue should always guarantee process forward with its own timeouts
	b.cq <- cmd

	return
}

func (b *BaseQueue) handleQueue() {
	q := b.queue

	b.cqClosed = false
	var err error
	doClose := func() {
		if b.cqClosed {
			log.Printf("%s: already closed\n", b.name)
			return
		}

		if err != nil {
			log.Printf("%s: %v\n", b.name, err)
		}

		log.Printf("%s: calling Close()\n", b.name)
		if q != nil {
			err = q.Close()
			if err != nil {
				log.Printf("%s: %v\n", b.name, err)
			}
		}

		log.Printf("%s: closing chan\n", b.name)
		b.cqClosed = true
		close(b.cq)
		log.Printf("%s: closed chan\n", b.name)
	}
	defer doClose()

channelLoop:
	for pair := range b.cq {
		//log.Printf("%s: dequeue command; len = %d\n", b.name, len(b.cq))
		cmd := pair.Command

		if cmd == nil {
			break
		}

		terminal := false

		if _, ok := cmd.(*CloseCommand); ok {
			log.Printf("%s: processing CloseCommand\n", b.name)
			terminal = true
		}

		{
			// give the command 15 seconds to execute:
			const timeoutDuration = time.Second * 15

			done := make(chan struct{})
			keepAlive := make(chan struct{}, 16)
			started := time.Now()
			go func() {
				err = cmd.Execute(q, keepAlive)
				close(done)
			}()

			timeout := time.NewTimer(timeoutDuration)
		timeoutLoop:
			for {
				select {
				case <-done:
					timeout.Stop()
					break timeoutLoop
				case <-keepAlive:
					timeout.Reset(timeoutDuration)
					break
				case <-timeout.C:
					timeout.Stop()
					log.Printf("%s: timed out executing command\n", b.name)
					break channelLoop
				}
			}
			stopped := time.Now()
			executionTime := stopped.Sub(started).Milliseconds()
			_ = executionTime
			//log.Printf("%s: command execution took %d msec", b.name, executionTime)
		}

		// wrap the error if it is a terminal case:
		if err != nil && q.IsTerminalError(err) {
			err = ErrDeviceDisconnected{err}
			terminal = true
		}
		if pair.Completion != nil {
			go pair.Completion(cmd, err)
		} else if err != nil {
			log.Println(err)
		}

		if terminal {
			break
		}
	}
}

func (b *BaseQueue) MakeReadCommands(reqs []Read, complete func(error)) CommandSequence {
	panic("implement me")
}

func (b *BaseQueue) MakeWriteCommands(reqs []Write, complete func(error)) CommandSequence {
	panic("implement me")
}
