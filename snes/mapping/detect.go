package mapping

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"google.golang.org/grpc/codes"
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
	// set default:
	if fallbackMapping != nil {
		mapping = *fallbackMapping
	} else {
		mapping = sni.MemoryMapping_Unknown
	}

	if inHeaderBytes == nil {
		outHeaderBytes, err = detectHeader(ctx, memory, mapping)
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

func detectHeader(
	ctx context.Context,
	memory snes.DeviceMemory,
	fallbackMapping sni.MemoryMapping,
) (outHeaderBytes []byte, err error) {
	guessMappings := [3]sni.MemoryMapping{}
	guessMappings[0] = sni.MemoryMapping_LoROM

	// use the fallback mapping mode as a "guess" mapping used to read the ROM header:
	if fallbackMapping != sni.MemoryMapping_Unknown {
		guessMappings[0] = fallbackMapping
	}

	// fill in the remaining mappings to iterate over:
	if guessMappings[0] == sni.MemoryMapping_LoROM {
		guessMappings[1] = sni.MemoryMapping_HiROM
		guessMappings[2] = sni.MemoryMapping_ExHiROM
	} else if guessMappings[0] == sni.MemoryMapping_HiROM {
		guessMappings[1] = sni.MemoryMapping_LoROM
		guessMappings[2] = sni.MemoryMapping_ExHiROM
	} else if guessMappings[0] == sni.MemoryMapping_ExHiROM {
		guessMappings[1] = sni.MemoryMapping_LoROM
		guessMappings[2] = sni.MemoryMapping_HiROM
	}

	bestScore := -1
	for _, guessMapping := range guessMappings {
		var responses []snes.MemoryReadResponse
		tuple := snes.AddressTuple{
			Address:       uint32(0x00FFB0),
			AddressSpace:  sni.AddressSpace_SnesABus,
			MemoryMapping: guessMapping,
		}
		readRequest := snes.MemoryReadRequest{
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
			err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf("detect: %w: %s", err, &tuple))
			return
		}

		// score the header heuristically:
		header := snes.Header{}
		data := responses[0].Data
		err = header.ReadHeader(bytes.NewReader(data))
		if err != nil {
			err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf("detect: %w: %s", err, &tuple))
			return
		}
		score := header.Score()

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

	if bestScore < 0 {
		err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf(
			"detect: unable to detect valid ROM header",
		))
		return
	}

	return
}

func scoreHeader(data []byte) (score int) {
	header := snes.Header{}
	err := header.ReadHeader(bytes.NewReader(data))
	if err != nil {
		return -1
	}

	//score += 2*isFixed(&header->licensee, sizeof(header->licensee), 0x33);
	if header.OldMakerCode == 0x33 {
		score += 2
	}
	//score += 4*checkChksum(header->cchk, header->chk);
	if uint32(header.CheckSum)+uint32(header.ComplementCheckSum) == 0xffff {
		score += 4
	}

	if header.CartridgeType < 0x08 {
		score++
	}
	if header.ROMSize < 0x10 {
		score++
	}
	if header.RAMSize < 0x08 {
		score++
	}
	if header.DestinationCode < 0x0e {
		score++
	}

	if header.MapMode&0b0010_0000 == 0b0010_0000 {
		score++
	}

	return
}
