package bus

import (
	"fmt"

	"github.com/yusukemisa/gones/ppu"
	"github.com/yusukemisa/gones/rom"
)

// Bus is a wire between CPU and RAM.
// Three buses are connected from CPU to RAM.
// Physically a wire is essential, but as an emulator program it is not necessary to implement it,
// because it can be used to access memory in the CPU structure.
// But it’s useful to keep the code clean.
type bus struct {
	cpuRAM []byte // 11bit = 2048 = 0x0800
	rom    *rom.Rom
	ppu    *ppu.PPU
	debug  bool
}

func NewBus(debug bool) *bus {
	return &bus{
		cpuRAM: make([]byte, 0x0800),
		debug:  debug,
	}
}

func (b *bus) Read(address uint16) byte {
	if b.debug {
		return b.cpuRAM[address]
	}
	// 0x0000～0x07FF	0x0800	WRAM
	// 0x0800～0x0FFF	-	    WRAMのミラー1
	// 0x1000～0x17FF	-	    WRAMのミラー2
	// 0x1800～0x1FFF	-	    WRAMのミラー3
	if 0 <= address && address < 0x2000 {
		mirrorDownAddress := address & 0b0000_0111_1111_1111
		return b.cpuRAM[mirrorDownAddress]
	}
	// 0x2000～0x2007	0x0008	PPU レジスタ
	// 0x2008～0x3FFF	-	    PPUレジスタのミラー
	if 0x2000 <= address && address < 0x4000 {
		//let _mirror_down_addr = addr & 0b00100000_00000111;
		registerNumber := address & 0b0000_0000_0000_0111
		switch registerNumber {
		case 7:
			return b.ppu.Read()
		}
		return 0
	}
	// 0x8000～0xBFFF	0x4000	PRG-ROM
	// 0xC000～0xFFFF	0x4000	PRG-ROM
	if 0x8000 <= address && address < 0xFFFF {
		address -= 0x8000
		mirrorDownAddress := address & 0b0011_1111_1111_1111
		return b.rom.ReadPRG(mirrorDownAddress)
	}
	return 0
}

func (b *bus) write(address uint16, data byte) {
	if 0 <= address && address < 0x2000 {
		mirrorDownAddress := address & 0b0000011111111111
		b.cpuRAM[mirrorDownAddress] = data
		return
	}
	if 0x2000 <= address && address < 0x4000 {
		mirrorDownAddress := address & 0b0000000000000111
		b.cpuRAM[mirrorDownAddress] = data
		return
	}
	fmt.Printf("unexpected memory addresses=%#04v, data=%#02x\n", address, data)
}
