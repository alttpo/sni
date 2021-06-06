package qusb2snes

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"syscall"
	"time"
)

type WebSocketClient struct {
	urlstr  string
	appName string

	lock    sync.Mutex
	closed  chan struct{}
	ws      net.Conn
	r       *wsutil.Reader
	w       *wsutil.Writer
	encoder *json.Encoder
	decoder *json.Decoder
}

type qusbCommand struct {
	Opcode   string   `json:"Opcode"`
	Space    string   `json:"Space"`
	Operands []string `json:"Operands"`
}

type qusbResult struct {
	Results []string `json:"Results"`
}

var timeout = time.Second * 2

func RandomName(prefix string) string {
	var bytes [4]byte
	_, _ = rand.Read(bytes[:])
	return fmt.Sprintf("%s-%08x", prefix, bytes)
}

func NewWebSocketClient(w *WebSocketClient, urlstr string, appName string) (err error) {
	w.closed = make(chan struct{})
	w.urlstr = urlstr
	w.appName = appName
	return w.Dial()
}

func (w *WebSocketClient) Dial() (err error) {
	//log.Printf("qusb2snes: [%s] dial %s", w.appName, w.urlstr)
	w.ws, _, _, err = ws.Dial(context.Background(), w.urlstr)
	if err != nil {
		err = fmt.Errorf("qusb2snes: [%s] dial: %w", w.appName, err)
		return
	}

	w.r = wsutil.NewClientSideReader(w.ws)
	w.w = wsutil.NewWriter(w.ws, ws.StateClientSide, ws.OpText)
	w.encoder = json.NewEncoder(w.w)
	w.decoder = json.NewDecoder(w.r)

	err = w.SendCommand(qusbCommand{
		Opcode:   "Name",
		Space:    "SNES",
		Operands: []string{w.appName},
	})
	if err != nil {
		var serr syscall.Errno
		if errors.As(err, &serr) {
			if !serr.Temporary() {
				w.Close()
			}
		}
	}

	return
}

func (w *WebSocketClient) checkOpen() (err error) {
	if w.ws == nil {
		err = fmt.Errorf("qusb2snes: websocket is closed")
	}
	return
}

func (w *WebSocketClient) Close() (err error) {
	//log.Printf("qusb2snes: [%s] close websocket\n", w.appName)

	defer w.lock.Unlock()
	w.lock.Lock()

	if w.ws != nil {
		err = w.ws.Close()
		close(w.closed)
	}

	w.closed = nil
	w.ws = nil
	w.r = nil
	w.w = nil
	w.encoder = nil
	w.decoder = nil

	return
}

func (w *WebSocketClient) SendCommand(cmd qusbCommand) (err error) {
	err = w.checkOpen()
	if err != nil {
		return
	}

	//log.Printf("qusb2snes: [%s] Encode(%s)\n", w.appName, cmd.Opcode)
	w.ws.SetWriteDeadline(time.Now().Add(timeout))
	err = w.encoder.Encode(cmd)
	if err != nil {
		var serr syscall.Errno
		if errors.As(err, &serr) {
			if !serr.Temporary() {
				w.Close()
			}
		}
		err = fmt.Errorf("qusb2snes: [%s] %s command encode: %w", w.appName, cmd.Opcode, err)
		return
	}

	//log.Printf("qusb2snes: [%s] Flush()\n", w.appName)
	w.ws.SetWriteDeadline(time.Now().Add(timeout))
	err = w.w.Flush()
	if err != nil {
		var serr syscall.Errno
		if errors.As(err, &serr) {
			if !serr.Temporary() {
				w.Close()
			}
		}
		err = fmt.Errorf("qusb2snes: [%s] %s command flush: %w", w.appName, cmd.Opcode, err)
		return
	}
	return
}

func (w *WebSocketClient) ReadCommandResponse(name string, rsp *qusbResult) (err error) {
	err = w.checkOpen()
	if err != nil {
		return
	}

	//log.Printf("qusb2snes: ReadCommandResponse: NextFrame(%s)\n", Name)
	w.ws.SetReadDeadline(time.Now().Add(timeout))
	hdr, err := w.r.NextFrame()
	if err != nil {
		var serr syscall.Errno
		if errors.As(err, &serr) {
			if !serr.Temporary() {
				w.Close()
			}
		}
		err = fmt.Errorf("qusb2snes: [%s] %s command response: error reading next websocket frame: %w", w.appName, name, err)
		return
	}
	if hdr.OpCode == ws.OpClose {
		w.Close()
		err = fmt.Errorf("qusb2snes: [%s] %s command response: websocket closed", w.appName, name)
		return
	}

	//log.Printf("qusb2snes: Decode(%s)\n", Name)
	w.ws.SetReadDeadline(time.Now().Add(timeout))
	err = w.decoder.Decode(rsp)
	if err != nil {
		var serr syscall.Errno
		if errors.As(err, &serr) {
			if !serr.Temporary() {
				w.Close()
			}
		}
		err = fmt.Errorf("qusb2snes: [%s] %s command response: decode response: %w", w.appName, name, err)
		return
	}

	//log.Println("qusb2snes: response received")
	return
}

func (w *WebSocketClient) ReadBinaryResponse(sumExpected int) (dataReceived []byte, err error) {
	err = w.checkOpen()
	if err != nil {
		return
	}

	// qusb2snes sends back randomly sized binary response messages:
	sumReceived := 0
	dataReceived = make([]byte, 0, sumExpected)
	for sumReceived < sumExpected {
		var hdr ws.Header
		w.ws.SetReadDeadline(time.Now().Add(timeout))

		hdr, err = w.r.NextFrame()
		if err != nil {
			err = fmt.Errorf("qusb2snes: ReadBinaryResponse: NextFrame: %w", err)
			w.Close()
			return
		}
		if hdr.OpCode == ws.OpClose {
			err = fmt.Errorf("qusb2snes: ReadBinaryResponse: NextFrame: server closed websocket")
			w.Close()
			return
		}
		if hdr.OpCode != ws.OpBinary {
			log.Printf("qusb2snes: ReadBinaryResponse: unexpected opcode %#x (expecting %#x)\n", hdr.OpCode, ws.OpBinary)
			return
		}

		var data []byte
		w.ws.SetReadDeadline(time.Now().Add(timeout))
		data, err = ioutil.ReadAll(w.r)
		if err != nil {
			err = fmt.Errorf("qusb2snes: ReadBinaryResponse: error reading binary response: %w", err)
			return
		}
		//log.Printf("qusb2snes: ReadBinaryResponse: %x binary bytes received\n", len(data))

		dataReceived = append(dataReceived, data...)
		sumReceived += len(data)
	}

	if sumReceived != sumExpected {
		err = fmt.Errorf("qusb2snes: ReadBinaryResponse: expected total of %x bytes but received %x", sumExpected, sumReceived)
		return
	}

	return
}
