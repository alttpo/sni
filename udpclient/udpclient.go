package udpclient

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sync"
	"time"
)

type UDPClient struct {
	name string

	c    *net.UDPConn
	addr *net.UDPAddr

	muteLog bool

	isConnected bool
	read        chan []byte
	write       chan []byte

	seqLock sync.Mutex
}

func NewUDPClient(name string) *UDPClient {
	return &UDPClient{
		name:  name,
		read:  make(chan []byte, 64),
		write: make(chan []byte, 64),
	}
}

func MakeUDPClient(name string, c *UDPClient) *UDPClient {
	c.name = name
	c.read = make(chan []byte, 64)
	c.write = make(chan []byte, 64)
	return c
}

func (c *UDPClient) MuteLog(muted bool) {
	c.muteLog = muted
}

func (c *UDPClient) Address() *net.UDPAddr { return c.addr }

func (c *UDPClient) Write() chan<- []byte { return c.write }
func (c *UDPClient) Read() <-chan []byte  { return c.read }

var ErrTimeout = fmt.Errorf("timeout")

func (c *UDPClient) WriteTimeout(m []byte, d time.Duration) error {
	timer := time.NewTimer(d)

	select {
	case c.write <- m:
		timer.Stop()
		return nil
	case <-timer.C:
		timer.Stop()
		return fmt.Errorf("%s: writeTimeout: %w\n", c.name, ErrTimeout)
	}
}

func (c *UDPClient) ReadTimeout(d time.Duration) ([]byte, error) {
	timer := time.NewTimer(d)

	select {
	case m := <-c.read:
		timer.Stop()
		return m, nil
	case <-timer.C:
		timer.Stop()
		return nil, fmt.Errorf("%s: readTimeout: %w", c.name, ErrTimeout)
	}
}

func (c *UDPClient) WriteThenReadTimeout(m []byte, d time.Duration) (rsp []byte, err error) {
	// hold a lock so we're guaranteed write->read consistency:
	defer c.seqLock.Unlock()
	c.seqLock.Lock()

	err = c.WriteTimeout(m, d)
	if err != nil {
		return
	}
	return c.ReadTimeout(d)
}

func (c *UDPClient) Lock() {
	//fmt.Printf("%s lock\n", c.name)
	c.seqLock.Lock()
}
func (c *UDPClient) Unlock() {
	//fmt.Printf("%s unlock\n", c.name)
	c.seqLock.Unlock()
}

func (c *UDPClient) SetReadDeadline(t time.Time) error  { return c.c.SetReadDeadline(t) }
func (c *UDPClient) SetWriteDeadline(t time.Time) error { return c.c.SetWriteDeadline(t) }

func (c *UDPClient) IsConnected() bool { return c.isConnected }

func (c *UDPClient) log(fmt string, args ...interface{}) {
	if c.muteLog {
		return
	}
	log.Printf(fmt, args...)
}

func (c *UDPClient) Connect(addr *net.UDPAddr) (err error) {
	c.log("%s: connect to server '%s'\n", c.name, addr)

	if c.isConnected {
		return fmt.Errorf("%s: already connected", c.name)
	}

	c.addr = addr

	c.c, err = net.DialUDP("udp", nil, addr)
	if err != nil {
		return
	}

	c.isConnected = true
	c.log("%s: connected to server '%s'\n", c.name, addr)

	go c.readLoop()
	go c.writeLoop()

	return
}

func (c *UDPClient) Disconnect() {
	c.log("%s: disconnect from server '%s'\n", c.name, c.addr)

	if !c.isConnected {
		return
	}

	c.isConnected = false
	err := c.c.SetReadDeadline(time.Now())
	if err != nil {
		c.log("%s: setreaddeadline: %v\n", c.name, err)
	}

	err = c.c.SetWriteDeadline(time.Now())
	if err != nil {
		c.log("%s: setwritedeadline: %v\n", c.name, err)
	}

	// signal a disconnect took place:
	c.read <- nil
	c.write <- nil

	// empty the write channel:
	for more := true; more; {
		select {
		case <-c.write:
		default:
			more = false
		}
	}

	// close the underlying connection:
	err = c.c.Close()
	if err != nil {
		c.log("%s: close: %v\n", c.name, err)
	}

	c.log("%s: disconnected from server '%s'\n", c.name, c.addr)

	c.c = nil
}

func (c *UDPClient) Close() {
	if c.read != nil {
		close(c.read)
	}
	if c.write != nil {
		close(c.write)
	}
	c.read = nil
	c.write = nil
}

// must run in a goroutine
func (c *UDPClient) readLoop() {
	c.log("%s: readLoop started\n", c.name)

	defer func() {
		c.Disconnect()
		c.log("%s: disconnected; readLoop exited\n", c.name)
	}()

	// we only need a single receive buffer:
	b := make([]byte, 1500)

	for c.isConnected {
		// wait for a packet from UDP socket:
		var n, _, err = c.c.ReadFromUDP(b)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				c.log("%s: read: %s\n", c.name, err)
			}
			return
		}

		// copy the envelope:
		envelope := make([]byte, n)
		copy(envelope, b[:n])

		c.read <- envelope
	}
}

// must run in a goroutine
func (c *UDPClient) writeLoop() {
	c.log("%s: writeLoop started\n", c.name)

	defer func() {
		c.Disconnect()
		c.log("%s: disconnected; writeLoop exited\n", c.name)
	}()

	for w := range c.write {
		if w == nil {
			return
		}

		// wait for a packet from UDP socket:
		var _, err = c.c.Write(w)
		if err != nil {
			if !errors.Is(err, net.ErrClosed) {
				c.log("%s: write: %s\n", c.name, err)
			}
			return
		}
	}
}
