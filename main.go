package main

import (
	"log"
	"os"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/yusukemisa/gones/cpu"
	"github.com/yusukemisa/gones/rom"
)

func main() {
	f, err := os.Open("sample1.nes")
	if err != nil {
		log.Fatal(err)
	}
	run(cpu.NewCPU(rom.NewRom(f), false))
}

func run(cpu *cpu.CPU) {
	for {
		cycle := cpu.Run()
		if screen := cpu.PPU.Run(cycle * 3); screen != nil {
			cpu.PPU.Canvas.Renderer.Present()
			cpu.PPU.Canvas.Renderer.Clear()
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
