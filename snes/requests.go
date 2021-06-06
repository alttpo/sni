package snes

type Response struct {
	IsWrite bool // was the request a read or write?
	Address uint32
	Size    uint8
	Data    []byte      // the data that was read or written
	Extra   interface{} // whatever extra data was passed in as part of the request is handed back
}

type Read struct {
	// E00000-EFFFFF = SRAM
	// F50000-F6FFFF = WRAM
	// F70000-F8FFFF = VRAM
	// F90000-F901FF = CGRAM
	// F90200-F904FF = OAM
	Address    uint32
	Size       uint8
	Extra      interface{} // extra data from the request handed back as part of the response
	Completion func(Response)
}

type Write struct {
	// E00000-EFFFFF = SRAM
	// F50000-F6FFFF = WRAM
	// F70000-F8FFFF = VRAM
	// F90000-F901FF = CGRAM
	// F90200-F904FF = OAM
	Address    uint32
	Size       uint8
	Data       []byte
	Extra      interface{} // extra data from the request handed back as part of the response
	Completion func(Response)
}
