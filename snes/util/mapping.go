package util

func BankToLinear(addr uint32) uint32 {
	bank := addr >> 16
	linbank := ((bank & 1) << 15) + ((bank >> 1) << 16)
	linoffs := linbank + (addr & 0x7FFF)
	return linoffs
}
