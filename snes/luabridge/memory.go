package luabridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
	"time"
)

const readWriteTimeout = time.Second * 15

func (d *Device) MultiReadMemory(ctx context.Context, reads ...snes.MemoryReadRequest) (rsp []snes.MemoryReadResponse, err error) {
	defer func() {
		if err != nil {
			rsp = nil
			closeErr := d.Close()
			if closeErr != nil {
				log.Printf("luabridge: close error: %v\n", closeErr)
			}
		}
	}()

	deadline, ok := ctx.Deadline()
	if !ok {
		deadline = time.Now().Add(readWriteTimeout)
	}

	rsp = make([]snes.MemoryReadResponse, len(reads))
	for j, read := range reads {
		addr := mapping.TranslateAddress(
			read.RequestAddress,
			read.RequestAddressSpace,
			d.Mapping,
			sni.AddressSpace_SnesABus,
		)

		sb := bytes.NewBuffer(make([]byte, 0, 64))
		_, _ = fmt.Fprintf(sb, "Read|%d|%d\n", addr, read.Size)

		data := make([]byte, 65536)
		var n int
		n, err = d.WriteThenRead(sb.Bytes(), data, deadline)
		if err != nil {
			return
		}

		// parse response as json:
		type tmpResultJson struct {
			Data []byte `json:"data"`
		}
		tmp := tmpResultJson{}
		err = json.Unmarshal(data[:n], &tmp)
		if err != nil {
			return
		}
		if actual, expected := len(tmp.Data), read.Size; actual != expected {
			err = fmt.Errorf("response did not provide enough data to meet request size; actual $%x, expected $%x", actual, expected)
			return
		}

		rsp[j] = snes.MemoryReadResponse{
			MemoryReadRequest:  read,
			DeviceAddress:      addr,
			DeviceAddressSpace: sni.AddressSpace_SnesABus,
			Data:               tmp.Data,
		}
	}

	return
}

func (d *Device) MultiWriteMemory(ctx context.Context, writes ...snes.MemoryWriteRequest) ([]snes.MemoryWriteResponse, error) {
	panic("implement me")
}
