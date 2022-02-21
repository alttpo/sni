package devices

import (
	"fmt"
	"sni/protos/sni"
)

type AddressTuple struct {
	Address uint32
	sni.AddressSpace
	sni.MemoryMapping
}

func (a *AddressTuple) String() string {
	return fmt.Sprintf("(%s $%06x %s)",
		sni.AddressSpace_name[int32(a.AddressSpace)],
		a.Address,
		a.MemoryMapping,
	)
}
