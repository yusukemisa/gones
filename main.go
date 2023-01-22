package main

import (
	"log"
	"os"
	"time"

	"github.com/yusukemisa/gones/bus"
	"github.com/yusukemisa/gones/cpu"
	"github.com/yusukemisa/gones/joypad"
	"github.com/yusukemisa/gones/ppu"
	"github.com/yusukemisa/gones/rom"
)

func main() {
	f, err := os.Open("sample1.nes")
	if err != nil {
		log.Fatal(err)
	}

	rom := rom.NewRom(f)
	ppu := ppu.NewPPU(rom.CHR, false)
	cpu := cpu.NewCPU(bus.NewBus(rom, ppu))

	run(cpu, ppu, &joypad.Joypad{})
}

func run(cpu *cpu.CPU, ppu *ppu.PPU, joyPad *joypad.Joypad) {
	for {
		cycle := cpu.Run()
		if screen := ppu.Run(cycle * 3); screen != nil {
			ppu.Canvas.Renderer.Present()
			ppu.Canvas.Renderer.Clear()
			time.Sleep(100 * time.Microsecond)
		}
		if quit := joyPad.PollEvent(); quit {
			println("Quit")
			return
		}
	}
}
