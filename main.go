package main

import (
	"github.com/veandco/go-sdl2/sdl"
	"io"
	"log"
	"os"
)

const (
	windowWidth, windowHeight = 256, 240
)

func main() {
	f, err := os.Open("sample1.nes")
	if err != nil {
		log.Fatal(err)
	}

	hr := io.NewSectionReader(f, 0, 0x10)
	buf := make([]byte, 0x10) // 16ByteのiNESヘッダ
	_, err = hr.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	// 00000000  4e 45 53 1a 02 01 01 00  00 00 00 00 00 00 00 00  |NES.............|
	// 0-3: Constant $4E $45 $53 $1A ("NES" followed by MS-DOS end-of-file)
	//   4: Size of PRG ROM in 16 KB units
	//   5: Size of CHR ROM in 8 KB units (Value 0 means the board uses CHR RAM)
	sizeOfPRG, sizeOfCHR := int(buf[4]), int(buf[5])
	pr := io.NewSectionReader(f, 0x10, int64(sizeOfPRG*0x4000))
	cr := io.NewSectionReader(f, int64(0x10+sizeOfPRG*0x4000), int64(sizeOfCHR*0x2000))
	PRGROM, CHRROM := make([]byte, sizeOfPRG*0x4000), make([]byte, sizeOfCHR*0x2000)

	_, err = pr.Read(PRGROM)
	if err != nil {
		log.Fatal("failed to read PRGROM:", err)
	}

	_, err = cr.Read(CHRROM)
	if err != nil {
		log.Fatal(err)
	}

	// 起動時/リセット時に0xFFFC/0xFFFDから開始アドレスをリードしてプログラムカウンタPCにセットしてやる必要があります。
	//fmt.Printf("0xFFFC: %#02x\n", cpu.memory[0xFFFC])
	//fmt.Printf("0xFFFD: %#02x\n", cpu.memory[0xFFFD])

	ppu := NewPPU(CHRROM)

	cpu := &CPU{
		register: &Register{
			PC: 0x8000,
		},
		memory: make([]byte, 0x10000),
		ppu:    ppu,
	}

	for b := 0; b < len(PRGROM); b++ {
		cpu.memory[0x8000+b] = PRGROM[b]
	}

	run(cpu)
}

func run(cpu *CPU) {

	for {
		cycle := cpu.run()
		if screen := cpu.ppu.run(cycle * 3); screen != nil {
			cpu.ppu.canvas.renderer.Present()
			cpu.ppu.canvas.renderer.Clear()
			sdl.Delay(100)
		}
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				println("Quit")
				return
			}
		}
	}
}
