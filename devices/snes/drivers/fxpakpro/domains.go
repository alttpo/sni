package fxpakpro

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sni/devices"
	"sni/protos/sni"
	"strings"
)

type snesDomain struct {
	name      string
	isExposed bool

	notes string

	isReadable  bool
	isWriteable bool

	start uint32
	size  uint32
}

var allDomains []snesDomain
var domainByName map[string]*snesDomain

func (d *Device) MemoryDomains(_ context.Context, request *sni.MemoryDomainsRequest) (rsp *sni.MemoryDomainsResponse, err error) {
	domains := make([]*sni.MemoryDomain, len(allDomains))
	for i := range allDomains {
		domains[i] = &sni.MemoryDomain{
			Name:      allDomains[i].name,
			Exposed:   allDomains[i].isExposed,
			Readable:  allDomains[i].isReadable,
			Writeable: allDomains[i].isWriteable,
		}

		if allDomains[i].notes != "" {
			domains[i].Notes = &allDomains[i].notes
		}

		size := uint64(allDomains[i].size)
		if size > 0 {
			domains[i].Size = new(uint64)
			*domains[i].Size = size
		}
	}

	rsp = &sni.MemoryDomainsResponse{
		Uri:     request.Uri,
		Domains: domains,
	}

	return
}

func (d *Device) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (rsp *sni.MultiDomainReadResponse, err error) {
	devReads := make([]devices.MemoryReadRequest, 0, 8)
	readResponses := make([]*sni.MemoryDomainOffsetData, 0, 8)
	rsp = &sni.MultiDomainReadResponse{
		Uri:       request.Uri,
		Responses: make([]*sni.GroupedDomainReadResponses, len(request.Requests)),
	}

	for i, requests := range request.Requests {
		domainName := requests.Name
		domainNameLower := strings.ToLower(domainName)

		var domain *snesDomain
		var ok bool
		domain, ok = domainByName[domainNameLower]
		if !ok {
			rsp = nil
			err = status.Errorf(codes.InvalidArgument, "fxpakpro: unrecognized domain name '%s'", domainName)
			return
		}

		rsp.Responses[i] = &sni.GroupedDomainReadResponses{
			Name:  requests.Name,
			Reads: make([]*sni.MemoryDomainOffsetData, len(requests.Reads)),
		}

		for j, read := range requests.Reads {
			// validate offset and size pair:
			if read.Offset >= uint64(domain.size) {
				rsp = nil
				err = status.Errorf(codes.InvalidArgument, "fxpakpro: read of domain '%s', offset %d would exceed domain size %d", domainName, read.Offset, domain.size)
				return
			}
			if read.Offset+read.Size > uint64(domain.size) {
				rsp = nil
				err = status.Errorf(codes.InvalidArgument, "fxpakpro: read of domain '%s', offset + size = %d would exceed domain size %d", domainName, read.Offset+read.Size, domain.size)
				return
			}

			rsp.Responses[i].Reads[j] = &sni.MemoryDomainOffsetData{
				Offset: read.Offset,
				Data:   nil,
			}
			readResponses = append(readResponses, rsp.Responses[i].Reads[j])

			devReads = append(devReads, devices.MemoryReadRequest{
				RequestAddress: devices.AddressTuple{
					Address:       domain.start + uint32(read.Offset),
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_Unknown,
				},
				Size: int(read.Size),
			})
		}
	}

	var mrsp []devices.MemoryReadResponse
	mrsp, err = d.MultiReadMemory(ctx, devReads...)
	if err != nil {
		rsp = nil
		return
	}

	for i := range mrsp {
		readResponses[i].Data = mrsp[i].Data
	}

	return
}

func (d *Device) MultiDomainWrite(ctx context.Context, request *sni.MultiDomainWriteRequest) (rsp *sni.MultiDomainWriteResponse, err error) {
	//mreqs := make([]devices.MemoryWriteRequest, 0)
	//rsp = &sni.MultiDomainWriteResponse{
	//	Responses: make([]*sni.GroupedDomainWriteResponses, len(request.Requests)),
	//}
	//addressSizes := make([]*sni.MemoryDomainOffsetSize, 0)
	//
	//for i, domainReqs := range request.Requests {
	//	snesDomainRef, ok := domainReqs.Domain.Type.(*sni.MemoryDomainRef_Snes)
	//	if !ok {
	//		err = status.Errorf(codes.InvalidArgument, "Domain.Type must be of SNES type")
	//		return
	//	}
	//
	//	var dm *domainDesc
	//	if snesDomainRef.Snes == sni.MemoryDomainTypeSNES_SNESCoreSpecificMemory {
	//		// look up by string name instead:
	//		if domainReqs.Name == nil {
	//			err = status.Error(codes.InvalidArgument, "domain name must be non-nil when using SNESCoreSpecificMemory type")
	//			return
	//		}
	//
	//		domainName := strings.ToUpper(*domainReqs.Domain.Name)
	//		dm, ok = domainDescByName[domainName]
	//		if !ok {
	//			err = status.Errorf(codes.InvalidArgument, "invalid domain name '%s'", domainName)
	//			return
	//		}
	//	} else {
	//		dm, ok = domainDescByType[snesDomainRef.Snes]
	//		if !ok {
	//			err = status.Errorf(codes.InvalidArgument, "invalid domain type '%s'", snesDomainRef.Snes)
	//			return
	//		}
	//	}
	//
	//	rsp.Responses[i] = &sni.GroupedDomainWriteResponses{
	//		Domain: dm.domainRef,
	//		Writes: make([]*sni.MemoryDomainOffsetSize, len(domainReqs.Writes)),
	//	}
	//	for j, write := range domainReqs.Writes {
	//		if write.Address >= dm.size {
	//			err = status.Errorf(codes.InvalidArgument, "request start address 0x%06x exceeds domain %s size 0x%06x", write.Address, *dm.domainRef.Name, dm.size)
	//			return
	//		}
	//		size := uint32(len(write.Data))
	//		if write.Address+size > dm.size {
	//			err = status.Errorf(codes.InvalidArgument, "request end address 0x%06x exceeds domain %s size 0x%06x", write.Address+size, *dm.domainRef.Name, dm.size)
	//			return
	//		}
	//
	//		mreq := devices.MemoryWriteRequest{
	//			RequestAddress: devices.AddressTuple{
	//				Address:       dm.start + write.Address,
	//				AddressSpace:  sni.AddressSpace_FxPakPro,
	//				MemoryMapping: sni.MemoryMapping_Unknown,
	//			},
	//			Data: write.Data,
	//		}
	//		mreqs = append(mreqs, mreq)
	//
	//		addressSize := &sni.MemoryDomainOffsetSize{
	//			Offset: write.Address,
	//			Size:   size,
	//		}
	//		addressSizes = append(addressSizes, addressSize)
	//		rsp.Responses[i].Writes[j] = addressSize
	//	}
	//}
	//
	//var mrsp []devices.MemoryWriteResponse
	//mrsp, err = d.MultiWriteMemory(ctx, mreqs...)
	//if err != nil {
	//	return
	//}
	//
	//_ = mrsp

	return
}
