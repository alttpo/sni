package fxpakpro

import (
	"log"
	"sni/snes"
	"strings"
)

func (q *Queue) MakeUploadROMCommands(folder string, filename string, rom []byte) (path string, cmds snes.CommandSequence) {
	// let the folder and filename be joined correctly:
	folder = strings.TrimRight(folder, "/")
	filename = strings.TrimLeft(filename, "/")
	filename = strings.ToLower(filename)
	path = strings.Join([]string{folder, filename}, "/")

	cmds = snes.CommandSequence{
		snes.CommandWithCompletion{Command: newMKDIR(folder)},
		snes.CommandWithCompletion{Command: newPUTFile(path, rom, func(sent, total int) {
			log.Printf("fxpakpro: upload '%s': %#06x of %#06x\n", path, sent, total)
		})},
	}

	return
}

func (q *Queue) MakeBootROMCommands(path string) snes.CommandSequence {
	return snes.CommandSequence{
		snes.CommandWithCompletion{Command: newBOOT(path)},
	}
}
