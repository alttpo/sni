package lorom

import (
	"sni/devices/snes/mapping/util"
)

func BusAddressToPak(busAddr uint32) (pakAddr uint32, err error) {
	if busAddr >= 0xF00000 && busAddr < 0x1_000000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $F0:8000-$F0:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else {
			// SRAM access:        $F0:0000-$FF:7FFF
			sram := util.BankToLinear(busAddr-0xF00000) + 0xE00000
			return sram, nil
		}
	} else if busAddr >= 0x800000 && busAddr < 0xF00000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $80:8000-$EF:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $80:0000-$EF:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	} else if busAddr >= 0x7E0000 && busAddr < 0x800000 {
		// WRAM access:
		wram := (busAddr - 0x7E0000) + 0xF50000
		return wram, nil
	} else if busAddr >= 0x700000 && busAddr < 0x7E0000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $70:8000-$7D:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else {
			// SRAM access:        $70:0000-$7D:7FFF
			sram := util.BankToLinear(busAddr-0x700000) + 0xE00000
			return sram, nil
		}
	} else if busAddr < 0x700000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $00:8000-$6F:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $00:0000-$6F:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	}
	return 0, util.ErrUnmappedAddress
}

func PakAddressToBus(pakAddr uint32) (busAddr uint32, err error) {
	// WRAM is easy:
	if pakAddr >= 0xF50000 && pakAddr < 0x1_000000 {
		// mirror bank $F7..FF back down into WRAM because these banks in FX Pak Pro space
		// are not available on the SNES bus; they are copies of otherwise inaccessible memory
		// like VRAM, CGRAM, OAM, etc.:
		busAddr = ((pakAddr - 0xF50000) & 0x01FFFF) + 0x7E0000
		return
	} else if pakAddr >= 0xE00000 && pakAddr < 0xF00000 {
		// SRAM is a little more complex, but not much:
		// bank $F0-$FF, $0000-$7FFF
		busAddr = (pakAddr - 0xE00000) & 0x07FFFF
		offs := busAddr & 0x7FFF
		bank := busAddr >> 15
		busAddr = ((0xF0 + bank) << 16) + offs
		return
	} else if pakAddr < 0xE00000 {
		// ROM access:
		busAddr = pakAddr & 0x3FFFFF
		offs := busAddr & 0x7FFF
		bank := busAddr >> 15
		busAddr = ((0x80 + bank) << 16) + (offs | 0x8000)
		return
	}
	return 0, util.ErrUnmappedAddress
}
