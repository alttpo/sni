package fxpakpro

import (
	"context"
	"go.bug.st/serial"
	"io"
	"reflect"
	"sni/protos/sni"
	"testing"
	"time"
)

type RWFunc = func(p []byte) (int, error)

type testablePort struct {
	DidWrite RWFunc
	NextRead RWFunc
}

func (d *testablePort) Read(p []byte) (n int, err error) {
	f := d.NextRead
	if f == nil {
		return 0, io.EOF
	}
	return f(p)
}

func (d *testablePort) Write(p []byte) (n int, err error) {
	f := d.DidWrite
	if f == nil {
		return 0, io.EOF
	}
	return f(p)
}

func (d *testablePort) GetModemStatusBits() (*serial.ModemStatusBits, error) {
	return &serial.ModemStatusBits{}, nil
}

func (d *testablePort) SetMode(mode *serial.Mode) error      { return nil }
func (d *testablePort) ResetInputBuffer() error              { return nil }
func (d *testablePort) ResetOutputBuffer() error             { return nil }
func (d *testablePort) SetDTR(dtr bool) error                { return nil }
func (d *testablePort) SetRTS(rts bool) error                { return nil }
func (d *testablePort) SetReadTimeout(t time.Duration) error { return nil }
func (d *testablePort) Close() error                         { return nil }

func TestDevice_MemoryDomains(t *testing.T) {
	type fields struct {
		f serial.Port
	}
	type args struct {
		ctx     context.Context
		request *sni.MemoryDomainsRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRsp *sni.MemoryDomainsResponse
		wantErr bool
	}{
		{
			name: "list",
			fields: fields{
				f: nil,
			},
			args: args{},
			wantRsp: &sni.MemoryDomainsResponse{
				Uri: "",
				Domains: []*sni.MemoryDomain{
					{DomainName: "CARTROM", Size: 0xE0_0000},
					{DomainName: "SRAM", Size: 0x10_0000},
					{DomainName: "WRAM", Size: 0x02_0000},
					{DomainName: "VRAM", Size: 0x01_0000},
					{DomainName: "APURAM", Size: 0x01_0000},
					{DomainName: "CGRAM", Size: 0x0200},
					{DomainName: "OAM", Size: 0x0420 - 0x0200},
					{DomainName: "MISC", Size: 0x0500 - 0x0420},
					{DomainName: "PPUREG", Size: 0x0700 - 0x0500},
					{DomainName: "CPUREG", Size: 0x0200},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Device{
				f: tt.fields.f,
			}
			gotRsp, err := d.MemoryDomains(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemoryDomains() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got, want := gotRsp.Domains, tt.wantRsp.Domains; !reflect.DeepEqual(got, want) {
				t.Errorf("MemoryDomains() gotRsp = %+v, want %+v", got, want)
			}
		})
	}
}

func TestDevice_MultiDomainRead(t *testing.T) {
	type fields struct {
		f serial.Port
	}
	type args struct {
		ctx     context.Context
		request *sni.MultiDomainReadRequest
	}

	tests := []struct {
		name    string
		fields  fields
		args    args
		wantRsp *sni.MultiDomainReadResponse
		wantErr bool
	}{
		{
			name: "",
			fields: fields{
				&testablePort{
					DidWrite: nil,
					NextRead: nil,
				},
			},
			args: args{
				ctx: context.Background(),
				request: &sni.MultiDomainReadRequest{
					Uri: "",
					Requests: []*sni.GroupedDomainReadRequests{
						{
							DomainName: "WRAM",
							Reads: []*sni.GroupedDomainReadRequests_AddressSize{
								{
									Address: 0,
									Size:    0x10,
								},
							},
						},
					},
				},
			},
			wantRsp: &sni.MultiDomainReadResponse{
				Uri: "",
				Responses: []*sni.GroupedDomainReadResponses{
					{
						DomainName: "WRAM",
						Reads: []*sni.GroupedDomainReadResponses_AddressData{
							{
								Address: 0,
								Data: []byte{
									0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
									0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
								},
							},
						},
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Device{
				f: tt.fields.f,
			}
			gotRsp, err := d.MultiDomainRead(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultiDomainRead() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotRsp, tt.wantRsp) {
				t.Errorf("MultiDomainRead() gotRsp = %v, want %v", gotRsp, tt.wantRsp)
			}
		})
	}
}

func TestDevice_MultiDomainWrite(t *testing.T) {
	type fields struct {
		f serial.Port
	}
	type args struct {
		ctx     context.Context
		request *sni.MultiDomainWriteRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *sni.MultiDomainWriteResponse
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Device{
				f: tt.fields.f,
			}
			got, err := d.MultiDomainWrite(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("MultiDomainWrite() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("MultiDomainWrite() got = %v, want %v", got, tt.want)
			}
		})
	}
}
