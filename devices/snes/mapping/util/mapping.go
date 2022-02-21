package util

func BankToLinear(addr uint32) uint32 {
	bank := addr >> 16
	linoffs := (bank << 15) + (addr & 0x7FFF)
	return linoffs
}
