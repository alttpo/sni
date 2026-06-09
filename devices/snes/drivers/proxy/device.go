package proxy

import (
	"bytes"
	"context"
	"io"
	"sni/devices"
	"sni/protos/sni"
	"sync"

	"google.golang.org/grpc"
)

// Device proxies all device operations to a remote SNI gRPC server.
type Device struct {
	lock sync.Mutex

	conn      *grpc.ClientConn
	remoteURI string

	devices    sni.DevicesClient
	memory     sni.DeviceMemoryClient
	control    sni.DeviceControlClient
	filesystem sni.DeviceFilesystemClient
	info       sni.DeviceInfoClient
	nwa        sni.DeviceNWAClient

	closed bool
}

func (d *Device) IsClosed() bool {
	d.lock.Lock()
	defer d.lock.Unlock()
	return d.closed
}

func (d *Device) Close() error {
	d.lock.Lock()
	defer d.lock.Unlock()

	d.closed = true
	if d.conn != nil {
		return d.conn.Close()
	}
	return nil
}

// DeviceMemory implementation

func (d *Device) RequiresMemoryMappingForAddressSpace(_ context.Context, addressSpace sni.AddressSpace) (bool, error) {
	if addressSpace == sni.AddressSpace_FxPakPro {
		return false, nil
	}
	if addressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	return true, nil
}

func (d *Device) RequiresMemoryMappingForAddress(_ context.Context, address devices.AddressTuple) (bool, error) {
	if address.AddressSpace == sni.AddressSpace_FxPakPro {
		return false, nil
	}
	if address.AddressSpace == sni.AddressSpace_Raw {
		return false, nil
	}
	return true, nil
}

func (d *Device) MultiReadMemory(ctx context.Context, reads ...devices.MemoryReadRequest) ([]devices.MemoryReadResponse, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return nil, devices.DeviceFatal("proxy: device is closed", nil)
	}

	requests := make([]*sni.ReadMemoryRequest, 0, len(reads))
	for _, read := range reads {
		requests = append(requests, &sni.ReadMemoryRequest{
			RequestAddress:       read.RequestAddress.Address,
			RequestAddressSpace:  read.RequestAddress.AddressSpace,
			RequestMemoryMapping: read.RequestAddress.MemoryMapping,
			Size:                 uint32(read.Size),
		})
	}

	resp, err := d.memory.MultiRead(ctx, &sni.MultiReadMemoryRequest{
		Uri:      d.remoteURI,
		Requests: requests,
	})
	if err != nil {
		return nil, devices.DeviceFatal("proxy: MultiRead failed", err)
	}

	responses := make([]devices.MemoryReadResponse, 0, len(resp.GetResponses()))
	for _, r := range resp.GetResponses() {
		responses = append(responses, devices.MemoryReadResponse{
			RequestAddress: devices.AddressTuple{
				Address:       r.GetRequestAddress(),
				AddressSpace:  r.GetRequestAddressSpace(),
				MemoryMapping: r.GetRequestMemoryMapping(),
			},
			DeviceAddress: devices.AddressTuple{
				Address:       r.GetDeviceAddress(),
				AddressSpace:  r.GetDeviceAddressSpace(),
				MemoryMapping: r.GetRequestMemoryMapping(),
			},
			Data: r.GetData(),
		})
	}

	return responses, nil
}

