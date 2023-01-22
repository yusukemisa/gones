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
type Bus struct {
	// memory map
	// Address          Size    Usage
	// 0x0000～0x07FF	0x0800	WRAM
	// 0x0800～0x0FFF	-	    WRAMのミラー1
	// 0x1000～0x17FF	-	    WRAMのミラー2
	// 0x1800～0x1FFF	-	    WRAMのミラー3
	// 0x2000～0x2007	0x0008	PPU レジスタ
	// 0x2008～0x3FFF	-	    PPUレジスタのミラー
	// 0x4000～0x401F	0x0020	APU I/O、PAD
	// 0x4020～0x5FFF	0x1FE0	拡張ROM
	// 0x6000～0x7FFF	0x2000	拡張RAM
	// 0x8000～0xBFFF	0x4000	PRG-ROM
	// 0xC000～0xFFFF	0x4000	PRG-ROM
	cpuRAM []byte // 11bit = 2048 = 0x0800
	rom    *rom.Rom
	ppu    *ppu.PPU
}

func NewBus(rom *rom.Rom, ppu *ppu.PPU) *Bus {
	return &Bus{
		cpuRAM: make([]byte, 0x0800),
		ppu:    ppu,
		rom:    rom,
	}
}

func (b *Bus) Read(address uint16) byte {
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
		switch address {
		case 0x2007:
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

func (b *Bus) Write(address uint16, data byte) {
	if 0 <= address && address < 0x2000 {
		mirrorDownAddress := address & 0b0000_0111_1111_1111
		b.cpuRAM[mirrorDownAddress] = data
		return
	}
	if 0x2000 <= address && address < 0x4000 {
		switch address {
		case 0x2000:
			b.ppu.WriteControl(data)
		case 0x2001:
			b.ppu.WriteMask(data)
		case 0x2005:
			b.ppu.WriteScroll(data)
		case 0x2006:
			b.ppu.WriteAddress(data)
		case 0x2007:
			b.ppu.WriteData(data)
		default:
			mirrorDownAddress := address & 0b0010_0000_0000_0111
			//fmt.Printf("mirrorDownAddress:%#04x,%#04x\n", mirrorDownAddress, address)
			b.Write(mirrorDownAddress, data)
		}
		return
	}
	if 0x8000 <= address && address < 0xFFFF {
		panic(fmt.Sprintf("attempt to write to PRG rom:%#04v", address))
	}
	fmt.Printf("unexpected memory addresses=%#04v, data=%#02x\n", address, data)
}
