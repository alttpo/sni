package fxpakpro

import (
	"context"
	"sni/protos/sni"
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
	rsp = &sni.MemoryDomainsResponse{
		Uri: request.Uri,
		//Domains: domains,
	}

	return
}

func (d *Device) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (rsp *sni.MultiDomainReadResponse, err error) {
	//mreqs := make([]devices.MemoryReadRequest, 0)
	//rsp = &sni.MultiDomainReadResponse{
	//	Responses: make([]*sni.GroupedDomainReadResponses, len(request.Requests)),
	//}
	//addressDatas := make([]*sni.MemoryDomainOffsetData, 0)
	//
	//for i, domainReqs := range request.Requests {
	//	rsp.Responses[i] = &sni.GroupedDomainReadResponses{
	//		Name:  domainReqs.Name,
	//		Reads: make([]*sni.MemoryDomainOffsetData, len(domainReqs.Reads)),
	//	}
	//	for j, read := range domainReqs.Reads {
	//		mreq := devices.MemoryReadRequest{
	//			RequestAddress: devices.AddressTuple{
	//				Address:       dm.start + read.Address,
	//				AddressSpace:  sni.AddressSpace_FxPakPro,
	//				MemoryMapping: sni.MemoryMapping_Unknown,
	//			},
	//			Size: int(read.Size),
	//		}
	//		mreqs = append(mreqs, mreq)
	//
	//		addressData := &sni.MemoryDomainOffsetData{
	//			Offset: read.Offset,
	//			Data:   nil,
	//		}
	//		addressDatas = append(addressDatas, addressData)
	//		rsp.Responses[i].Reads[j] = addressData
	//	}
	//}
	//
	//var mrsp []devices.MemoryReadResponse
	//mrsp, err = d.MultiReadMemory(ctx, mreqs...)
	//if err != nil {
	//	return
	//}
	//
	//for k := range mrsp {
	//	// update the `GroupedDomainReadResponses_AddressData`s across the groupings:
	//	addressDatas[k].Data = mrsp[k].Data
	//}

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