func (d *Device) MultiWriteMemory(ctx context.Context, writes ...devices.MemoryWriteRequest) ([]devices.MemoryWriteResponse, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return nil, devices.DeviceFatal("proxy: device is closed", nil)
	}

	requests := make([]*sni.WriteMemoryRequest, 0, len(writes))
	for _, write := range writes {
		requests = append(requests, &sni.WriteMemoryRequest{
			RequestAddress:       write.RequestAddress.Address,
			RequestAddressSpace:  write.RequestAddress.AddressSpace,
			RequestMemoryMapping: write.RequestAddress.MemoryMapping,
			Data:                 write.Data,
		})
	}

	resp, err := d.memory.MultiWrite(ctx, &sni.MultiWriteMemoryRequest{
		Uri:      d.remoteURI,
		Requests: requests,
	})
	if err != nil {
		return nil, devices.DeviceFatal("proxy: MultiWrite failed", err)
	}

	responses := make([]devices.MemoryWriteResponse, 0, len(resp.GetResponses()))
	for _, r := range resp.GetResponses() {
		responses = append(responses, devices.MemoryWriteResponse{
			RequestAddress: devices.AddressTuple{
				Address:       r.GetRequestAddress(),
				AddressSpace:  r.GetRequestAddressSpace(),
				MemoryMapping: r.GetRequestMemoryMapping(),
			},
			DeviceAddress: devices.AddressTuple{
				Address:       r.GetDeviceAddress(),
				AddressSpace:  r.GetDeviceAddressSpace(),
				MemoryMapping: r.GetRequestMemoryMapping(),
			},
			Size: int(r.GetSize()),
		})
	}

	return responses, nil
}

// DeviceControl implementation

func (d *Device) ResetSystem(ctx context.Context) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return devices.DeviceFatal("proxy: device is closed", nil)
	}

	_, err := d.control.ResetSystem(ctx, &sni.ResetSystemRequest{
		Uri: d.remoteURI,
	})
	if err != nil {
		return devices.DeviceFatal("proxy: ResetSystem failed", err)
	}
	return nil
}

func (d *Device) ResetToMenu(ctx context.Context) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return devices.DeviceFatal("proxy: device is closed", nil)
	}

	_, err := d.control.ResetToMenu(ctx, &sni.ResetToMenuRequest{
		Uri: d.remoteURI,
	})
	if err != nil {
		return devices.DeviceFatal("proxy: ResetToMenu failed", err)
	}
	return nil
}

func (d *Device) PauseUnpause(ctx context.Context, pausedState bool) (bool, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return false, devices.DeviceFatal("proxy: device is closed", nil)
	}

	resp, err := d.control.PauseUnpauseEmulation(ctx, &sni.PauseEmulationRequest{
		Uri:    d.remoteURI,
		Paused: pausedState,
	})
	if err != nil {
		return false, devices.DeviceFatal("proxy: PauseUnpause failed", err)
	}
	return resp.GetPaused(), nil
}

func (d *Device) PauseToggle(ctx context.Context) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return devices.DeviceFatal("proxy: device is closed", nil)
	}

	_, err := d.control.PauseToggleEmulation(ctx, &sni.PauseToggleEmulationRequest{
		Uri: d.remoteURI,
	})
	if err != nil {
		return devices.DeviceFatal("proxy: PauseToggle failed", err)
	}
	return nil
}

// DeviceFilesystem implementation

func (d *Device) ReadDirectory(ctx context.Context, path string) ([]devices.DirEntry, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return nil, devices.DeviceFatal("proxy: device is closed", nil)
	}

	resp, err := d.filesystem.ReadDirectory(ctx, &sni.ReadDirectoryRequest{
		Uri:  d.remoteURI,
		Path: path,
	})
	if err != nil {
		return nil, devices.DeviceFatal("proxy: ReadDirectory failed", err)
	}

	entries := make([]devices.DirEntry, 0, len(resp.GetEntries()))
	for _, entry := range resp.GetEntries() {
		entries = append(entries, devices.DirEntry{
			Name: entry.GetName(),
			Type: entry.GetType(),
		})
	}
	return entries, nil
}

func (d *Device) MakeDirectory(ctx context.Context, path string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return devices.DeviceFatal("proxy: device is closed", nil)
	}

	_, err := d.filesystem.MakeDirectory(ctx, &sni.MakeDirectoryRequest{
		Uri:  d.remoteURI,
		Path: path,
	})
	if err != nil {
		return devices.DeviceFatal("proxy: MakeDirectory failed", err)
	}
	return nil
}

func (d *Device) RemoveFile(ctx context.Context, path string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return devices.DeviceFatal("proxy: device is closed", nil)
	}

	_, err := d.filesystem.RemoveFile(ctx, &sni.RemoveFileRequest{
		Uri:  d.remoteURI,
		Path: path,
	})
	if err != nil {
		return devices.DeviceFatal("proxy: RemoveFile failed", err)
	}
	return nil
}

