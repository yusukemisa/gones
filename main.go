package main

import (
	"flag"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/yusukemisa/gones/bus"
	"github.com/yusukemisa/gones/cpu"
	"github.com/yusukemisa/gones/ppu"
	"github.com/yusukemisa/gones/rom"
	"log"
	"os"
)

func main() {
	// コマンドライン引数の解析
	debug := flag.Bool("debug", false, "Enable debug mode")
	flag.Parse()

	if len(flag.Args()) < 1 {
		log.Fatal("ROM file path is required")
	}

	// ROMファイルを読み込む
	nesFile, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatal("failed to open ROM file:", err)
	}
	defer nesFile.Close()

	// ROMを初期化
	Rom := rom.NewRom(nesFile, *debug)
	Bus := bus.NewBus(Rom, nil)
	CPU := cpu.NewCPU(Bus, *debug)
	PPU := ppu.NewPPU(Rom.CHRROM, *debug, false, CPU)
	Bus.SetPPU(PPU)

	// PPUの初期化待ち
	for i := 0; i < PPU.InitWaitCyclesNTSC; i++ {
		PPU.Run()
	}

	// CPUをリセット
	CPU.Reset()

	// メインループ
	running := true

	// 初期画面をクリア
	PPU.Canvas.Clear()
	PPU.Canvas.Present()

	for running {
		// イベント処理
		for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
			switch event.(type) {
			case *sdl.QuitEvent:
				running = false
			case *sdl.KeyboardEvent:
				if event.(*sdl.KeyboardEvent).Keysym.Scancode == sdl.SCANCODE_ESCAPE {
					running = false
				}
			}
		}

		// CPUを実行
		cycles := CPU.Run()

		// PPUをCPUのサイクル数の3倍実行（NTSCの場合、CPUクロックの3倍）
		for i := 0; i < cycles*3; i++ {
			PPU.Run()
		}
	}

	// 終了処理
	PPU.Canvas.Shutdown()
}
