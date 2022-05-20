package fxpakpro

import (
	"context"
	"sni/protos/sni"
)

func (d *Device) MemoryDomains(ctx context.Context, request *sni.MemoryDomainsRequest) (*sni.MemoryDomainsResponse, error) {
	return &sni.MemoryDomainsResponse{
		Uri: "",
		Domains: []*sni.MemoryDomain{
			{
				DomainName: "WRAM",
				// $F5_0000 .. $F7_0000
				Size: 0x2_0000,
			},
			{
				DomainName: "SRAM",
				// $E0_0000 .. $F0_0000
				Size: 0x10_0000,
				// TODO: how to determine size of current SRAM?
			},
			{
				DomainName: "ROM",
				Size:       0xE0_0000,
			},

			// TODO: CPUREG, PPUREG, APUREG, VRAM, ARAM, IRAM, etc.
		},
	}, nil
}

func (d *Device) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (*sni.MultiDomainReadResponse, error) {
	//TODO implement me
	panic("implement me")
}

func (d *Device) MultiDomainWrite(ctx context.Context, request *sni.MultiDomainWriteRequest) (*sni.MultiDomainWriteResponse, error) {
	//TODO implement me
	panic("implement me")
}
