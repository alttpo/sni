package mapping

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"sni/protos/sni"
	"sni/snes"
)

func Detect(
	ctx context.Context,
	memory snes.DeviceMemory,
	fallbackMapping *sni.MemoryMapping,
	inHeaderBytes []byte,
) (mapping sni.MemoryMapping, confidence bool, outHeaderBytes []byte, err error) {
	if inHeaderBytes == nil {
		// use the fallback mapping mode as a "guess" mapping used to read the ROM header:
		guessMapping := sni.MemoryMapping_LoROM
		if fallbackMapping != nil && *fallbackMapping != sni.MemoryMapping_Unknown {
			guessMapping = *fallbackMapping
			// TODO: could loop over all mapping modes and attempt reads
			// then use heuristics to detect the ROM header
		}

		// read the ROM header:
		var responses []snes.MemoryReadResponse
		readRequest := snes.MemoryReadRequest{
			RequestAddress:      0x40FFB0,
			RequestAddressSpace: sni.AddressSpace_SnesABus,
			RequestMapping:      guessMapping,
			Size:                0x50,
		}
		log.Printf(
			"detect: read {address:%s($%06x),size:$%x}\n",
			sni.AddressSpace_name[int32(readRequest.RequestAddressSpace)],
			readRequest.RequestAddress,
			readRequest.Size,
		)
		responses, err = memory.MultiReadMemory(ctx, readRequest)
		if err != nil {
			return
		}

		outHeaderBytes = responses[0].Data

		log.Printf(
			"detect: read {address:%s($%06x),size:$%x} complete:\n%s",
			sni.AddressSpace_name[int32(responses[0].DeviceAddressSpace)],
			responses[0].DeviceAddress,
			len(outHeaderBytes),
			hex.Dump(outHeaderBytes),
		)
	} else {
		if len(inHeaderBytes) < 0x30 {
			err = fmt.Errorf("input ROM header must be at least $30 bytes")
			return
		}
		outHeaderBytes = inHeaderBytes
		log.Printf(
			"detect: provided header bytes {size:$%x}:\n%s",
			len(outHeaderBytes),
			hex.Dump(outHeaderBytes),
		)
	}

	header := snes.Header{}
	err = header.ReadHeader(bytes.NewReader(outHeaderBytes))
	if err != nil {
		return
	}

	// detection does not have to be perfect (and never could be) since the client
	// always has the ability to override it or not use it at all and set their own
	// memory mapping.

	log.Printf(
		"detect: map mode %02x\n",
		header.MapMode&0b1110_1111,
	)

	confidence = true

	// mask off SlowROM vs FastROM bit:
	switch header.MapMode & 0b1110_1111 {
	case 0x20: // LoROM
		mapping = sni.MemoryMapping_LoROM
	case 0x21: // HiROM
		mapping = sni.MemoryMapping_HiROM
	case 0x22: // ExLoROM
		mapping = sni.MemoryMapping_LoROM
	case 0x23: // SA-1
		mapping = sni.MemoryMapping_HiROM
	case 0x25: // ExHiROM
		mapping = sni.MemoryMapping_ExHiROM
	default:
		confidence = false
		if fallbackMapping != nil {
			mapping = *fallbackMapping
			log.Printf(
				"detect: unable to detect mapping mode; falling back to provided default %s\n",
				sni.MemoryMapping_name[int32(mapping)],
			)
		} else {
			// revert to a simple LoROM vs HiROM:
			mapping = sni.MemoryMapping_LoROM - sni.MemoryMapping(header.MapMode&1)
			log.Printf(
				"detect: unable to detect mapping mode; guessing %s\n",
				sni.MemoryMapping_name[int32(mapping)],
			)
		}
	}

	if confidence {
		log.Printf(
			"detect: detected mapping mode = %s\n",
			sni.MemoryMapping_name[int32(mapping)],
		)
	}

	return
}
