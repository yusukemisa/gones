package main

import (
	"fmt"
	"log"
)

type CPU struct {
	A byte // アキュムレータ
	X byte // インデックスレジスタ
	Y byte // インデックスレジスタ
	S byte // スタックポインタ

	// プログラムカウンタ
	// CPUはfetch
	// CPUはfetchでPCのアドレスから命令を読む
	PC uint16

	// ステータスレジスタ
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

	memory []byte
}

type instruction struct {
	code        byte
	name        string
	mode        string
	description string
}

// run is main processing in CPU
func (c *CPU) run() {
	for i := 0; i < 20; i++ {
		code := c.fetch()
		inst, ok := opecodes[code]
		if !ok {
			log.Fatalf("opecode not found:%#02x", code)
		}
		c.exec(inst)
	}
}

func (c *CPU) fetch() byte {
	address := c.PC
	c.PC++
	return c.memory[address]
}

func (c *CPU) exec(inst *instruction) {
	fmt.Printf("%#v\n", inst)
	switch inst.name {
	case "SEI":
		// IRQ割り込み禁止
		// bit2を立てる
		c.P = c.P + 4
	case "LDX":
		if inst.mode == "Immediate" {
			c.X = c.fetch()
		}
	case "LDY":
		if inst.mode == "Immediate" {
			c.Y = c.fetch()
		}
	case "LDA":
		switch inst.mode {
		case "Immediate":
			c.A = c.fetch()
		case "AbsoluteX":
			//(IM16+X)番地の値をAにロード
			l, h := uint16(c.fetch()), uint16(c.fetch())
			addr := l | h<<8 + uint16(c.X)
			fmt.Printf("A:%#02x\n", c.A)
			c.A = c.memory[addr]
			fmt.Printf("AbsoluteX:%#04x\n", addr)
			fmt.Printf("A:%#02x\n", c.A)
		}
	case "STA":
		if inst.mode == "Absolute" {
			l, h := uint16(c.fetch()), uint16(c.fetch())
			//fmt.Printf("%#08b,%#08b\n", l, h)
			fmt.Printf("%#04x\n", l|h<<8)
			c.A = c.memory[l|h<<8]
		}
	case "TXS":
		c.S = c.X
	case "INX":
		c.X++
	case "DEY":
		c.Y--
	default:
		fmt.Printf("unknown code:%+v\n", inst)
	}
}
