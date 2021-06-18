package exhirom

import (
	"sni/snes/util"
)

// https://thepoorstudenthobbyist.com/2019/05/18/custom-pcb-explanation/#exhirom

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
	if pakAddr >= 0xF50000 {
		// mirror bank $F7..FF back down into WRAM because these banks in FX Pak Pro space
		// are not available on the SNES bus; they are copies of otherwise inaccessible memory
		// like VRAM, CGRAM, OAM, etc.:
		return ((pakAddr - 0xF50000) & 0x01FFFF) + 0x7E0000
	}
	// SRAM is a little more complex, but not much:
	if pakAddr >= 0xE00000 && pakAddr < 0xF00000 {
		busAddr := pakAddr - 0xE00000
		offs := busAddr & 0x1FFF
		bank := (busAddr >> 13) & 0x1F
		busAddr = ((0xA0 + bank) << 16) + (offs + 0x6000)
		return busAddr
	}
	// ROM access:
	if pakAddr < 0xE00000 {
		var busAddr uint32
		if pakAddr >= 0x800000 {
			// map FX Pak Pro $800000 to SlowROM banks $00-3F (mirrored to $40..$5F)
			busAddr = (pakAddr - 0x800000) & 0x3FFFFF
			offs := busAddr & 0x7FFF
			bank := busAddr >> 15
			// avoid WRAM conflict area:
			if bank >= 0x7E {
				bank += 0x80
			}
			busAddr = (bank << 16) + (offs | 0x8000)
		} else if pakAddr >= 0x7E0000 && pakAddr < 0x800000 {
			// program ROM area 3 is top half of banks $3E and $3F
			busAddr = pakAddr - 0x7E0000
			offs := busAddr & 0x7FFF
			bank := 0x3E + (busAddr >> 15)
			busAddr = (bank << 16) + (offs | 0x8000)
		} else if pakAddr >= 0x400000 && pakAddr < 0x7E0000 {
			// program ROM area 2 is full banks $40-$7D
			busAddr = 0x400000 + (pakAddr & 0x3FFFFF)
		} else {
			// program ROM area 1 is full banks $C0-$FF
			busAddr = 0xC00000 + (pakAddr & 0x3FFFFF)
		}
		return busAddr
	}
	// /shrug
	return pakAddr
}
