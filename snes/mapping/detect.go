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
	"strings"
)

func Detect(
	ctx context.Context,
	useMemory snes.UseMemory,
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
		// use the fallback mapping mode as a "guess" mapping used to read the ROM header:
		guessMapping := sni.MemoryMapping_LoROM
		if fallbackMapping != nil && *fallbackMapping != sni.MemoryMapping_Unknown {
			guessMapping = *fallbackMapping
			// TODO: could loop over all mapping modes and attempt reads
			// then use heuristics to detect the ROM header
		}

		var deviceAddress = snes.AddressTuple{}
		headerAddresses := []uint32{0x00FFB0, 0x40FFB0, 0x80FFB0, 0xC0FFB0}
		errors := make([]string, len(headerAddresses))

		_ = useMemory.UseMemory(
			ctx,
			[]sni.DeviceCapability{sni.DeviceCapability_ReadMemory},
			func(mctx context.Context, memory snes.DeviceMemory) error {
				for j, headerAddress := range headerAddresses {
					// read the ROM header:
					var responses []snes.MemoryReadResponse
					tuple := snes.AddressTuple{
						Address:       headerAddress,
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
					responses, err = memory.MultiReadMemory(mctx, readRequest)
					if err != nil {
						err = fmt.Errorf("%w: %s", err, &tuple)
						errors[j] = err.Error()
						log.Printf("detect: %v\n", errors[j])
						continue
					}

					outHeaderBytes = responses[0].Data
					deviceAddress = responses[0].DeviceAddress
					deviceAddress.MemoryMapping = tuple.MemoryMapping
					break
				}
				return nil
			},
		)

		if outHeaderBytes == nil {
			err = snes.WithCode(codes.FailedPrecondition, fmt.Errorf(
				"detect: unable to read ROM header:\n%v",
				strings.Join(errors, "\n"),
			))
			return
		}

		log.Printf(
			"detect: read {address:%s,size:$%x} complete:\n%s",
			&deviceAddress,
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
