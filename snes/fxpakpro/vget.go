package fxpakpro

import (
	"fmt"
	"sni/snes"
)

type vget struct {
	batch []snes.Read
}

func (q *Queue) newVGET(batch []snes.Read) *vget {
	return &vget{batch: batch}
}

// Command interface:
func (c *vget) Execute(queue snes.Queue, keepAlive snes.KeepAlive) error {
	f := queue.(*Queue).f

	reqs := c.batch
	if len(reqs) > 8 {
		return fmt.Errorf("vget: cannot have more than 8 requests in batch")
	}

	sb := make([]byte, 64)
	sb[0] = byte('U')
	sb[1] = byte('S')
	sb[2] = byte('B')
	sb[3] = byte('A')
	sb[4] = byte(OpVGET)
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

	// read the expected number of 64-byte packets:
	expected := packets * 64
	rsp := make([]byte, expected)
	err = recvSerial(f, rsp, expected)
	if err != nil {
		return err
	}

	// shrink down to exact size:
	rsp = rsp[0:total]

	// make completed callbacks:
	o := 0
	for i := 0; i < len(reqs); i++ {
		size := int(reqs[i].Size)

		// make response callback:
		completed := reqs[i].Completion
		if completed != nil {
			completed(snes.Response{
				IsWrite: false,
				Address: reqs[i].Address,
				Size:    reqs[i].Size,
				Extra:   reqs[i].Extra,
				Data:    rsp[o : o+size],
			})
		}

		o += size
	}

	return nil
}
