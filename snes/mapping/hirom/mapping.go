package hirom

import (
	"sni/snes/mapping/util"
)

func BusAddressToPak(busAddr uint32) (pakAddr uint32, err error) {
	if busAddr >= 0xFE0000 && busAddr < 0x1_000000 {
		// ROM access:             $FE:0000-$FF:FFFF
		rom := (busAddr & 0x3FFFFF) + 0x000000
		return rom, nil
	} else if busAddr >= 0xC00000 && busAddr < 0xFE0000 {
		// ROM access:             $C0:0000-$FD:FFFF
		rom := (busAddr & 0x3FFFFF) + 0x000000
		return rom, nil
	} else if busAddr >= 0xA00000 && busAddr < 0xC00000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $A0:8000-$BF:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else if busAddr&0x7FFF >= 0x6000 {
			// TODO: A0-AF, B0-BF mirroring?
			// SRAM access:        $A0:6000-$BF:7FFF
			bank := (busAddr >> 16) - 0xA0
			sram := ((bank << 13) + (busAddr & 0x1FFF)) + 0xE00000
			return sram, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $A0:0000-$BF:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	} else if busAddr >= 0x800000 && busAddr < 0xA00000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $80:8000-$9F:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $80:0000-$9F:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	} else if busAddr >= 0x7E0000 && busAddr < 0x800000 {
		// WRAM access:
		wram := (busAddr - 0x7E0000) + 0xF50000
		return wram, nil
	} else if busAddr >= 0x400000 && busAddr < 0x7E0000 {
		// ROM access:             $40:0000-$7D:FFFF
		rom := (busAddr & 0x3FFFFF) + 0x000000
		return rom, nil
	} else if busAddr >= 0x200000 && busAddr < 0x400000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $20:8000-$3F:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else if busAddr&0x7FFF >= 0x6000 {
			// TODO: 20-2F, 30-3F mirroring?
			// SRAM access:        $20:6000-$3F:7FFF
			bank := (busAddr >> 16) - 0x20
			sram := ((bank << 13) + (busAddr & 0x1FFF)) + 0xE00000
			return sram, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $00:0000-$1F:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	} else if busAddr >= 0x000000 && busAddr < 0x200000 {
		if busAddr&0x8000 != 0 {
			// ROM access:         $00:8000-$1F:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x000000
			return rom, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $00:0000-$1F:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	}
	return 0, util.ErrUnmappedAddress
}

func PakAddressToBus(pakAddr uint32) (busAddr uint32, err error) {
	// WRAM is easy:
	if pakAddr >= 0xF50000 {
		// mirror bank $F7..FF back down into WRAM because these banks in FX Pak Pro space
		// are not available on the SNES bus; they are copies of otherwise inaccessible memory
		// like VRAM, CGRAM, OAM, etc.:
		busAddr = ((pakAddr - 0xF50000) & 0x01FFFF) + 0x7E0000
		return
	} else if pakAddr >= 0xE00000 && pakAddr < 0xF00000 {
		// SRAM is a little more complex, but not much:
		// TODO: handle A0-AF, B0-BF mirroring depending on SRAM size
		// bank $A0-$BF
		busAddr = pakAddr - 0xE00000
		offs := busAddr & 0x1FFF
		bank := (busAddr >> 13) & 0x1F
		busAddr = ((0xA0 + bank) << 16) + (offs + 0x6000)
		return
	} else if pakAddr < 0xE00000 {
		// ROM access:
		// HiROM is limited to $40 full banks

		busAddr = pakAddr & 0x3FFFFF
		// Accessing memory in banks $80-$FF is done at 3.58 MHz (120 ns) if the value at address $420D (hardware register) is set to 1.
		// Starting at $C0 gets us the full linear mapping of ROM without having to cut up in $8000 sized chunks:
		busAddr = 0xC00000 + busAddr
		return
	}
	return 0, util.ErrUnmappedAddress
}
