package asm

import (
	"bytes"
	"fmt"
	"strings"
)

// Emitter implements Assembler and bytes.Buffer; a 65816 immediate assembler that emits to the buffer
type Emitter struct {
	flagsTracker

	Code *bytes.Buffer
	Text *strings.Builder

	address uint32
	baseSet bool
}

func (a *Emitter) Clone() *Emitter {
	return &Emitter{
		flagsTracker: a.flagsTracker,
		Code:         &bytes.Buffer{},
		Text:         &strings.Builder{},
		address:      a.address,
		baseSet:      a.baseSet,
	}
}

func (a *Emitter) Append(e *Emitter) {
	a.address = e.address
	a.baseSet = e.baseSet
	a.flagsTracker = e.flagsTracker

	_, _ = e.Code.WriteTo(a.Code)
	_, _ = a.Text.WriteString(e.Text.String())
}

func (a *Emitter) SetBase(addr uint32) {
	a.address = addr
	a.baseSet = true
}

func (a *Emitter) GetBase() uint32 {
	return a.address
}

func (a *Emitter) emitBase() {
	if a.Text == nil {
		return
	}
	if !a.baseSet {
		return
	}

	_, _ = a.Text.WriteString(fmt.Sprintf("base $%06x\n", a.address))
	a.baseSet = false
}

func (a *Emitter) emit1(ins string, d [1]byte) {
	if a.Code != nil {
		_, _ = a.Code.Write(d[:])
	}
	if a.Text != nil {
		a.emitBase()
		// TODO: adjust these format widths
		_, _ = a.Text.WriteString(fmt.Sprintf("    %-5s %-8s ; $%06x  %02x\n", ins, "", a.address, d[0]))
	}
	a.address += 1
}

func (a *Emitter) emit2(ins, argsFormat string, d [2]byte) {
	if a.Code != nil {
		_, _ = a.Code.Write(d[:])
	}
	if a.Text != nil {
		a.emitBase()
		args := fmt.Sprintf(argsFormat, d[1])
		// TODO: adjust these format widths
		_, _ = a.Text.WriteString(fmt.Sprintf("    %-5s %-8s ; $%06x  %02x %02x\n", ins, args, a.address, d[0], d[1]))
	}
	a.address += 2
}

func (a *Emitter) emit3(ins, argsFormat string, d [3]byte) {
	if a.Code != nil {
		_, _ = a.Code.Write(d[:])
	}
	if a.Text != nil {
		a.emitBase()
		args := fmt.Sprintf(argsFormat, d[1], d[2])
		// TODO: adjust these format widths
		_, _ = a.Text.WriteString(fmt.Sprintf("    %-5s %-8s ; $%06x  %02x %02x %02x\n", ins, args, a.address, d[0], d[1], d[2]))
	}
	a.address += 3
}

func (a *Emitter) emit4(ins, argsFormat string, d [4]byte) {
	if a.Code != nil {
		_, _ = a.Code.Write(d[:])
	}
	if a.Text != nil {
		a.emitBase()
		args := fmt.Sprintf(argsFormat, d[1], d[2], d[3])
		// TODO: adjust these format widths
		_, _ = a.Text.WriteString(fmt.Sprintf("    %-5s %-8s ; $%06x  %02x %02x %02x %02x\n", ins, args, a.address, d[0], d[1], d[2], d[3]))
	}
	a.address += 4
}

func imm24(v uint32) (byte, byte, byte) {
	return byte(v), byte(v >> 8), byte(v >> 16)
}

func imm16(v uint16) (byte, byte) {
	return byte(v), byte(v >> 8)
}

func (a *Emitter) Comment(s string) {
	if a.Text != nil {
		a.emitBase()
		_, _ = a.Text.WriteString(fmt.Sprintf("    ; %s\n", s))
	}
}

const hextable = "0123456789abcdef"

func (a *Emitter) EmitBytes(b []byte) {
	if a.Text != nil {
		a.emitBase()
		s := strings.Builder{}
		blen := len(b)
		for i, v := range b {
			s.Write([]byte{'$', hextable[(v>>4)&0xF], hextable[v&0xF]})
			if i < blen-1 {
				s.Write([]byte{',', ' '})
			}
		}
		_, _ = fmt.Fprintf(a.Text, "; $%06x\n", a.address)
		_, _ = fmt.Fprintf(a.Text, "    %-5s %s\n", "db", s.String())
	}
	if a.Code != nil {
		_, _ = a.Code.Write(b)
	}
	a.address += uint32(len(b))
}

func (a *Emitter) REP(c Flags) {
	a.AssumeREP(c)
	a.emit2("rep", "#$%02x", [2]byte{0xC2, byte(c)})
}

func (a *Emitter) SEP(c Flags) {
	a.AssumeSEP(c)
	a.emit2("sep", "#$%02x", [2]byte{0xE2, byte(c)})
}

func (a *Emitter) NOP() {
	a.emit1("nop", [1]byte{0xEA})
}

