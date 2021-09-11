package luabridge

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sni/cmd/sni/config"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping"
	"time"
)

const readWriteTimeout = time.Second * 15

func (d *Device) DefaultAddressSpace(context.Context) (sni.AddressSpace, error) {
	return defaultAddressSpace, nil
}

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
		var addr uint32
		var addressSpace sni.AddressSpace

		sb := bytes.NewBuffer(make([]byte, 0, 64))
		if d.isBizHawk {
			var domain mapping.MemoryType
			var offset uint32
			addressSpace = sni.AddressSpace_FxPakPro
			domain, addr, offset = mapping.MemoryTypeFor(&read.RequestAddress)
			if domain == "SRAM" {
				domain = "CARTRAM"
			}
			_, _ = fmt.Fprintf(sb, "Read|%d|%d|%s\n", offset, read.Size, domain)
		} else {
			addressSpace = sni.AddressSpace_SnesABus
			addr, err = mapping.TranslateAddress(read.RequestAddress, addressSpace)
			if err != nil {
				return
			}
			_, _ = fmt.Fprintf(sb, "Read|%d|%d\n", addr, read.Size)
		}

		if config.VerboseLogging {
			log.Printf("luabridge: > %s", sb.Bytes())
		}

		data := make([]byte, 65536)
		var n int
		n, err = d.WriteThenRead(sb.Bytes(), data, deadline)
		if err != nil {
			return
		}

		data = data[:n]
		if config.VerboseLogging {
			log.Printf("luabridge: < %s", data)
		}

		// parse response as json:
		type tmpResultJson struct {
			Data []byte `json:"data"`
		}
		tmp := tmpResultJson{}
		err = json.Unmarshal(data, &tmp)
		if err != nil {
			return
		}
		if actual, expected := len(tmp.Data), read.Size; actual != expected {
			err = fmt.Errorf("response did not provide enough data to meet request size; actual $%x, expected $%x", actual, expected)
			err = d.FatalError(err)
			return
		}

		rsp[j] = snes.MemoryReadResponse{
			RequestAddress: read.RequestAddress,
			DeviceAddress: snes.AddressTuple{
				Address:       addr,
				AddressSpace:  addressSpace,
				MemoryMapping: read.RequestAddress.MemoryMapping,
			},
			Data: tmp.Data,
		}
	}

	return
}

func (d *Device) MultiWriteMemory(ctx context.Context, writes ...snes.MemoryWriteRequest) (rsp []snes.MemoryWriteResponse, err error) {
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

	rsp = make([]snes.MemoryWriteResponse, len(writes))
	for j, write := range writes {
		var addr uint32
		var addressSpace sni.AddressSpace

		// preallocate enough space to write the whole command:
		sb := bytes.NewBuffer(make([]byte, 0, 24+4*len(write.Data)))
		if d.isBizHawk {
			var domain mapping.MemoryType
			var offset uint32
			addressSpace = sni.AddressSpace_FxPakPro
			domain, addr, offset = mapping.MemoryTypeFor(&write.RequestAddress)
			if domain == "SRAM" {
				domain = "CARTRAM"
			}
			_, _ = fmt.Fprintf(sb, "Write|%d|%s", offset, domain)
		} else {
			addressSpace = sni.AddressSpace_SnesABus
			addr, err = mapping.TranslateAddress(write.RequestAddress, addressSpace)
			if err != nil {
				return
			}
			_, _ = fmt.Fprintf(sb, "Write|%d", addr)
		}
		for _, b := range write.Data {
			_, _ = fmt.Fprintf(sb, "|%d", b)
		}
		sb.WriteByte('\n')

		if config.VerboseLogging {
			log.Printf("luabridge: > %s", sb.Bytes())
		}

		// send the command:
		var n int
		n, err = d.WriteDeadline(sb.Bytes(), deadline)
		if err != nil {
			return
		}
		_ = n

		rsp[j] = snes.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			DeviceAddress: snes.AddressTuple{
				Address:       addr,
				AddressSpace:  addressSpace,
				MemoryMapping: write.RequestAddress.MemoryMapping,
			},
			Size: len(write.Data),
		}
	}

	return
}
