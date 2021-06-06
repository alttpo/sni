package snes

type KeepAlive chan<- struct{}

type Command interface {
	// Execute takes exclusive control of the queue device; keepAlive must be sent to periodically
	Execute(queue Queue, keepAlive KeepAlive) error
}

type Completion func(Command, error)

type CommandWithCompletion struct {
	Command    Command
	Completion Completion
}

type CommandSequence []CommandWithCompletion

func (seq CommandSequence) EnqueueTo(queue Queue) (err error) {
	for _, cmd := range seq {
		err = queue.Enqueue(cmd)
		if err != nil {
			return
		}
	}
	return
}

type NoOpCommand struct{}

func (c *NoOpCommand) Execute(queue Queue, keepAlive KeepAlive) error {
	return nil
}

// Special Command to close the device connection
type CloseCommand struct{}

func (c *CloseCommand) Execute(queue Queue, keepAlive KeepAlive) error {
	return nil
}
