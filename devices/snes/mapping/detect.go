package mapping

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/alttpo/snes"
	"google.golang.org/grpc/codes"
	"log"
	"sni/devices"
	"sni/protos/sni"
)

func Detect(
	ctx context.Context,
	memory devices.DeviceMemory,
	fallbackMapping *sni.MemoryMapping,
	inHeaderBytes []byte,
) (mapping sni.MemoryMapping, confidence bool, outHeaderBytes []byte, err error) {
	// set default:
	if fallbackMapping != nil {
		mapping = *fallbackMapping
	} else {
		mapping = sni.MemoryMapping_Unknown
	}

	if inHeaderBytes == nil {
		outHeaderBytes, err = detectHeader(ctx, memory)
		if err != nil {
			return
		}
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

func detectHeader(ctx context.Context, memory devices.DeviceMemory) (outHeaderBytes []byte, err error) {
	addresses := [3]uint32{
		uint32(0x007FB0),
		uint32(0x00FFB0),
		uint32(0x40FFB0),
	}
	mappings := []sni.MemoryMapping{
		sni.MemoryMapping_LoROM,
		sni.MemoryMapping_HiROM,
		sni.MemoryMapping_ExHiROM,
	}

	bestScore := 0
	for _, address := range addresses {
		for _, mapping := range mappings {
			var responses []devices.MemoryReadResponse
			tuple := devices.AddressTuple{
				Address:       address,
				AddressSpace:  sni.AddressSpace_FxPakPro,
				MemoryMapping: mapping,
			}
			readRequest := devices.MemoryReadRequest{
				RequestAddress: tuple,
				Size:           0x50,
			}
			log.Printf(
				"detect: read {address:%s,size:$%x}\n",
				&tuple,
				readRequest.Size,
			)

			// read the ROM header:
			responses, err = memory.MultiReadMemory(ctx, readRequest)
			if err != nil {
				if devices.IsFatal(err) {
					err = fmt.Errorf("detect: %w: %s", err, &tuple)
					return
				}
				log.Printf("detect: ignoring non-fatal error: %v\n", err)
				err = nil
				continue
			}

			// score the header heuristically:
			header := snes.Header{}
			data := responses[0].Data
			err = header.ReadHeader(bytes.NewReader(data))
			if err != nil {
				err = devices.WithCode(codes.FailedPrecondition, fmt.Errorf("detect: %w: %s", err, &tuple))
				return
			}
			score := header.Score(address)

			log.Printf(
				"detect: read {address:%s,deviceAddress:%s,size:$%x} complete: score=%d\n%s",
				&tuple,
				&responses[0].DeviceAddress,
				len(data),
				score,
				hex.Dump(data),
			)

			if score > bestScore {
				bestScore = score
				outHeaderBytes = data
			}
		}
	}

	if bestScore <= 0 {
		err = devices.DeviceNonFatal("detect: unable to detect valid ROM header", nil)
		return
	}

	return
}
