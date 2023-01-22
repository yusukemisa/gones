package joypad

import "github.com/veandco/go-sdl2/sdl"

type Joypad struct {
	strobe       bool // if waiting for button input
	buttonIndex  byte
	buttonStatus byte
}

func PollEvent() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return true
		case *sdl.KeyboardEvent:
			e := event.(*sdl.KeyboardEvent)
			if e.Type == sdl.KEYDOWN {
				if e.Keysym.Sym == sdl.K_DOWN {
					println("put DOWN")
				}
				if e.Keysym.Sym == sdl.K_LEFT {
					println("put LEFT")
				}
				if e.Keysym.Sym == sdl.K_RIGHT {
					println("put RIGHT")
				}
				if e.Keysym.Sym == sdl.K_UP {
					println("put UP")
				}
				if e.Keysym.Sym == sdl.K_a {
					println("put a")
				}
				if e.Keysym.Sym == sdl.K_s {
					println("put s")
				}
				if e.Keysym.Sym == sdl.K_x {
					println("put s")
				}
				if e.Keysym.Sym == sdl.K_z {
					println("put z")
				}
			}
		}
	}
	return false
}

func (j *Joypad) Write(data byte) {
	j.strobe = data == 0x01
	if j.strobe {
		j.buttonIndex = 0
	}
}

func (j *Joypad) Read() byte {
	// index > 7 の場合の挙動の根拠は？
	if j.buttonIndex > 7 {
		return 1
	}

	res := (j.buttonStatus & (0b00000001 << j.buttonIndex)) >> j.buttonIndex // 0/1に変換
	if !j.strobe && j.buttonIndex <= 7 {
		j.buttonIndex++
	}
	return res
}
