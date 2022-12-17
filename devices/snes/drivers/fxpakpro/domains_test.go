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

var testDomainRefs = [...]sni.MemoryDomainRef_Snes{
	{Snes: sni.MemoryDomainTypeSNES_SNESCartROM},
	{Snes: sni.MemoryDomainTypeSNES_SNESCartRAM},
	{Snes: sni.MemoryDomainTypeSNES_SNESWorkRAM},
	{Snes: sni.MemoryDomainTypeSNES_SNESAPURAM},
	{Snes: sni.MemoryDomainTypeSNES_SNESVideoRAM},
	{Snes: sni.MemoryDomainTypeSNES_SNESCGRAM},
	{Snes: sni.MemoryDomainTypeSNES_SNESObjectAttributeMemory},
	{Snes: sni.MemoryDomainTypeSNES_SNESCoreSpecificMemory},
	{Snes: sni.MemoryDomainTypeSNES_SNESCoreSpecificMemory},
}
var testDomainNames = [...]string{
	"CARTROM",
	"CARTRAM",
	"WRAM",
	"APURAM",
	"VRAM",
	"CGRAM",
	"OAM",
	"FXPAKPRO_SNES",
	"FXPAKPRO_CMD",
}
var testDomainSizes = [...]uint32{
	0xE0_0000,
	0x10_0000,
	0x02_0000,
	0x01_0000,
	0x01_0000,
	0x0200,
	0x0420 - 0x0200,
	0x100_0000,
	0x100_0000,
}
var testDomainWritable = [...]bool{
	true,
	true,
	false,
	false,
	false,
	false,
	false,
	true,
	true,
}

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
			args: args{
				ctx: context.Background(),
				request: &sni.MemoryDomainsRequest{
					Uri: "test",
				},
			},
			wantRsp: &sni.MemoryDomainsResponse{
				Uri: "",
				Domains: []*sni.MemoryDomain{
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[0], Type: &testDomainRefs[0]}, Size: testDomainSizes[0], Readable: true, Writeable: testDomainWritable[0]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[1], Type: &testDomainRefs[1]}, Size: testDomainSizes[1], Readable: true, Writeable: testDomainWritable[1]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[2], Type: &testDomainRefs[2]}, Size: testDomainSizes[2], Readable: true, Writeable: testDomainWritable[2]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[3], Type: &testDomainRefs[3]}, Size: testDomainSizes[3], Readable: true, Writeable: testDomainWritable[3]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[4], Type: &testDomainRefs[4]}, Size: testDomainSizes[4], Readable: true, Writeable: testDomainWritable[4]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[5], Type: &testDomainRefs[5]}, Size: testDomainSizes[5], Readable: true, Writeable: testDomainWritable[5]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[6], Type: &testDomainRefs[6]}, Size: testDomainSizes[6], Readable: true, Writeable: testDomainWritable[6]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[7], Type: &testDomainRefs[7]}, Size: testDomainSizes[7], Readable: true, Writeable: testDomainWritable[7]},
					{Domain: &sni.MemoryDomainRef{Name: &testDomainNames[8], Type: &testDomainRefs[8]}, Size: testDomainSizes[8], Readable: true, Writeable: testDomainWritable[8]},
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
							// WRAM:
							Domain: &sni.MemoryDomainRef{Type: &domainRefs[2]},
							Reads: []*sni.MemoryDomainAddressSize{
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
						// WRAM:
						Domain: &sni.MemoryDomainRef{Type: &domainRefs[2]},
						Reads: []*sni.MemoryDomainAddressData{
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

//func TestDevice_MultiDomainWrite(t *testing.T) {
//	type fields struct {
//		f serial.Port
//	}
//	type args struct {
//		ctx     context.Context
//		request *sni.MultiDomainWriteRequest
//	}
//	tests := []struct {
//		name    string
//		fields  fields
//		args    args
//		want    *sni.MultiDomainWriteResponse
//		wantErr bool
//	}{
//		// TODO: Add test cases.
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Device{
//				f: tt.fields.f,
//			}
//			got, err := d.MultiDomainWrite(tt.args.ctx, tt.args.request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("MultiDomainWrite() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("MultiDomainWrite() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
