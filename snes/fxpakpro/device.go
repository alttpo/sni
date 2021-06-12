package fxpakpro

// TODO: algorithm to break up reads into VGET 255-byte chunks:
//addr := request.GetAddress()
//size := int32(request.GetSize())
//reads := make([]snes.Read, 0, 8)
//data := make([]byte, 0, size)
//for size > 0 {
//chunkSize := int32(255)
//if size < chunkSize {
//chunkSize = size
//}
//
//reads = append(reads, snes.Read{
//Address: addr,
//Size:    uint8(chunkSize),
//Extra:   nil,
//Completion: func(response snes.Response) {
//data = append(data, response.Data...)
//},
//})
//
//size -= 255
//addr += 255
//}
