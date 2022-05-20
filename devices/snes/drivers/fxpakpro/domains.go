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
	start uint32
	size  uint32
}

var domains = map[string]domain{
	"CARTROM": {0x00_0000, 0xE0_0000},
	"SRAM":    {0xE0_0000, 0x10_0000},
	"WRAM":    {0xF5_0000, 0x02_0000},
	"VRAM":    {0xF7_0000, 0x01_0000},
	"APURAM":  {0xF8_0000, 0x01_0000},
	"CGRAM":   {0xF9_0000, 0x0200},
	"OAM":     {0xF9_0200, 0x0420 - 0x0200},
	"MISC":    {0xF9_0420, 0x0500 - 0x0420},
	"PPUREG":  {0xF9_0500, 0x0700 - 0x0500},
	"CPUREG":  {0xF9_0700, 0x0200},
}

func (d *Device) MemoryDomains(ctx context.Context, request *sni.MemoryDomainsRequest) (rsp *sni.MemoryDomainsResponse, err error) {
	rsp = &sni.MemoryDomainsResponse{
		Domains: make([]*sni.MemoryDomain, 0, len(domains)),
	}

	for name, d := range domains {
		rsp.Domains = append(rsp.Domains, &sni.MemoryDomain{
			DomainName: name,
			Size:       d.size,
		})
	}

	return
}

func (d *Device) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (rsp *sni.MultiDomainReadResponse, err error) {
	mreqs := make([]devices.MemoryReadRequest, 0)
	rsp = &sni.MultiDomainReadResponse{
		Responses: make([]*sni.GroupedDomainReadResponses, len(request.Requests)),
	}
	addressDatas := make([]*sni.GroupedDomainReadResponses_AddressData, 0)

	for i, domainReqs := range request.Requests {
		domainName := strings.ToUpper(domainReqs.DomainName)
		dm, ok := domains[domainName]
		if !ok {
			err = status.Errorf(codes.InvalidArgument, "invalid domain name '%s'", domainName)
			return
		}

		rsp.Responses[i] = &sni.GroupedDomainReadResponses{
			DomainName: domainName,
			Reads:      make([]*sni.GroupedDomainReadResponses_AddressData, len(domainReqs.Reads)),
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
