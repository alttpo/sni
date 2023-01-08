package luabridge

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"sni/cmd/sni/config"
	"sni/devices"
	"sni/devices/snes/mapping"
	"sni/protos/sni"
	"strconv"
	"time"
)
import jsoniter "github.com/json-iterator/go"

var json = jsoniter.ConfigCompatibleWithStandardLibrary

//var json = jsoniter.ConfigFastest

const readWriteTimeout = time.Second * 15

func (d *Device) DefaultAddressSpace(context.Context) (sni.AddressSpace, error) {
	return defaultAddressSpace, nil
}

func (d *Device) RequiresMemoryMappingForAddressSpace(ctx context.Context, addressSpace sni.AddressSpace) (bool, error) {
	if d.isBizHawk {
		if addressSpace == sni.AddressSpace_FxPakPro {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		if addressSpace == sni.AddressSpace_SnesABus {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func (d *Device) RequiresMemoryMappingForAddress(ctx context.Context, address devices.AddressTuple) (bool, error) {
	if d.isBizHawk {
		if address.AddressSpace == sni.AddressSpace_FxPakPro {
			return false, nil
		} else {
			return true, nil
		}
	} else {
		if address.AddressSpace == sni.AddressSpace_SnesABus {
			return false, nil
		} else {
			return true, nil
		}
	}
}

func (d *Device) MultiReadMemory(ctx context.Context, reads ...devices.MemoryReadRequest) (rsp []devices.MemoryReadResponse, err error) {
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

	rsp = make([]devices.MemoryReadResponse, len(reads))
	for j, read := range reads {
		var addr uint32
		var addressSpace sni.AddressSpace

		sb := bytes.NewBuffer(make([]byte, 0, 64))
		if d.isBizHawk {
			var domain mapping.MemoryType
			var offset uint32
			addressSpace = sni.AddressSpace_FxPakPro
			domain, addr, offset = mapping.MemoryTypeFor(read.RequestAddress)
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

		var rspstr []byte
		rspstr, err = d.WriteThenReadUntilNewline(sb.Bytes(), deadline)
		if err != nil {
			return
		}

		if config.LogResponses {
			log.Printf("luabridge: < %s", rspstr)
		}

		var data []byte
		if d.clientName == "SNI Connector" {
			version, _ := strconv.Atoi(d.version)
			if version >= 3 {
				// parse response as hex bytes:
				data, err = d.parseHexResponse(rspstr)
			} else {
				// parse response as json:
				data, err = d.parseJsonResponse(rspstr)
			}
		} else {
			// parse response as json:
			data, err = d.parseJsonResponse(rspstr)
		}
		if err != nil {
			err = d.FatalError(err)
			return
		}

		if actual, expected := len(data), read.Size; actual != expected {
			err = fmt.Errorf("response did not provide enough data to meet request size; actual $%x, expected $%x", actual, expected)
			err = d.FatalError(err)
			return
		}

		rsp[j] = devices.MemoryReadResponse{
			RequestAddress: read.RequestAddress,
			DeviceAddress: devices.AddressTuple{
				Address:       addr,
				AddressSpace:  addressSpace,
				MemoryMapping: read.RequestAddress.MemoryMapping,
			},
			Data: data,
		}
	}

	return
}

func (d *Device) parseHexResponse(hexstr []byte) (data []byte, err error) {
	err = nil
	tnl := bytes.LastIndexByte(hexstr, '\n')
	if tnl < 0 {
		err = fmt.Errorf("invalid response: no newline terminator found")
		return
	}

	data = make([]byte, tnl/2)
	for i := range data {
		v := byte(0)

		c := hexstr[i*2+0]
		if 'A' <= c && c <= 'F' {
			v += c - 'A' + 10
		} else if 'a' <= c && c <= 'f' {
			v += c - 'a' + 10
		} else if '0' <= c && c <= '9' {
			v += c - '0'
		} else {
			err = fmt.Errorf("invalid character '%c' seen in hex byte stream position %d", c, i*2+0)
			return
		}
		v <<= 4

		c = hexstr[i*2+1]
		if 'A' <= c && c <= 'F' {
			v += c - 'A' + 10
		} else if 'a' <= c && c <= 'f' {
			v += c - 'a' + 10
		} else if '0' <= c && c <= '9' {
			v += c - '0'
		} else {
			err = fmt.Errorf("invalid character '%c' seen in hex byte stream position %d", c, i*2+1)
			return
		}

		data[i] = v
	}

	return
}

func (d *Device) parseJsonResponse(rspstr []byte) (data []byte, err error) {
	type tmpResultJson struct {
		Data []byte `json:"data"`
	}

	tmp := tmpResultJson{}
	err = json.Unmarshal(rspstr, &tmp)
	if err != nil {
		return
	}

	data = tmp.Data
	return
}

func (d *Device) MultiWriteMemory(ctx context.Context, writes ...devices.MemoryWriteRequest) (rsp []devices.MemoryWriteResponse, err error) {
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

	rsp = make([]devices.MemoryWriteResponse, len(writes))
	for j, write := range writes {
		var addr uint32
		var addressSpace sni.AddressSpace

		// preallocate enough space to write the whole command:
		sb := bytes.NewBuffer(make([]byte, 0, 24+4*len(write.Data)))
		if d.isBizHawk {
			var domain mapping.MemoryType
			var offset uint32
			addressSpace = sni.AddressSpace_FxPakPro
			domain, addr, offset = mapping.MemoryTypeFor(write.RequestAddress)
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

		rsp[j] = devices.MemoryWriteResponse{
			RequestAddress: write.RequestAddress,
			DeviceAddress: devices.AddressTuple{
				Address:       addr,
				AddressSpace:  addressSpace,
				MemoryMapping: write.RequestAddress.MemoryMapping,
			},
			Size: len(write.Data),
		}
	}

	return
}
