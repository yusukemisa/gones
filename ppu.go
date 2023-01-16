package main

import (
	"fmt"
	"image/color"
)

type PPURegister struct {
	// Control
	CTRL   byte // 0x2000 割り込み
	MASK   byte // 0x2001 背景イネーブル
	SCROLL byte // 0x2005 背景スクロール

	// Object Attribute Memory - スプライトに対応する空間
	OAMAddress byte // 0x2003
	OAMData    byte // 0x2004

	// PPU memory R/W用
	Address byte // 0x2006
	Data    byte // 0x2007
}

type AddressRegister struct {
	high, low        byte
	lowShouldBeWrite bool
}

// Screen represents a screen to be displayed in a window.
// Window here means the window of this application viewed directly by the user on the monitor.
type Screen struct {
	tile  []Tile
	pixel []byte
}

type Tile struct {
	x, y int
	c    color.RGBA
}

func NewPPU(CHRROM []byte) *PPU {

	// Spriteの初期化
	sprites := make(map[int][]byte)
	var count int
	for i := 0; i < len(CHRROM); i += 0x10 {
		high, low := CHRROM[i:i+8], CHRROM[i+0x08:i+0x10]
		//fmt.Printf("i=%#04x:high=%#v,low=%#v\n", i, high, low)
		sprite := make([]byte, 64)
		for j := byte(0); j < 8; j++ {
			for k := byte(8); 0 < k; k-- {
				var v byte
				if testBit(high[j], k) {
					v = setBit(v, 1)
				}
				if testBit(low[j], k) {
					v = setBit(v, 0)
				}
				//fmt.Printf("%d,%d,%d,%d,%#02x\n", j*8, (j*8)+k-8, j, k, v)
				sprite[(j*8)+8-k] = v
			}
		}

		sprites[count] = sprite
		count++
	}

	canvas := &SDL2Canvas{}
	canvas.Setup("gones", windowWidth, windowHeight)

	//printSprite(sprites[0x48])

	return &PPU{
		address: &AddressRegister{},
		memory:  append(CHRROM, make([]byte, 0x2000)...),
		sprites: sprites,
		canvas:  canvas,
	}
}

func printSprite(sprite []byte) {
	for i, v := range sprite {
		fmt.Printf("%02x ", v)
		if (i+1)%8 == 0 {
			fmt.Println("")
		}
	}
	fmt.Println("----")
	//for i := 0; i < 8; i++ {
	//	for j := 0; j < 8; j++ {
	//		fmt.Printf("%02x ", sprite[j+(8*i)])
	//	}
	//	fmt.Println("")
	//}
}

func (ar *AddressRegister) set(data uint16) {
	ar.high = byte(data >> 8)                // 右シフトして上位8桁をとりだす
	ar.low = byte(data & 0b0000000011111111) // 下位8ビットの&をとる
}

func (ar *AddressRegister) get() uint16 {
	return uint16(ar.high)<<8 | uint16(ar.low)
}

// 8bitごと書き込む
func (ar *AddressRegister) update(data byte) {
	if ar.lowShouldBeWrite {
		ar.low = data
	} else {
		ar.high = data
	}

	// mirror down address above 0x3FFF
	if ar.get() > 0x3FFF {
		ar.set(ar.get() & 0b1111111111111111)
	}
	ar.lowShouldBeWrite = !ar.lowShouldBeWrite
}

func (ar *AddressRegister) increment() {
	ar.set(ar.get() + 1)
}

type PPU struct {
	cycle int
	line  int
	// CPUに読ませるのはこちらの内部バッファ.
	// 直接PPUメモリやROMから読んだ内容にアクセスさせない
	internalDataBuf byte
	address         *AddressRegister

	register *PPURegister
	// memory map
	// Address          Size    Usage
	// 0x0000～0x0FFF	0x1000	パターンテーブル0
	// 0x1000～0x1FFF	0x1000	パターンテーブル1
	// 0x2000～0x23BF	0x03c0	ネームテーブル0
	// 0x23C0～0x23FF	0x0040	属性テーブル0
	// 0x2400～0x27BF	0x03c0	ネームテーブル1
	// 0x27C0～0x27FF	0x0040	属性テーブル1
	// 0x2800～0x2BBF	0x03c0	ネームテーブル2
	// 0x2BC0～0x2BFF	0x0040	属性テーブル2
	// 0x2C00～0x2FBF	0x03c0	ネームテーブル3
	// 0x2FC0～0x2FFF	0x0040	属性テーブル3
	// 0x3000～0x3EFF	-    	0x2000~0x2EFFのミラー
	// 0x3F00～0x3F0F	0x0010	バックグラウンドパレット
	// 0x3F10～0x3F1F	0x0010	スプライトパレット
	// 0x3F20～0x3FFF	0x0040	0x3F00~0x3F1Fのミラー
	memory  []byte
	sprites map[int][]byte
	tiles   []*Tile
	canvas  *SDL2Canvas
}

func (p *PPU) read() byte {
	addr := p.address.get()
	p.address.increment()

	result := p.internalDataBuf
	p.internalDataBuf = p.memory[addr]
	return result
}

func (p *PPU) writeAddress(data byte) {
	p.address.update(data)
}

func (p *PPU) writeData(data byte) {
	addr := p.address.get()
	p.memory[addr] = data
	p.address.increment()
}

func (p *PPU) run(cycle int) *Screen {
	p.cycle += cycle
	if p.cycle >= 341 {
		p.cycle -= 341
		p.line++

		// line=1行目から始まる
		if p.line <= 240 && p.line%8 == 0 {
			p.buildBackGround(p.line)
		}
		if p.line == 262 {
			p.line = 0
			return &Screen{}
		}
	}
	return nil
}

// 1行分だけつくる
func (p *PPU) buildBackGround(line int) {
	// line=120; 0x1C0~0x1E0 14 * 32 = 448(1C0)
	index := (line / 8) - 1
	for i := 0; i < 0x20; i++ {
		p.buildTile(0x20*index + i)
	}
}

// レンダリングに必要な形式(x,y,RGB)にコンバートする
func (p *PPU) buildTile(tileAddress int) {
	// sprite取得
	spriteNum := p.memory[0x2000+tileAddress]
	sprite := p.sprites[int(spriteNum)]

	for y := 0; y < 8; y++ {
		for x := 0; x < 8; x++ {
			colorNum := getSpriteColor(x, y, sprite)
			//colorNum := sprite[x+y*8]
			var c *color.RGBA
			switch colorNum {
			case 0:
				// 黒
				c = &color.RGBA{
					R: 0x50,
					G: 0x50,
					B: 0x50,
				}
			case 1:
				// グレー
				c = &color.RGBA{
					R: 0x80,
					G: 0x80,
					B: 0x80,
				}
			case 2:
				// ライトグレー
				c = &color.RGBA{
					R: 0xC7,
					G: 0xC7,
					B: 0xC7,
				}
			case 3:
				// 白
				c = &color.RGBA{
					R: 0xFF,
					G: 0xFF,
					B: 0xFF,
				}
			}

			p.canvas.SetPixel(x+(tileAddress%0x20)*8, y+(tileAddress/0x20)*8, *c)
		}
	}
}

func getSpriteColor(x int, y int, sprite []byte) byte {
	// (x,y) = (0,0) -> byte[0]
	// (x,y) = (7,0) -> byte[7]
	// (x,y) = (0,1) -> byte[8]
	// (x,y) = (7,1) -> byte[15]
	// (x,y) = (0,2) -> byte[16]
	n := sprite[y*8+x]
	if 0 <= n && n <= 3 {
		return sprite[y*8+x]
	}
	return 0
}
