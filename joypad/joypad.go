package joypad

import (
	"fmt"

	"github.com/veandco/go-sdl2/sdl"

	"github.com/yusukemisa/gones/util"
)

// button input
const (
	RIGHT  = 0b10000000
	LEFT   = 0b01000000
	DOWN   = 0b00100000
	UP     = 0b00010000
	START  = 0b00001000
	SELECT = 0b00000100
	A      = 0b00000010
	B      = 0b00000001
)

var bitKeyMap = map[byte]string{
	0b10000000: "RIGHT",
	0b01000000: "LEFT",
	0b00100000: "DOWN",
	0b00010000: "UP",
	0b00001000: "START",
	0b00000100: "SELECT",
	0b00000010: "A",
	0b00000001: "B",
}

var keyMap = map[sdl.Keycode]byte{
	sdl.K_RIGHT: RIGHT,
	sdl.K_LEFT:  LEFT,
	sdl.K_DOWN:  DOWN,
	sdl.K_UP:    UP,
	sdl.K_z:     START,
	sdl.K_x:     SELECT,
	sdl.K_a:     A,
	sdl.K_s:     B,
}

type Joypad struct {
	strobe       bool // if waiting for button input
	buttonIndex  byte
	buttonStatus byte
}

func (j *Joypad) PollEvent() bool {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch event.(type) {
		case *sdl.QuitEvent:
			return true
		case *sdl.KeyboardEvent:
			e := event.(*sdl.KeyboardEvent)
			if e.Type == sdl.KEYDOWN {
				j.setButtonPressedStatus(keyMap[e.Keysym.Sym], true)
			}
			if e.Type == sdl.KEYUP {
				j.setButtonPressedStatus(keyMap[e.Keysym.Sym], false)
			}
		}
	}
	return false
}

func (j *Joypad) setButtonPressedStatus(key byte, press bool) {
	fmt.Printf("press: %v, %s\n", press, bitKeyMap[key])
	if press {
		j.buttonStatus = util.SetBit(j.buttonStatus, key)
	} else {
		j.buttonStatus = util.ClearBit(j.buttonStatus, key)
	}
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
