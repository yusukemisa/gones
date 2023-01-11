package main

import "math/rand"

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
	pixels []uint32
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
	memory []byte
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
	pixels := make([]uint32, 800*600*4)
	for i := range pixels {
		pixels[i] = 0x00777777 + uint32(rand.Intn(0x00AAAAAA))
	}

	return &Screen{pixels}
}
