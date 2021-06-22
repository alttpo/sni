package mapping

import (
	"fmt"
	"sni/protos/sni"
	"sni/snes"
	"sni/snes/mapping/exhirom"
	"sni/snes/mapping/hirom"
	"sni/snes/mapping/lorom"
)

var ErrUnknownMapping = fmt.Errorf("cannot remap an address using an Unknown memory mapping; call MappingDetect to detect it from the ROM")

func TranslateAddress(
	sourceAddress snes.AddressTuple,
	deviceSpace sni.AddressSpace,
) (deviceAddress uint32, err error) {
	address := sourceAddress.Address
	switch sourceAddress.AddressSpace {
	case sni.AddressSpace_Raw:
		return address, nil
	case sni.AddressSpace_FxPakPro:
		switch deviceSpace {
		case sni.AddressSpace_Raw:
			return address, nil
		case sni.AddressSpace_FxPakPro:
			return address, nil
		case sni.AddressSpace_SnesABus:
			switch sourceAddress.MemoryMapping {
			case sni.MemoryMapping_LoROM:
				return lorom.PakAddressToBus(address)
			case sni.MemoryMapping_HiROM:
				return hirom.PakAddressToBus(address)
			case sni.MemoryMapping_ExHiROM:
				return exhirom.PakAddressToBus(address)
			default:
				return 0, ErrUnknownMapping
			}
		}
	case sni.AddressSpace_SnesABus:
		switch deviceSpace {
		case sni.AddressSpace_Raw:
			return address, nil
		case sni.AddressSpace_SnesABus:
			return address, nil
		case sni.AddressSpace_FxPakPro:
			switch sourceAddress.MemoryMapping {
			case sni.MemoryMapping_LoROM:
				return lorom.BusAddressToPak(address)
			case sni.MemoryMapping_HiROM:
				return hirom.BusAddressToPak(address)
			case sni.MemoryMapping_ExHiROM:
				return exhirom.BusAddressToPak(address)
			default:
				return 0, ErrUnknownMapping
			}
		}
	}
	return address, nil
}
