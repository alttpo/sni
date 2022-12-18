package fxpakpro

type opcode uint8

const (
	OpGET opcode = iota
	OpPUT
	OpVGET
	OpVPUT

	OpLS
	OpMKDIR
	OpRM
	OpMV

	OpRESET
	OpBOOT
	OpPOWER_CYCLE
	OpINFO
	OpMENU_RESET
	OpSTREAM
	OpTIME
	OpRESPONSE
)

type space uint8

func (s space) String() string {
	switch s {
	case SpaceFILE:
		return "FILE"
	case SpaceSNES:
		return "SNES"
	case SpaceMSU:
		return "MSU"
	case SpaceCMD:
		return "CMD"
	case SpaceCONFIG:
		return "CONFIG"
	default:
		return "unknown"
	}
}

const (
	SpaceFILE space = iota
	SpaceSNES
	SpaceMSU
	SpaceCMD
	SpaceCONFIG
)

type server_flags uint8

const FlagNONE server_flags = 0
const (
	FlagSKIPRESET server_flags = 1 << iota
	FlagONLYRESET
	FlagCLRX
	FlagSETX
	FlagSTREAM_BURST
	_
	FlagNORESP
	FlagDATA64B
)

type info_flags uint8

const (
	FeatDSPX info_flags = 1 << iota
	FeatST0010
	FeatSRTC
	FeatMSU1
	Feat213F
	FeatCMD_UNLOCK
	FeatUSB1
	FeatDMA1
)

type file_type uint8

const (
	FtDIRECTORY file_type = 0
	FtFILE      file_type = 1
)
