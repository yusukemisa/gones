package main

import (
	"fmt"
	"log"
)

type CPU struct {
	register *Register
	// memory map
	// Address          Size    Usage
	// 0x0000～0x07FF	0x0800	WRAM
	// 0x0800～0x1FFF	-	    WRAMのミラー
	// 0x2000～0x2007	0x0008	PPU レジスタ
	// 0x2008～0x3FFF	-	    PPUレジスタのミラー
	// 0x4000～0x401F	0x0020	APU I/O、PAD
	// 0x4020～0x5FFF	0x1FE0	拡張ROM
	// 0x6000～0x7FFF	0x2000	拡張RAM
	// 0x8000～0xBFFF	0x4000	PRG-ROM
	// 0xC000～0xFFFF	0x4000	PRG-ROM
	memory []byte

	ppu *PPU

	// test時は任意のメモリマップにしたい
	debug bool
}

type Register struct {
	A byte // アキュムレータ
	X byte // インデックスレジスタ
	Y byte // インデックスレジスタ
	S byte // スタックポインタ

	// プログラムカウンタ
	// CPUはfetch
	// CPUはfetchでPCのアドレスから命令を読む
	PC uint16

	// ステータスレジスタ
	// 条件付きの分岐命令を実行するために演算結果を保持する
	//bit	名称	詳細	            内容
	//bit7	N	ネガティブ	    演算結果のbit7が1の時にセット
	//bit6	V	オーバーフロー	P演算結果がオーバーフローを起こした時にセット
	//bit5	R	予約済み	        常にセットされている
	//bit4	B	ブレークモード	BRK発生時にセット、IRQ発生時にクリア
	//bit3	D	デシマルモード	0:デフォルト、1:BCDモード (未実装)
	//bit2	I	IRQ禁止	        0:IRQ許可、1:IRQ禁止
	//bit1	Z	ゼロ	            演算結果が0の時にセット
	//bit0	C	キャリー	        キャリー発生時にセット
	P byte
}

type instruction struct {
	code        byte
	name        string
	mode        string
	description string
}

// run is main processing in CPU
func (c *CPU) run() {
	for i := 0; i < 200; i++ {
		code := c.fetch()
		fmt.Printf("i=%d:code:%#02x\n", i, code)
		inst, ok := opecodes[code]
		if !ok {
			log.Fatalf("opecode not found:%#02x", code)
		}
		c.exec(inst)
	}
}

func (c *CPU) fetch() byte {
	address := c.register.PC
	c.register.PC++
	return c.read(address)
}

func (c *CPU) exec(inst *instruction) {
	fmt.Printf("%#v, \n", inst)
	switch inst.name {
	case "JMP":
		l, h := uint16(c.fetch()), uint16(c.fetch())
		c.register.PC = l | h<<8
	case "SEI":
		// IRQ割り込み禁止
		// bit2を立てる
		c.register.P = c.register.P + 4
	case "LDX":
		if inst.mode == "Immediate" {
			c.register.X = c.fetch()
			c.updateStatusRegister(c.register.X)
		}
	case "LDY":
		if inst.mode == "Immediate" {
			c.register.Y = c.fetch()
			c.updateStatusRegister(c.register.Y)
		}
	case "LDA":
		switch inst.mode {
		case "Immediate":
			c.register.A = c.fetch()
		case "AbsoluteX":
			//(IM16+X)番地の値をAにロード
			l, h := uint16(c.fetch()), uint16(c.fetch())
			addr := l | h<<8 + uint16(c.register.X)
			c.register.A = c.read(addr)
		}
		c.updateStatusRegister(c.register.A)
	case "STA":
		if inst.mode == "Absolute" {
			l, h := uint16(c.fetch()), uint16(c.fetch())
			c.register.A = c.read(l | h<<8)
		}
	case "TXS":
		c.register.S = c.register.X
	case "INX":
		c.register.X++
		c.updateStatusRegister(c.register.X)
	case "INC":
		if inst.mode == "ZeroPageX" {
			//fmt.Printf("PC=%#04x,X=%#04x\n", c.PC, uint16(c.X))
			addr := c.register.PC + uint16(c.register.X)
			//fmt.Printf("addr=%#04x\n", addr)
			c.write(addr, c.read(addr)+1)
			c.updateStatusRegister(c.read(addr))
		}
	case "DEY":
		c.register.Y--
		c.updateStatusRegister(c.register.Y)
	case "BNE":
		if inst.mode == "Relative" {
			// 分岐するしないに関係なくPCが2byte回る必要ある
			relAddr := int8(c.fetch())
			if !testBit(c.register.P, 1) {
				// uint8で取得した値を-128~127の範囲にキャストしてアドレスを計算
				// 0xFFの場合アドレスを-1することになる
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	default:
		fmt.Printf("unknown code:%#v\n", inst)
	}
	fmt.Printf("A:%#02x,X:%#02x,Y:%#02x,PC:%#04x\n", c.register.A, c.register.X, c.register.Y, c.register.PC)
}

// TODO:Bus導入
func (c *CPU) write(address uint16, data byte) {
	switch address {
	case 0x2006:
		c.ppu.write(data)
	}
	c.memory[address] = data
}

func (c *CPU) read(address uint16) byte {
	if c.debug {
		return c.memory[address]
	}
	if address < 0x2000 {
		// 0x0000～0x07FF	0x0800	WRAM
		// 0x0800～0x1FFF	-	    WRAMのミラー
		return 0
	} else if address < 0x4000 {
		// 0x2000～0x2007	0x0008	PPU レジスタ
		// 0x2008～0x3FFF	-	    PPUレジスタのミラー
		registerNumber := (address - 0x2000) % 8
		switch registerNumber {
		case 7:
			return c.ppu.read()
		}
		return 0
	}
	if address >= 0x8000 {
		return c.memory[address]
	}
	return 0
}

// updateStatusRegister updates status register.
// bit	名称	詳細	            内容
// bit7	N	ネガティブ	    演算結果のbit7が1の時にセット
// bit1	Z	ゼロ	            演算結果が0の時にセット
func (c *CPU) updateStatusRegister(result byte) {
	// bit1	Z
	if result == 0 {
		c.register.P = setBit(c.register.P, 1)
	} else {
		c.register.P = clearBit(c.register.P, 1)
	}

	// Bit7 N
	// Aの最上部bitの値とのORをとる
	if testBit(result, 7) {
		c.register.P = setBit(c.register.P, 7)
	} else {
		c.register.P = clearBit(c.register.P, 7)
	}
	fmt.Printf("result=%#02x,Z=%v,N=%v\n", result, testBit(c.register.P, 1), testBit(c.register.P, 7))
}

func testBit(x, n byte) bool {
	return x&(1<<n) != 0
}

func setBit(x, n byte) byte {
	return x | (1 << n)
}

func clearBit(x, n byte) byte {
	return x &^ (1 << n)
}
