package fxpakpro

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sni/devices"
	"sni/protos/sni"
	"strings"
)

type domain struct {
	domainRef *sni.MemoryDomainRef
	start     uint32
	size      uint32
	writeable bool
}

func s(v string) *string {
	return &v
}

var domainRefs = [...]*sni.MemoryDomainRef_Snes{
	{Snes: sni.MemoryDomainTypeSnes_SNESCustomName},
	{Snes: sni.MemoryDomainTypeSnes_SNESCartROM},
	{Snes: sni.MemoryDomainTypeSnes_SNESCartSRAM},
	{Snes: sni.MemoryDomainTypeSnes_SNESWorkRAM},
	{Snes: sni.MemoryDomainTypeSnes_SNESVRAM},
	{Snes: sni.MemoryDomainTypeSnes_SNESAPURAM},
	{Snes: sni.MemoryDomainTypeSnes_SNESCGRAM},
	{Snes: sni.MemoryDomainTypeSnes_SNESOAM},
}

var domainDescs = [...]domain{
	{&sni.MemoryDomainRef{Name: s("CARTROM"), Type: domainRefs[1]}, 0x00_0000, 0xE0_0000, true},
	{&sni.MemoryDomainRef{Name: s("SRAM"), Type: domainRefs[2]}, 0xE0_0000, 0x10_0000, true},
	// TODO: convert WRAM writes to asm generation and nmi exe feature
	{&sni.MemoryDomainRef{Name: s("WRAM"), Type: domainRefs[3]}, 0xF5_0000, 0x02_0000, true},
	{&sni.MemoryDomainRef{Name: s("VRAM"), Type: domainRefs[4]}, 0xF7_0000, 0x01_0000, false},
	{&sni.MemoryDomainRef{Name: s("APURAM"), Type: domainRefs[5]}, 0xF8_0000, 0x01_0000, false},
	{&sni.MemoryDomainRef{Name: s("CGRAM"), Type: domainRefs[6]}, 0xF9_0000, 0x0200, false},
	{&sni.MemoryDomainRef{Name: s("OAM"), Type: domainRefs[7]}, 0xF9_0200, 0x0420 - 0x0200, false},
	{&sni.MemoryDomainRef{Name: s("MISC"), Type: domainRefs[0]}, 0xF9_0420, 0x0500 - 0x0420, false},
	{&sni.MemoryDomainRef{Name: s("PPUREG"), Type: domainRefs[0]}, 0xF9_0500, 0x0700 - 0x0500, false},
	{&sni.MemoryDomainRef{Name: s("CPUREG"), Type: domainRefs[0]}, 0xF9_0700, 0x0200, false},
}
var domains []*sni.MemoryDomain

func init() {
	domains = make([]*sni.MemoryDomain, 0, len(domainDescs))

	for _, d := range domainDescs {
		domains = append(domains, &sni.MemoryDomain{
			Domain:    d.domainRef,
			Size:      d.size,
			Readable:  true,
			Writeable: d.writeable,
		})
	}
}

func (d *Device) MemoryDomains(ctx context.Context, request *sni.MemoryDomainsRequest) (rsp *sni.MemoryDomainsResponse, err error) {
	rsp = &sni.MemoryDomainsResponse{
		Domains: domains,
	}

	return
}

func (d *Device) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (rsp *sni.MultiDomainReadResponse, err error) {
	mreqs := make([]devices.MemoryReadRequest, 0)
	rsp = &sni.MultiDomainReadResponse{
		Responses: make([]*sni.GroupedDomainReadResponses, len(request.Requests)),
	}
	addressDatas := make([]*sni.MemoryDomainAddressData, 0)

	for i, domainReqs := range request.Requests {
		domainName := strings.ToUpper(domainReqs.DomainName)
		dm, ok := domainMap[domainName]
		if !ok {
			err = status.Errorf(codes.InvalidArgument, "invalid domain name '%s'", domainName)
			return
		}

		rsp.Responses[i] = &sni.GroupedDomainReadResponses{
			DomainName: domainName,
			Reads:      make([]*sni.MemoryDomainAddressData, len(domainReqs.Reads)),
		}
		for j, read := range domainReqs.Reads {
			if read.Address >= dm.size {
				err = status.Errorf(codes.InvalidArgument, "request start address 0x%06x exceeds domain %s size 0x%06x", read.Address, domainName, dm.size)
				return
			}
			if read.Address+read.Size > dm.size {
				err = status.Errorf(codes.InvalidArgument, "request end address 0x%06x exceeds domain %s size 0x%06x", read.Address+read.Size, domainName, dm.size)
				return
			}

			mreq := devices.MemoryReadRequest{
				RequestAddress: devices.AddressTuple{
					Address:       dm.start + read.Address,
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_Unknown,
				},
				Size: int(read.Size),
			}
			mreqs = append(mreqs, mreq)

			addressData := &sni.GroupedDomainReadResponses_AddressData{
				Address: mreq.RequestAddress.Address,
				Data:    nil,
			}
			addressDatas = append(addressDatas, addressData)
			rsp.Responses[i].Reads[j] = addressData
		}
	}

	var mrsp []devices.MemoryReadResponse
	mrsp, err = d.MultiReadMemory(ctx, mreqs...)
	if err != nil {
		return
	}

	for k := range mrsp {
		if addressDatas[k].Address != mrsp[k].RequestAddress.Address {
			err = status.Errorf(codes.Internal, "internal consistency error aligning read response with request")
		}
		// update the `GroupedDomainReadResponses_AddressData`s across the groupings:
		addressDatas[k].Data = mrsp[k].Data
	}

	return
}

func (d *Device) MultiDomainWrite(ctx context.Context, request *sni.MultiDomainWriteRequest) (*sni.MultiDomainWriteResponse, error) {
	//TODO implement me
	panic("implement me")
}
