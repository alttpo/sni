package fxpakpro

import (
	"context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"sni/devices"
	"sni/devices/platforms"
	"sni/protos/sni"
	"strings"
)

type snesDomain struct {
	platforms.Domain

	// space must be either "SNES" or "CMD"
	space string
	// start address in the fx pak pro address space:
	start uint32
}

var allDomains []snesDomain
var domainByName map[string]*snesDomain

func (d *Device) MemoryDomains(_ context.Context, request *sni.MemoryDomainsRequest) (rsp *sni.MemoryDomainsResponse, err error) {
	domains := make([]*sni.MemoryDomain, len(allDomains))
	for i := range allDomains {
		domains[i] = &sni.MemoryDomain{
			Name:           allDomains[i].Name,
			IsExposed:      allDomains[i].IsExposed,
			IsCoreSpecific: allDomains[i].IsCoreSpecific,
			IsReadable:     allDomains[i].IsReadable,
			IsWriteable:    allDomains[i].IsWriteable,
			Size:           allDomains[i].Size,
		}
	}

	rsp = &sni.MemoryDomainsResponse{
		Uri:      request.Uri,
		CoreName: driverName,
		Domains:  domains,
	}

	return
}

func (d *Device) MultiDomainRead(ctx context.Context, request *sni.MultiDomainReadRequest) (rsp *sni.MultiDomainReadResponse, err error) {
	mreqs := make([]devices.MemoryReadRequest, 0, 8)
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
			if read.Offset >= domain.Size {
				rsp = nil
				err = status.Errorf(codes.InvalidArgument, "fxpakpro: read of domain '%s', offset %d would exceed domain size %d", domainName, read.Offset, domain.Size)
				return
			}
			if read.Offset+read.Size > domain.Size {
				rsp = nil
				err = status.Errorf(codes.InvalidArgument, "fxpakpro: read of domain '%s', offset + size = %d would exceed domain size %d", domainName, read.Offset+read.Size, domain.Size)
				return
			}

			rsp.Responses[i].Reads[j] = &sni.MemoryDomainOffsetData{
				Offset: read.Offset,
				Data:   nil,
			}
			readResponses = append(readResponses, rsp.Responses[i].Reads[j])

			address := domain.start + uint32(read.Offset)
			mreqs = append(mreqs, devices.MemoryReadRequest{
				RequestAddress: devices.AddressTuple{
					Address:       address,
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_Unknown,
				},
				Size: int(read.Size),
			})
		}
	}

	var mrsp []devices.MemoryReadResponse
	mrsp, err = d.MultiReadMemory(ctx, mreqs...)
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
	mreqs := make([]devices.MemoryWriteRequest, 0)
	rsp = &sni.MultiDomainWriteResponse{
		Responses: make([]*sni.GroupedDomainWriteResponses, len(request.Requests)),
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

		rsp.Responses[i] = &sni.GroupedDomainWriteResponses{
			Name:   requests.Name,
			Writes: make([]*sni.MemoryDomainOffsetSize, len(requests.Writes)),
		}

		for j, write := range requests.Writes {
			// validate offset and size pair:
			size := uint64(len(write.Data))
			if write.Offset >= domain.Size {
				rsp = nil
				err = status.Errorf(codes.InvalidArgument, "fxpakpro: read of domain '%s', offset %d would exceed domain size %d", domainName, write.Offset, domain.Size)
				return
			}
			if write.Offset+size > domain.Size {
				rsp = nil
				err = status.Errorf(codes.InvalidArgument, "fxpakpro: read of domain '%s', offset + size = %d would exceed domain size %d", domainName, write.Offset+size, domain.Size)
				return
			}

			rsp.Responses[i].Writes[j] = &sni.MemoryDomainOffsetSize{
				Offset: write.Offset,
				Size:   size,
			}

			mreqs = append(mreqs, devices.MemoryWriteRequest{
				RequestAddress: devices.AddressTuple{
					Address:       domain.start + uint32(write.Offset),
					AddressSpace:  sni.AddressSpace_FxPakPro,
					MemoryMapping: sni.MemoryMapping_Unknown,
				},
				Data: write.Data,
			})
		}
	}

	_, err = d.MultiWriteMemory(ctx, mreqs...)
	if err != nil {
		rsp = nil
		return
	}

	return
}