func (a *Emitter) JSR_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x20
	d[1], d[2] = imm16(addr)
	a.emit3("jsr", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) JSL(addr uint32) {
	var d [4]byte
	d[0] = 0x22
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("jsl", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) JSL_lhb(lo, hi, bank uint8) {
	var d [4]byte
	d[0] = 0x22
	d[1], d[2], d[3] = lo, hi, bank
	a.emit4("jsl", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) JML(addr uint32) {
	var d [4]byte
	d[0] = 0x5C
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("jml", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) RTS() {
	a.emit1("rts", [1]byte{0x60})
}

func (a *Emitter) RTL() {
	a.emit1("rtl", [1]byte{0x6B})
}

func (a *Emitter) LDA_imm8_b(m uint8) {
	if a.IsM16bit() {
		panic(fmt.Errorf("asm: LDA_imm8_b called but 'm' flag is 16-bit; call SEP(0x20) or AssumeSEP(0x20) first"))
	}
	var d [2]byte
	d[0] = 0xA9
	d[1] = m
	a.emit2("lda.b", "#$%02x", d)
}

func (a *Emitter) LDA_imm16_w(m uint16) {
	if !a.IsM16bit() {
		panic(fmt.Errorf("asm: LDA_imm16_w called but 'm' flag is 8-bit; call REP(0x20) or AssumeREP(0x20) first"))
	}
	var d [3]byte
	d[0] = 0xA9
	d[1], d[2] = imm16(m)
	a.emit3("lda.w", "#$%02[2]x%02[1]x", d)
}

func (a *Emitter) LDA_imm16_lh(lo, hi uint8) {
	if !a.IsM16bit() {
		panic(fmt.Errorf("asm: LDA_imm16_lh called but 'm' flag is 8-bit; call REP(0x20) or AssumeREP(0x20) first"))
	}
	var d [3]byte
	d[0] = 0xA9
	d[1], d[2] = lo, hi
	a.emit3("lda.w", "#$%02[2]x%02[1]x", d)
}

func (a *Emitter) LDA_long(addr uint32) {
	var d [4]byte
	d[0] = 0xAF
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("lda.l", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) LDA_abs(addr uint16) {
	var d [3]byte
	d[0] = 0xAD
	d[1], d[2] = imm16(addr)
	a.emit3("lda.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) LDA_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0xBD
	d[1], d[2] = imm16(addr)
	a.emit3("lda.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) STA_long(addr uint32) {
	var d [4]byte
	d[0] = 0x8F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("sta.l", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) STA_abs(addr uint16) {
	var d [3]byte
	d[0] = 0x8D
	d[1], d[2] = imm16(addr)
	a.emit3("sta.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) STA_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x9D
	d[1], d[2] = imm16(addr)
	a.emit3("sta.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) STA_dp(addr uint8) {
	var d [2]byte
	d[0] = 0x85
	d[1] = addr
	a.emit2("sta.b", "$%02[1]x", d)
}

func (a *Emitter) ORA_long(addr uint32) {
	var d [4]byte
	d[0] = 0x0F
	d[1], d[2], d[3] = imm24(addr)
	a.emit4("ora.l", "$%02[3]x%02[2]x%02[1]x", d)
}

func (a *Emitter) ORA_imm8_b(m uint8) {
	if a.IsM16bit() {
		panic(fmt.Errorf("asm: ORA_imm8_b called but 'm' flag is 16-bit; call SEP(0x20) or AssumeSEP(0x20) first"))
	}
	var d [2]byte
	d[0] = 0x09
	d[1] = m
	a.emit2("ora.b", "#$%02x", d)
}

func (a *Emitter) CMP_imm8_b(m uint8) {
	if a.IsM16bit() {
		panic(fmt.Errorf("asm: CMP_imm8_b called but 'm' flag is 16-bit; call SEP(0x20) or AssumeSEP(0x20) first"))
	}
	var d [2]byte
	d[0] = 0xC9
	d[1] = m
	a.emit2("cmp.b", "#$%02x", d)
}

func (a *Emitter) BNE(m int8) {
	var d [2]byte
	d[0] = 0xD0
	d[1] = uint8(m)
	a.emit2("bne", "$%02x", d)
}

func (a *Emitter) BEQ(m int8) {
	var d [2]byte
	d[0] = 0xF0
	d[1] = uint8(m)
	a.emit2("beq", "$%02x", d)
}

func (a *Emitter) BPL(m int8) {
	var d [2]byte
	d[0] = 0x10
	d[1] = uint8(m)
	a.emit2("bpl", "$%02x", d)
}

func (a *Emitter) BRA(m int8) {
	var d [2]byte
	d[0] = 0x80
	d[1] = uint8(m)
	a.emit2("bra", "$%02x", d)
}

func (a *Emitter) ADC_imm8_b(m uint8) {
	if a.IsM16bit() {
		panic(fmt.Errorf("asm: ADC_imm8_b called but 'm' flag is 16-bit; call SEP(0x20) or AssumeSEP(0x20) first"))
	}
	var d [2]byte
	d[0] = 0x69
	d[1] = m
	a.emit2("adc.b", "#$%02x", d)
}

func (a *Emitter) CPY_imm8_b(m uint8) {
	if a.IsX16bit() {
		panic(fmt.Errorf("asm: CPY_imm8_b called but 'x' flag is 16-bit; call SEP(0x10) or AssumeSEP(0x10) first"))
	}
	var d [2]byte
	d[0] = 0xC0
	d[1] = m
	a.emit2("cpy.b", "#$%02x", d)
}

func (a *Emitter) LDY_abs(offs uint16) {
	var d [3]byte
	d[0] = 0xAC
	d[1], d[2] = imm16(offs)
	a.emit3("ldy.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) STZ_abs(offs uint16) {
	var d [3]byte
	d[0] = 0x9C
	d[1], d[2] = imm16(offs)
	a.emit3("stz.w", "$%02[2]x%02[1]x", d)
}

func (a *Emitter) STZ_abs_x(addr uint16) {
	var d [3]byte
	d[0] = 0x9E
	d[1], d[2] = imm16(addr)
	a.emit3("stz.w", "$%02[2]x%02[1]x,X", d)
}

func (a *Emitter) INC_dp(addr uint8) {
	var d [2]byte
	d[0] = 0xE6
	d[1] = addr
	a.emit2("inc.b", "$%02[1]x", d)
}

func (a *Emitter) LDA_dp(addr uint8) {
	var d [2]byte
	d[0] = 0xA5
	d[1] = addr
	a.emit2("lda.b", "$%02[1]x", d)
}

func (a *Emitter) LDX_imm8_b(m uint8) {
	if a.IsX16bit() {
		panic(fmt.Errorf("asm: LDX_imm8_b called but 'x' flag is 16-bit; call SEP(0x10) or AssumeSEP(0x10) first"))
	}
	var d [2]byte
	d[0] = 0xA2
	d[1] = m
	a.emit2("ldx.b", "#$%02x", d)
}

func (a *Emitter) DEX() {
	a.emit1("dex", [1]byte{0xCA})
}

func (a *Emitter) DEY() {
	a.emit1("dey", [1]byte{0x88})
}

func (a *Emitter) AND_imm8_b(m uint8) {
	if a.IsM16bit() {
		panic(fmt.Errorf("asm: AND_imm8_b called but 'm' flag is 16-bit; call SEP(0x20) or AssumeSEP(0x20) first"))
	}
	var d [2]byte
	d[0] = 0x29
	d[1] = m
	a.emit2("and.b", "#$%02x", d)
}

func (a *Emitter) PHB() {
	a.emit1("phb", [1]byte{0x8B})
}

func (a *Emitter) PHA() {
	a.emit1("pha", [1]byte{0x48})
}

func (a *Emitter) PHX() {
	a.emit1("phx", [1]byte{0xDA})
}

func (a *Emitter) PHY() {
	a.emit1("phy", [1]byte{0x5A})
}

func (a *Emitter) PLY() {
	a.emit1("ply", [1]byte{0x7A})
}

func (a *Emitter) PLX() {
	a.emit1("plx", [1]byte{0xFA})
}

func (a *Emitter) PLA() {
	a.emit1("pla", [1]byte{0x68})
}

func (a *Emitter) PLB() {
	a.emit1("plb", [1]byte{0xAB})
}

func (a *Emitter) LDX_imm16_w(m uint16) {
	if !a.IsX16bit() {
		panic(fmt.Errorf("asm: LDA_imm16_w called but 'x' flag is 8-bit; call REP(0x10) or AssumeREP(0x10) first"))
	}
	var d [3]byte
	d[0] = 0xA2
	d[1], d[2] = imm16(m)
	a.emit3("ldx.w", "#$%02[2]x%02[1]x", d)
}

func (a *Emitter) LDY_imm16_w(m uint16) {
	if !a.IsX16bit() {
		panic(fmt.Errorf("asm: LDA_imm16_w called but 'x' flag is 8-bit; call REP(0x10) or AssumeREP(0x10) first"))
	}
	var d [3]byte
	d[0] = 0xA0
	d[1], d[2] = imm16(m)
	a.emit3("ldy.w", "#$%02[2]x%02[1]x", d)
}

func (a *Emitter) MVN(sourceBank uint8, destinationBank uint8) {
	var d [3]byte
	d[0] = 0x54
	d[1], d[2] = sourceBank, destinationBank
	a.emit3("mvn", "$%02[1]x,$%02[2]x", d)
}

func (a *Emitter) JMP_indirect(addr uint16) {
	var d [3]byte
	d[0] = 0x6C
	d[1], d[2] = imm16(addr)
	a.emit3("jmp", "($%02[2]x%02[1]x)", d)
}
