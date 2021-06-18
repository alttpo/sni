package lorom

import (
	"sni/snes/util"
)

func BusAddressToPC(busAddr uint32) uint32 {
	page := busAddr & 0xFFFF
	if page < 0x8000 {
		return 0x1000000
	}

	bank := busAddr >> 16
	pcAddr := (bank << 15) | (page - 0x8000)
	return pcAddr
}

func BusAddressToPak(busAddr uint32) uint32 {
	if busAddr >= 0x700000 && busAddr < 0x7E0000 {
		sram := (busAddr-0x700000) + 0xE00000
		return sram
	} else if busAddr >= 0x7E0000 && busAddr < 0x800000 {
		wram := (busAddr - 0x7E0000) + 0xF50000
		return wram
	} else if busAddr & 0x8000 != 0 {
		sram := util.BankToLinear(busAddr & 0x3FFFFF) + 0x000000
		return sram
	}
	return busAddr
}

func PakAddressToBus(pakAddr uint32) uint32 {
	// WRAM is easy:
	if pakAddr >= 0xF50000 && pakAddr < 0xF70000 {
		return pakAddr - 0xF50000 + 0x7E0000
	}
	// SRAM is a little more complex, but not much:
	if pakAddr >= 0xE00000 && pakAddr < 0xF00000 {
		// bank $70-$7D
		busAddr := (pakAddr - 0xE00000) & 0x0DFFFF
		offs := busAddr & 0x7FFF
		bank := busAddr >> 15
		busAddr = ((0x70 + bank) << 16) + offs
		return busAddr
	}
	// ROM access:
	if pakAddr < 0xE00000 {
		busAddr := pakAddr & 0x3FFFFF
		offs := busAddr & 0x7FFF
		bank := busAddr >> 15
		busAddr = (bank << 16) + (offs | 0x8000)
		return busAddr
	}
	// /shrug
	return pakAddr
}
