package mapping

import (
	"github.com/alttpo/snes/mapping/exhirom"
	"github.com/alttpo/snes/mapping/hirom"
	"github.com/alttpo/snes/mapping/lorom"
	"sni/devices"
	"sni/protos/sni"
)

type MemoryType string

// NOTE: these names (should) align with emunw-access protocol
const (
	MemoryTypeUnknown MemoryType = "unknown"
	MemoryTypeROM     MemoryType = "CARTROM"
	MemoryTypeSRAM    MemoryType = "SRAM"
	MemoryTypeWRAM    MemoryType = "WRAM"
)

func MemoryTypeFor(a devices.AddressTuple) (memoryType MemoryType, pakAddress uint32, offset uint32) {
	var err error

	switch a.AddressSpace {
	case sni.AddressSpace_FxPakPro:
		pakAddress, err = a.Address, nil
		break
	case sni.AddressSpace_SnesABus:
		switch a.MemoryMapping {
		case sni.MemoryMapping_LoROM:
			pakAddress, err = lorom.BusAddressToPak(a.Address)
		case sni.MemoryMapping_HiROM:
			pakAddress, err = hirom.BusAddressToPak(a.Address)
		case sni.MemoryMapping_ExHiROM:
			pakAddress, err = exhirom.BusAddressToPak(a.Address)
		}
	case sni.AddressSpace_Raw:
		err = ErrUnknownMapping
		break
	}

	if err != nil {
		memoryType = MemoryTypeUnknown
		return
	}

	memoryType, offset = MemoryTypeForPakAddress(pakAddress)
	return
}

func MemoryTypeForPakAddress(pakAddress uint32) (memoryType MemoryType, offset uint32) {
	if pakAddress < 0xE0_0000 {
		memoryType, offset = MemoryTypeROM, pakAddress
	} else if pakAddress < 0xF0_0000 {
		memoryType, offset = MemoryTypeSRAM, pakAddress-0xE0_0000
	} else if pakAddress < 0xF5_0000 {
		memoryType, offset = MemoryTypeUnknown, pakAddress-0xF0_0000
	} else if pakAddress < 0xF7_0000 {
		memoryType, offset = MemoryTypeWRAM, pakAddress-0xF5_0000
	} else {
		// TODO: VRAM, APU, CGRAM, OAM, MISC, PPUREG, CPUREG, etc.
		memoryType, offset = MemoryTypeUnknown, pakAddress-0xF7_0000
	}
	return
}
