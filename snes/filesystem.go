package snes

import "sni/protos/sni"

type DirEntry struct {
	Name string
	Type sni.DirEntryType
	Size uint64
}

type DeviceFilesystem interface {
	ReadDirectory(path string) ([]*DirEntry, error)
}
