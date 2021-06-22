package exhirom

import "sni/snes/mapping/util"

// https://thepoorstudenthobbyist.com/2019/05/18/custom-pcb-explanation/#exhirom

func BusAddressToPak(busAddr uint32) (pakAddr uint32, err error) {
	if busAddr >= 0xC00000 && busAddr < 0x1_000000 {
		// program area 1
		// ROM access:             $C0:0000-$FF:FFFF
		rom := (busAddr & 0x3FFFFF) + 0x000000
		return rom, nil
	} else if busAddr >= 0xA00000 && busAddr < 0xC00000 {
		if busAddr&0x8000 != 0 {
			// program area 1
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
			// program area 1
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
		// program area 2
		// ROM access:             $40:0000-$7D:FFFF
		rom := (busAddr & 0x3FFFFF) + 0x400000
		return rom, nil
	} else if busAddr >= 0x3E0000 && busAddr < 0x400000 {
		if busAddr&0x8000 != 0 {
			// program area 3 $3E-3F
			// ROM access:             $3E:8000-$3F:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x400000
			return rom, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $00:0000-$1F:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	} else if busAddr >= 0x200000 && busAddr < 0x3E0000 {
		if busAddr&0x8000 != 0 {
			// program area 2
			// ROM access:         $20:8000-$3D:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x400000
			return rom, nil
		} else if busAddr&0xFFFF < 0x2000 {
			// Lower 8KiB of WRAM: $00:0000-$1F:1FFF
			wram := (busAddr & 0x1FFF) + 0xF50000
			return wram, nil
		}
	} else if busAddr < 0x200000 {
		if busAddr&0x8000 != 0 {
			// program area 2
			// ROM access:         $00:8000-$1F:FFFF
			rom := util.BankToLinear(busAddr&0x3F7FFF) + 0x400000
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
	if pakAddr >= 0xF50000 {
		// WRAM is easy:
		// mirror bank $F7..FF back down into WRAM because these banks in FX Pak Pro space
		// are not available on the SNES bus; they are copies of otherwise inaccessible memory
		// like VRAM, CGRAM, OAM, etc.:
		busAddr = ((pakAddr - 0xF50000) & 0x01FFFF) + 0x7E0000
		return
	} else if pakAddr >= 0xE00000 && pakAddr < 0xF00000 {
		// SRAM is a little more complex, but not much:
		// TODO: handle A0-AF, B0-BF mirroring depending on SRAM size
		// A0:6000-7FFF to BF:6000-7FFF
		busAddr = pakAddr - 0xE00000
		offs := busAddr & 0x1FFF
		bank := (busAddr >> 13) & 0x1F
		busAddr = ((0xA0 + bank) << 16) + (offs + 0x6000)
		return
	} else if pakAddr < 0xE00000 {
		// ROM access:
		if pakAddr >= 0x800000 {
			// map FX Pak Pro $800000 to SlowROM banks $00-3F (mirrored to $40..$5F)
			busAddr = (pakAddr - 0x800000) & 0x3FFFFF
			offs := busAddr & 0x7FFF
			bank := busAddr >> 15
			// avoid WRAM conflict area:
			if bank >= 0x7E {
				bank += 0x80
			}
			// $00:8000-FFFF to $7D:8000-FFFF
			// $FE:8000-FFFF to $FF:8000-FFFF
			busAddr = (bank << 16) + (offs | 0x8000)
		} else if pakAddr >= 0x7E0000 && pakAddr < 0x800000 {
			// program ROM area 3 is top half of banks $3E and $3F
			busAddr = pakAddr - 0x7E0000
			offs := busAddr & 0x7FFF
			bank := 0x3E + (busAddr >> 15)
			// $3E:8000-FFFF to $3F:8000-FFFF
			busAddr = (bank << 16) + (offs | 0x8000)
		} else if pakAddr >= 0x400000 && pakAddr < 0x7E0000 {
			// program ROM area 2 is full banks $40-$7D
			// $40:0000-FFFF to $7D:0000-FFFF
			busAddr = 0x400000 + (pakAddr & 0x3FFFFF)
		} else {
			// program ROM area 1 is full banks $C0-$FF
			// $C0:0000-FFFF to $FF:0000-FFFF
			busAddr = 0xC00000 + (pakAddr & 0x3FFFFF)
		}
		return
	}
	return 0, util.ErrUnmappedAddress
}
