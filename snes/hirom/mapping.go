package hirom

import (
	"sni/snes/util"
)

func BusAddressToPak(busAddr uint32) uint32 {
	if busAddr&0x8000 == 0 {
		if busAddr >= 0x700000 && busAddr < 0x7E0000 {
			sram := util.BankToLinear(busAddr-0x700000) + 0xE00000
			return sram
		} else if busAddr >= 0x7E0000 && busAddr < 0x800000 {
			wram := (busAddr - 0x7E0000) + 0xF50000
			return wram
		}
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
		// bank $A0-$BF
		busAddr := pakAddr - 0xE00000
		offs := busAddr & 0x1FFF
		bank := (busAddr >> 13) & 0x1F
		busAddr = ((0xA0 + bank) << 16) + (offs + 0x6000)
		return busAddr
	}
	// ROM access:
	if pakAddr < 0xE00000 {
		// HiROM is limited to $40 full banks
		busAddr := pakAddr & 0x3FFFFF
		// Accessing memory in banks $80-$FF is done at 3.58 MHz (120 ns) if the value at address $420D (hardware register) is set to 1.
		// Starting at $C0 gets us the full linear mapping of ROM without having to cut up in $8000 sized chunks:
		busAddr = 0xC00000 + busAddr
		return busAddr
	}
	// /shrug
	return pakAddr
}
