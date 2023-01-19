package main

import (
	"log"
	"os"

	"github.com/veandco/go-sdl2/sdl"
	
	"github.com/yusukemisa/gones/rom"
)

const (
	windowWidth, windowHeight = 256, 240
)

func main() {
	f, err := os.Open("sample1.nes")
	if err != nil {
		log.Fatal(err)
	}

	rom := rom.NewRom(f)
	ppu := NewPPU(rom.CHR)
	cpu := &CPU{
		register: &Register{
			PC: 0x8000,
		},
		memory: make([]byte, 0x10000),
		ppu:    ppu,
	}

	for b := 0; b < len(rom.PRG); b++ {
		cpu.memory[0x8000+b] = rom.PRG[b]
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
