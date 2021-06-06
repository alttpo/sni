package fxpakpro

import (
	"fmt"
	"sni/snes"
)

type vput struct {
	batch []snes.Write
}

func (q *Queue) newVPUT(batch []snes.Write) *vput {
	return &vput{batch: batch}
}

// Command interface:
func (c *vput) Execute(queue snes.Queue, keepAlive snes.KeepAlive) error {
	f := queue.(*Queue).f

	reqs := c.batch
	if len(reqs) > 8 {
		return fmt.Errorf("vput: cannot have more than 8 requests in batch")
	}

	sb := make([]byte, 64)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpVPUT)
	sb[5] = byte(SpaceSNES)
	sb[6] = byte(FlagDATA64B | FlagNORESP)

	total := 0
	for i := 0; i < len(reqs); i++ {
		// 4-byte struct: 1 byte size, 3 byte address
		sb[32+(i*4)] = reqs[i].Size
		sb[33+(i*4)] = byte((reqs[i].Address >> 16) & 0xFF)
		sb[34+(i*4)] = byte((reqs[i].Address >> 8) & 0xFF)
		sb[35+(i*4)] = byte((reqs[i].Address >> 0) & 0xFF)
		total += int(reqs[i].Size)
	}

	err := sendSerial(f, sb)
	if err != nil {
		return err
	}

	// calculate expected number of packets:
	packets := total / 64
	remainder := total & 63
	if remainder > 0 {
		packets++
	}

	// concatenate all accompanying data together in one large slice:
	expected := packets * 64
	whole := make([]byte, expected)
	o := 0
	for i := 0; i < len(reqs); i++ {
		copy(whole[o:], reqs[i].Data)
		o += len(reqs[i].Data)
	}

	// send the expected number of 64-byte packets:
	err = sendSerial(f, whole)
	if err != nil {
		return err
	}

	// make completed callbacks:
	for i := 0; i < len(reqs); i++ {
		// make response callback:
		completed := reqs[i].Completion
		if completed != nil {
			completed(snes.Response{
				IsWrite: true,
				Address: reqs[i].Address,
				Size:    reqs[i].Size,
				Data:    reqs[i].Data,
				Extra:   reqs[i].Extra,
			})
		}
	}

	return nil
}
