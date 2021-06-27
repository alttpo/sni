package main

import (
	"fmt"
	"sni/protos/sni"
	"sni/snes"
	"strings"
)

func (s *devicesService) MethodRequestString(method string, req interface{}) string {
	if req == nil {
		return "nil"
	}

	return fmt.Sprintf("%+v", req)
}

func (s *devicesService) MethodResponseString(method string, rsp interface{}) string {
	if rsp == nil {
		return "nil"
	}

	return fmt.Sprintf("%+v", rsp)
}

func ReadMemoryRequestString(m *sni.ReadMemoryRequest) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetRequestAddress(),
			AddressSpace:  m.GetRequestAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		m.GetSize(),
	)
}

func WriteMemoryRequestString(m *sni.WriteMemoryRequest) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetRequestAddress(),
			AddressSpace:  m.GetRequestAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		len(m.GetData()),
	)
}

func ReadMemoryResponseString(m *sni.ReadMemoryResponse) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetDeviceAddress(),
			AddressSpace:  m.GetDeviceAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		len(m.GetData()),
	)
}

func WriteMemoryResponseString(m *sni.WriteMemoryResponse) string {
	return fmt.Sprintf(
		"{address:%s,size:%#x}",
		&snes.AddressTuple{
			Address:       m.GetDeviceAddress(),
			AddressSpace:  m.GetDeviceAddressSpace(),
			MemoryMapping: m.GetRequestMemoryMapping(),
		},
		m.GetSize(),
	)
}

func (s *deviceMemoryService) MethodRequestString(method string, req interface{}) string {
	if req == nil {
		return "nil"
	}

	switch method {
	case "/DeviceMemory/SingleRead":
		srReq := req.(*sni.SingleReadMemoryRequest)
		return fmt.Sprintf("uri:\"%s\",request:%s", srReq.GetUri(), ReadMemoryRequestString(srReq.GetRequest()))
	case "/DeviceMemory/SingleWrite":
		swReq := req.(*sni.SingleWriteMemoryRequest)
		return fmt.Sprintf("uri:\"%s\",request:%s", swReq.GetUri(), WriteMemoryRequestString(swReq.GetRequest()))
	case "/DeviceMemory/MultiRead":
		mrReq := req.(*sni.MultiReadMemoryRequest)

		sb := strings.Builder{}
		for i, rReq := range mrReq.GetRequests() {
			sb.WriteString(ReadMemoryRequestString(rReq))
			if i != len(mrReq.GetRequests())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",requests:[%s]", mrReq.GetUri(), sb.String())
	case "/DeviceMemory/MultiWrite":
		mwReq := req.(*sni.MultiWriteMemoryRequest)

		sb := strings.Builder{}
		for i, wReq := range mwReq.GetRequests() {
			sb.WriteString(WriteMemoryRequestString(wReq))
			if i != len(mwReq.GetRequests())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",requests:[%s]", mwReq.GetUri(), sb.String())
	}

	return fmt.Sprintf("%+v", req)
}

func (s *deviceMemoryService) MethodResponseString(method string, rsp interface{}) string {
	if rsp == nil {
		return "nil"
	}

	switch method {
	case "/DeviceMemory/SingleRead":
		srReq := rsp.(*sni.SingleReadMemoryResponse)
		return fmt.Sprintf("uri:\"%s\",response:%s", srReq.GetUri(), ReadMemoryResponseString(srReq.GetResponse()))
	case "/DeviceMemory/SingleWrite":
		swReq := rsp.(*sni.SingleWriteMemoryResponse)
		return fmt.Sprintf("uri:\"%s\",response:%s", swReq.GetUri(), WriteMemoryResponseString(swReq.GetResponse()))
	case "/DeviceMemory/MultiRead":
		mrReq := rsp.(*sni.MultiReadMemoryResponse)

		sb := strings.Builder{}
		for i, rReq := range mrReq.GetResponses() {
			sb.WriteString(ReadMemoryResponseString(rReq))
			if i != len(mrReq.GetResponses())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",responses:[%s]", mrReq.GetUri(), sb.String())
	case "/DeviceMemory/MultiWrite":
		mwReq := rsp.(*sni.MultiWriteMemoryResponse)

		sb := strings.Builder{}
		for i, wReq := range mwReq.GetResponses() {
			sb.WriteString(WriteMemoryResponseString(wReq))
			if i != len(mwReq.GetResponses())-1 {
				sb.WriteRune(',')
			}
		}

		return fmt.Sprintf("uri:\"%s\",responses:[%s]", mwReq.GetUri(), sb.String())
	}

	return fmt.Sprintf("%+v", rsp)
}
