package mapping

import (
	"sni/protos/sni"
	"sni/snes/mapping/exhirom"
	"sni/snes/mapping/hirom"
	"sni/snes/mapping/lorom"
)

func TranslateAddress(
	address uint32,
	space sni.AddressSpace,
	mapping sni.MemoryMapping,
	deviceSpace sni.AddressSpace,
) uint32 {
	switch space {
	case sni.AddressSpace_Raw:
		return address
	case sni.AddressSpace_FxPakPro:
		switch deviceSpace {
		case sni.AddressSpace_Raw:
			return address
		case sni.AddressSpace_FxPakPro:
			return address
		case sni.AddressSpace_SnesABus:
			switch mapping {
			case sni.MemoryMapping_LoROM:
				return lorom.PakAddressToBus(address)
			case sni.MemoryMapping_HiROM:
				return hirom.PakAddressToBus(address)
			case sni.MemoryMapping_ExHiROM:
				return exhirom.PakAddressToBus(address)
			}
		}
	case sni.AddressSpace_SnesABus:
		switch deviceSpace {
		case sni.AddressSpace_Raw:
			return address
		case sni.AddressSpace_SnesABus:
			return address
		case sni.AddressSpace_FxPakPro:
			switch mapping {
			case sni.MemoryMapping_LoROM:
				return lorom.BusAddressToPak(address)
			case sni.MemoryMapping_HiROM:
				return hirom.BusAddressToPak(address)
			case sni.MemoryMapping_ExHiROM:
				return exhirom.BusAddressToPak(address)
			}
		}
	}
	return address
}
