package fxpakpro

import (
	"context"
	"reflect"
	"sni/devices/platforms"
	"sni/protos/sni"
	"testing"
)

var testDomains = [...]snesDomain{
	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/bus/system/main",
				Size: 0x1_000000,
			},
			IsExposed:      false,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    true,
		},
		start: 0,
	},

	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/mem/system/WRAM",
				Size: 0x2_0000,
			},
			IsExposed:      true,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    false,
		},
		start: 0,
	},
	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/mem/system/APURAM",
				Size: 0x1_0000,
			},
			IsExposed:      true,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    false,
		},
		start: 0,
	},
	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/mem/system/VRAM",
				Size: 0x1_0000,
			},
			IsExposed:      true,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    false,
		},
		start: 0,
	},
	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/mem/system/CGRAM",
				Size: 0x200,
			},
			IsExposed:      true,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    false,
		},
		start: 0,
	},
	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/mem/system/OAM",
				Size: 0x220,
			},
			IsExposed:      true,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    false,
		},
		start: 0,
	},

	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/mem/cart/ROM",
				Size: 0xE0_0000,
			},
			IsExposed:      true,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    true,
		},
		start: 0,
	},
	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/mem/cart/SRAM",
				Size: 0x10_0000,
			},
			IsExposed:      true,
			IsCoreSpecific: false,
			IsReadable:     true,
			IsWriteable:    true,
		},
		start: 0xE0_0000,
	},

	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/space/fxpakpro/SNES",
				Size: 0x1_000000,
			},
			IsExposed:      true,
			IsCoreSpecific: true,
			IsReadable:     true,
			IsWriteable:    true,
		},
		start: 0,
	},
	{
		Domain: platforms.Domain{
			DomainConf: platforms.DomainConf{
				Name: "snes/space/fxpakpro/CMD",
				Size: 0x1_000000,
			},
			IsExposed:      true,
			IsCoreSpecific: true,
			IsReadable:     true,
			IsWriteable:    true,
		},
		start: 0,
	},
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
			Name:           testDomains[i].Name,
			IsExposed:      testDomains[i].IsExposed,
			IsCoreSpecific: testDomains[i].IsCoreSpecific,
			Size:           testDomains[i].Size,
			IsReadable:     testDomains[i].IsReadable,
			IsWriteable:    testDomains[i].IsWriteable,
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
			name: "WRAM read",
			fields: fields{
				c: &commandsMock{
					vgetMock: func(ctx context.Context, space space, chunks ...vgetChunk) (err error) {
						copy(chunks[0].target, []byte{
							0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
							0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
						})
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
							Name: testDomains[1].Name,
							Reads: []*sni.MemoryDomainOffsetSize{
								{
									Offset: 0,
									Size:   0x10,
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
						Name: testDomains[1].Name,
						Reads: []*sni.MemoryDomainOffsetData{
							{
								Offset: 0,
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
		{
			name: "WRAM read 2 chunks",
			fields: fields{
				c: &commandsMock{
					vgetMock: func(ctx context.Context, space space, chunks ...vgetChunk) (err error) {
						if chunks[0].addr == 0xf5_0010 {
							copy(chunks[0].target, []byte{
								0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
								0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
							})
						}
						if chunks[1].addr == 0xf5_0400 {
							copy(chunks[1].target, []byte{
								0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
								0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
							})
						}
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
							Name: testDomains[1].Name,
							Reads: []*sni.MemoryDomainOffsetSize{
								{
									Offset: 0x10,
									Size:   0x10,
								},
								{
									Offset: 0x400,
									Size:   0x10,
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
						Name: testDomains[1].Name,
						Reads: []*sni.MemoryDomainOffsetData{
							{
								Offset: 0x10,
								Data: []byte{
									0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
									0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
								},
							},
							{
								Offset: 0x400,
								Data: []byte{
									0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
									0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07,
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
//		c *commandsMock
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
//		{
//			name: "WRAM write",
//			fields: fields{
//				c: &commandsMock{
//					// encode test expectations:
//					t: t,
//					next: func(c *commandsMock) {
//						c.vgetMock = nil
//						c.vputMock = nil
//						if c.n == 0 || c.n == 2 {
//							c.vgetMock = func(ctx context.Context, space space, chunks ...vgetChunk) (err error) {
//								if len(chunks) != 1 {
//									c.t.Errorf("expected len(chunks)=1; got %d", len(chunks))
//								}
//								if space != SpaceCMD {
//									c.t.Errorf("expected space=CMD; got %s", space)
//								}
//								if chunks[0].addr != 0x00_2c00 {
//									c.t.Errorf("expected addr=002c00; got %06x", chunks[0].addr)
//								}
//								t.Log(hex.Dump(chunks[0].target))
//								return nil
//							}
//						} else if c.n == 1 {
//							c.vputMock = func(ctx context.Context, space space, chunks ...vputChunk) (err error) {
//								return nil
//							}
//						}
//					},
//					teardown: func(c *commandsMock) {
//						if c.n != 3 {
//							c.t.Errorf("expected 3 calls to commands interface; got %d", c.n)
//						}
//					},
//				},
//			},
//			args: args{
//				ctx: context.Background(),
//				request: &sni.MultiDomainWriteRequest{
//					Uri: "",
//					Requests: []*sni.GroupedDomainWriteRequests{
//						{
//							Domain: &sni.MemoryDomainRef{Name: s("WRAM"), Type: &domainRefs[3]},
//							Writes: []*sni.MemoryDomainAddressData{
//								{
//									Address: 0x10,
//									Data:    []byte{0x11},
//								},
//							},
//						},
//					},
//				},
//			},
//			want: &sni.MultiDomainWriteResponse{
//				Uri: "",
//				Responses: []*sni.GroupedDomainWriteResponses{
//					{
//						Domain: &sni.MemoryDomainRef{
//							Name: s("WRAM"),
//							Type: &domainRefs[3],
//						},
//						Writes: []*sni.MemoryDomainAddressSize{
//							{
//								Address: 0x10,
//								Size:    1,
//							},
//						},
//					},
//				},
//			},
//			wantErr: false,
//		},
//	}
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			d := &Device{
//				c: tt.fields.c,
//			}
//			got, err := d.MultiDomainWrite(tt.args.ctx, tt.args.request)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("MultiDomainWrite() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if tt.fields.c.teardown != nil {
//				tt.fields.c.teardown(tt.fields.c)
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("MultiDomainWrite() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}