func (d *Device) RenameFile(ctx context.Context, path, newFilename string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return devices.DeviceFatal("proxy: device is closed", nil)
	}

	_, err := d.filesystem.RenameFile(ctx, &sni.RenameFileRequest{
		Uri:      d.remoteURI,
		Path:     path,
		NewFilename: newFilename,
	})
	if err != nil {
		return devices.DeviceFatal("proxy: RenameFile failed", err)
	}
	return nil
}

func (d *Device) PutFile(ctx context.Context, path string, size uint32, r io.Reader, progress devices.ProgressReportFunc) (uint32, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return 0, devices.DeviceFatal("proxy: device is closed", nil)
	}

	data, err := io.ReadAll(r)
	if err != nil {
		return 0, devices.DeviceNonFatal("proxy: failed to read file data", err)
	}

	resp, err := d.filesystem.PutFile(ctx, &sni.PutFileRequest{
		Uri:  d.remoteURI,
		Path: path,
		Data: data,
	})
	if err != nil {
		return 0, devices.DeviceFatal("proxy: PutFile failed", err)
	}

	return resp.GetSize(), nil
}

func (d *Device) GetFile(ctx context.Context, path string, w io.Writer, sizeReceived devices.SizeReceivedFunc, progress devices.ProgressReportFunc) (uint32, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return 0, devices.DeviceFatal("proxy: device is closed", nil)
	}

	resp, err := d.filesystem.GetFile(ctx, &sni.GetFileRequest{
		Uri:  d.remoteURI,
		Path: path,
	})
	if err != nil {
		return 0, devices.DeviceFatal("proxy: GetFile failed", err)
	}

	data := resp.GetData()
	size := uint32(len(data))

	if sizeReceived != nil {
		sizeReceived(size)
	}

	n, err := io.Copy(w, bytes.NewReader(data))
	if err != nil {
		return uint32(n), devices.DeviceNonFatal("proxy: failed to write file data", err)
	}

	return size, nil
}

func (d *Device) BootFile(ctx context.Context, path string) error {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return devices.DeviceFatal("proxy: device is closed", nil)
	}

	_, err := d.filesystem.BootFile(ctx, &sni.BootFileRequest{
		Uri:  d.remoteURI,
		Path: path,
	})
	if err != nil {
		return devices.DeviceFatal("proxy: BootFile failed", err)
	}
	return nil
}

// DeviceInfo implementation

func (d *Device) FetchFields(ctx context.Context, fields ...sni.Field) ([]string, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return nil, devices.DeviceFatal("proxy: device is closed", nil)
	}

	resp, err := d.info.FetchFields(ctx, &sni.FieldsRequest{
		Uri:    d.remoteURI,
		Fields: fields,
	})
	if err != nil {
		return nil, devices.DeviceFatal("proxy: FetchFields failed", err)
	}

	return resp.GetValues(), nil
}

// DeviceNWA implementation

func (d *Device) NWACommand(ctx context.Context, cmd string, args string, binaryArg []byte) ([]map[string]string, []byte, error) {
	d.lock.Lock()
	defer d.lock.Unlock()

	if d.closed {
		return nil, nil, devices.DeviceFatal("proxy: device is closed", nil)
	}

	resp, err := d.nwa.NWACommand(ctx, &sni.NWACommandRequest{
		Uri:       d.remoteURI,
		Command:   cmd,
		Args:      args,
		BinaryArg: binaryArg,
	})
	if err != nil {
		return nil, nil, devices.DeviceFatal("proxy: NWACommand failed", err)
	}

	// convert proto reply to map slice
	asciiReply := make([]map[string]string, 0, len(resp.GetAsciiReply()))
	for _, item := range resp.GetAsciiReply() {
		asciiReply = append(asciiReply, item.GetItem())
	}

	return asciiReply, resp.GetBinaryReplay(), nil
}
