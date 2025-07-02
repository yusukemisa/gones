package bus

import (
	"fmt"

	"github.com/yusukemisa/gones/joypad"
	"github.com/yusukemisa/gones/rom"
)

// PPUInterface defines the interface for PPU operations
type PPUInterface interface {
	ReadRegister(addr uint16) byte
	WriteRegister(addr uint16, data byte)
}

// Bus is a wire between CPU and RAM.
// Three buses are connected from CPU to RAM.
// Physically a wire is essential, but as an emulator program it is not necessary to implement it,
// because it can be used to access memory in the CPU structure.
// But it's useful to keep the code clean.
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
	ppu    PPUInterface

	joyPad1 *joypad.Joypad
}

func NewBus(rom *rom.Rom, ppu PPUInterface) *Bus {
	return &Bus{
		cpuRAM:  make([]byte, 0x0800),
		ppu:     ppu,
		rom:     rom,
		joyPad1: &joypad.Joypad{},
	}
}

func (b *Bus) Read(address uint16) byte {
	switch {
	case address >= 0x0000 && address <= 0x1FFF:
		return b.cpuRAM[address&0x07FF] // Mirror down 0x0800 bytes to 0x0000-0x07FF
	case address >= 0x2000 && address <= 0x3FFF:
		return b.ppu.ReadRegister(0x2000 + (address & 0x7))
	case address >= 0x8000 && address <= 0xFFFF:
		// PRG-ROM mirroring
		prgAddr := address - 0x8000
		if b.rom.PRGSize == 0x4000 && address >= 0xC000 {
			// If PRG-ROM is 16KB, mirror 0x8000-0xBFFF to 0xC000-0xFFFF
			prgAddr = prgAddr & 0x3FFF
		}
		if address == 0xFFFC || address == 0xFFFD {
			fmt.Printf("Reading Reset Vector at %04x: %02x\n", address, b.rom.ReadPRG(prgAddr))
		}
		return b.rom.ReadPRG(prgAddr)
	default:
		return 0
	}
}

func (b *Bus) Write(address uint16, data byte) {
	if 0 <= address && address < 0x2000 {
		mirrorDownAddress := address & 0b0000_0111_1111_1111
		b.cpuRAM[mirrorDownAddress] = data
		return
	}
	if 0x2000 <= address && address < 0x4000 {
		fmt.Printf("Bus Write to PPU Register: addr=%04x, data=%02x\n", address, data)
		switch address {
		case 0x2000:
			b.ppu.WriteRegister(address, data)
		case 0x2001:
			b.ppu.WriteRegister(address, data)
		case 0x2005:
			b.ppu.WriteRegister(address, data)
		case 0x2006:
			b.ppu.WriteRegister(address, data)
		case 0x2007:
			b.ppu.WriteRegister(address, data)
		default:
			mirrorDownAddress := address & 0b0010_0000_0000_0111
			//fmt.Printf("mirrorDownAddress:%#04x,%#04x\n", mirrorDownAddress, address)
			b.Write(mirrorDownAddress, data)
		}
		return
	}
	if address == 0x4016 {
		b.joyPad1.Write(data)
	}
	if 0x8000 <= address && address < 0xFFFF {
		panic(fmt.Sprintf("attempt to write to PRG rom:%#04v", address))
	}
	fmt.Printf("unexpected memory addresses=%#04v, data=%#02x\n", address, data)
}

// SetPPUはPPUを更新する
func (b *Bus) SetPPU(ppu PPUInterface) {
	b.ppu = ppu
}

// UpdatePPUはPPUを更新する
func (b *Bus) UpdatePPU(ppu PPUInterface) {
	b.ppu = ppu
}
