package snes

// Queue interfaces may also implement this ROMControl interface if they allow for uploading a new ROM and booting a ROM
type ROMControl interface {
	// Uploads the ROM contents to a file called 'name' in a dedicated sni folder
	// Returns the path to pass to BootROM.
	MakeUploadROMCommands(folder string, filename string, rom []byte) (path string, cmds CommandSequence)

	// Boots the given ROM into the system and resets.
	MakeBootROMCommands(path string) CommandSequence
}
