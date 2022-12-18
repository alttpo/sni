package fxpakpro

import (
	"context"
	"reflect"
	"sni/protos/sni"
	"testing"
)

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
		commands commands
	}
	type args struct {
		ctx     context.Context
		request *sni.MemoryDomainsRequest
	}

	createDomainFromDesc := func(i int) *sni.MemoryDomain {
		return &sni.MemoryDomain{
			Core: driverName,
			Domain: &sni.MemoryDomainRef{
				Name: domainDescs[i].domainRef.Name,
				Type: domainDescs[i].domainRef.Type,
			},
			Notes:     domainDescs[i].notes,
			Size:      domainDescs[i].size,
			Readable:  true,
			Writeable: domainDescs[i].writeable,
		}
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
				commands: nil,
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
					createDomainFromDesc(0),
					createDomainFromDesc(1),
					createDomainFromDesc(2),
					createDomainFromDesc(3),
					createDomainFromDesc(4),
					createDomainFromDesc(5),
					createDomainFromDesc(6),
					createDomainFromDesc(7),
					createDomainFromDesc(8),
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := &Device{
				c: tt.fields.commands,
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
		c commands
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
				c: &commandsMock{
					vgetMock: func(ctx context.Context, space space, chunks ...vgetChunk) (err error) {
						return nil
					},
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
				c: tt.fields.c,
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
