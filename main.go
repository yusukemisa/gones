package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

func main() {
	f, err := os.Open("sample1.nes")
	if err != nil {
		log.Fatal(err)
	}

	r := bufio.NewReader(f)
	buf := make([]byte, 0x10) // 16ByteのiNESヘッダ
	_, err = r.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	// 00000000  4e 45 53 1a 02 01 01 00  00 00 00 00 00 00 00 00  |NES.............|
	// 0-3: Constant $4E $45 $53 $1A ("NES" followed by MS-DOS end-of-file)
	//   4: Size of PRG ROM in 16 KB units
	//   5: Size of CHR ROM in 8 KB units (Value 0 means the board uses CHR RAM)
	sizeOfPRG, sizeOfCHR := int(buf[4]), int(buf[5])
	PRGROM, CHRROM := make([]byte, sizeOfPRG*0x4000), make([]byte, sizeOfCHR*0x2000)

	_, err = r.Read(PRGROM)
	if err != nil {
		log.Fatal(err)
	}
	_, err = r.Read(CHRROM)
	if err != nil {
		log.Fatal(err)
	}

	cpu := &CPU{
		PC:     0x8000,
		memory: nil,
	}
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
	cpu.memory = append(make([]byte, 0x8000), PRGROM...)

	//fmt.Printf("%#x\n", len(cpu.memory))

	// 起動時/リセット時に0xFFFC/0xFFFDから開始アドレスをリードしてプログラムカウンタPCにセットしてやる必要があります。
	//fmt.Printf("0xFFFC: %#02x\n", cpu.memory[0xFFFC])
	//fmt.Printf("0xFFFD: %#02x\n", cpu.memory[0xFFFD])

	cpu.run()
}

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
	name        string
	mode        string
	description string
}

// run is main processing in CPU
func (c *CPU) run() {
	for i := 0; i < 10; i++ {
		code := c.fetch()
		fmt.Printf("%#02x\n", code)
		inst := opecodes[code]
		c.exec(inst)
	}
}

func (c *CPU) fetch() byte {
	address := c.PC
	c.PC++
	return c.memory[address]
}

func (c *CPU) exec(inst *instruction) {
	switch inst.name {
	case "LDX":
		if inst.mode == "Immediate" {
			c.X = c.fetch()
		}
	case "LDA":
		if inst.mode == "Immediate" {
			c.A = c.fetch()
		}
	case "STA":
		if inst.mode == "Absolute" {
			l, h := uint16(c.fetch()), uint16(c.fetch())
			//fmt.Printf("%#08b,%#08b\n", l, h)
			//fmt.Printf("%#04x\n", l|h<<8)
			c.A = c.memory[l|h<<8]
		}
	case "TSX":
		c.S = c.X
	}
	//fmt.Printf("%#v\n", inst)
}

var opecodes = map[byte]*instruction{
	0x78: {
		name: "SEI",
		mode: "Implied",
	},
	0xA2: {
		name:        "LDX",
		mode:        "Immediate",
		description: "次アドレスの即値をXにロード",
	},
	0xA9: {
		name:        "LDA",
		mode:        "Immediate",
		description: "次アドレスの即値をAにロード",
	},
	0x8D: {
		name:        "STA",
		mode:        "Absolute",
		description: "アドレス「IM16」の8bit値をAにストア",
	},
	0x9A: {
		name:        "TXS",
		mode:        "Implied",
		description: "XをSへコピー",
	},
}
