package cpu

import (
	"fmt"
	"log"

	"github.com/yusukemisa/gones/bus"
	"github.com/yusukemisa/gones/util"
)

type CPU struct {
	register *Register
	bus      *bus.Bus
}

func NewCPU(bus *bus.Bus) *CPU {
	cpu := &CPU{
		register: &Register{},
		bus:      bus,
	}
	return cpu
}

type Register struct {
	A byte // アキュムレータ
	X byte // インデックスレジスタ
	Y byte // インデックスレジスタ

	// スタックポインタ
	// スタックは割り込みが発生する直前まで実行していたプログラムのPCの値を一時的に格納します。
	// 割り込み後、スタックに退避させていたアドレスはPCに戻し処理を再開します。
	// PCは2byte情報なので
	S byte

	// プログラムカウンタ
	// CPUはfetch
	// CPUはfetchでPCのアドレスから命令を読む
	PC uint16

	// ステータスレジスタ
	// 条件付きの分岐命令を実行するために演算結果を保持する
	//bit	名称	詳細	            内容
	//bit7	N	ネガティブ	    演算結果のbit7が1の時にセット
	//bit6	V	オーバーフロー	演算結果がオーバーフローを起こした時にセット
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
	cycle       int
}

func (c *CPU) Reset() {
	// 開始アドレスを取得しPCにセット
	//l, h := uint16(c.read(0xFFFC)), uint16(c.read(0xFFFD))
	//fmt.Printf("%#02x,%#02x\n", l, h)
	//c.register.PC = l | h<<8
	// なんかうまく行かないので固定で0x8000
	c.register.PC = 0x8000
}

// Run is main processing in CPU
func (c *CPU) Run() int {
	code := c.fetch()
	inst, ok := opecodes[code]
	if !ok {
		log.Fatalf("opecode not found:%#02x", code)
	}

	c.exec(inst)
	// 分岐でcycle数変わる場合があるのでexecが返した方が良い
	return inst.cycle
}

func (c *CPU) fetch() byte {
	address := c.register.PC
	c.register.PC++
	return c.bus.Read(address)
}

func (c *CPU) exec(inst *instruction) {
	fmt.Printf("%04X, %#v,\n", c.register.PC-1, inst)
	switch inst.name {
	case "NOP":
	case "JMP":
		l, h := uint16(c.fetch()), uint16(c.fetch())
		c.register.PC = l | h<<8
	case "JSR":
		// 今のPCをスタックに退避し、PC=MI16にする
		l, h := uint16(c.fetch()), uint16(c.fetch())
		c.pushAddressToStack(c.register.PC)
		c.register.PC = l | h<<8
	case "RTS":
		// スタックから戻り番地を取得しPCに格納する
		c.register.PC = c.popAddressFromStack()
	case "SEC":
		c.register.P = util.SetBit(c.register.P, 0)
	case "CLC":
		c.register.P = util.ClearBit(c.register.P, 0)
	case "SED":
		// デシマルモードをON
		// bit3を立てる
		c.register.P = util.SetBit(c.register.P, 3)
	case "SEI":
		// IRQ割り込み禁止
		// bit2を立てる
		c.register.P = util.SetBit(c.register.P, 2)
	case "BIT":
		l, h := c.fetch(), byte(0x00)
		addr := uint16(l | h<<8)
		and := c.register.A & c.read(addr)
		//fmt.Printf("l=%#02x, h=%#02x, addr=%#04x,and=%#02x, A=%#02x\n", l, h, addr, and, c.register.A)
		if util.TestBit(and, 6) {
			c.register.P = util.SetBit(c.register.P, 6)
		} else {
			c.register.P = util.ClearBit(c.register.P, 6)
		}
		c.updateStatusRegister(and)
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
			c.write(l|h<<8, c.register.A)
		}
		if inst.mode == "ZeroPage" {
			l, h := uint16(c.fetch()), uint16(0x00)
			c.write(l|h<<8, c.register.A)
		}
	case "STX":
		if inst.mode == "ZeroPage" {
			l, h := uint16(c.fetch()), uint16(0x00)
			c.write(l|h<<8, c.register.X)
		}
	case "TXS":
		c.register.S = c.register.X
	case "INX":
		c.register.X++
		c.updateStatusRegister(c.register.X)
	case "INC":
		if inst.mode == "ZeroPageX" {
			//fmt.Printf("PC=%#04x,X=%#04x\n", c.register.PC, uint16(c.register.X))
			addr := c.register.PC + uint16(c.register.X)
			c.write(addr, c.read(addr)+1)
			c.updateStatusRegister(c.read(addr))
		}
	case "DEY":
		c.register.Y--
		c.updateStatusRegister(c.register.Y)
	case "BCS":
		if inst.mode == "Relative" {
			relAddr := int8(c.fetch())
			if util.TestBit(c.register.P, 0) {
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	case "BCC":
		if inst.mode == "Relative" {
			relAddr := int8(c.fetch())
			if !util.TestBit(c.register.P, 0) {
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	case "BVS":
		if inst.mode == "Relative" {
			relAddr := int8(c.fetch())
			if util.TestBit(c.register.P, 6) {
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	case "BVC":
		if inst.mode == "Relative" {
			relAddr := int8(c.fetch())
			if !util.TestBit(c.register.P, 6) {
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	case "BPL":
		if inst.mode == "Relative" {
			relAddr := int8(c.fetch())
			if !util.TestBit(c.register.P, 7) {
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	case "BNE":
		if inst.mode == "Relative" {
			// 分岐するしないに関係なくPCが2byte回る必要ある
			relAddr := int8(c.fetch())
			if !util.TestBit(c.register.P, 1) {
				// uint8で取得した値を-128~127の範囲にキャストしてアドレスを計算
				// 0xFFの場合アドレスを-1することになる
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	case "BEQ": // ステータスレジスタのZがセットされている場合アドレス「PC + IM8」へジャンプ"
		if inst.mode == "Relative" {
			relAddr := int8(c.fetch())
			if util.TestBit(c.register.P, 1) {
				addr := int(relAddr) + int(c.register.PC)
				c.register.PC = uint16(addr)
			}
		}
	default:
		fmt.Printf("unknown code:%#v\n", inst)
	}
	//fmt.Printf("A:%#02x,X:%#02x,Y:%#02x,PC:%#04x\n", c.register.A, c.register.X, c.register.Y, c.register.PC)
}

func (c *CPU) write(address uint16, data byte) {
	c.bus.Write(address, data)
}

func (c *CPU) read(address uint16) byte {
	return c.bus.Read(address)
}

// updateStatusRegister updates status register.
// bit	名称	詳細	            内容
// bit7	N	ネガティブ	    演算結果のbit7が1の時にセット
// bit6	V	オーバーフロー  	演算結果がオーバーフローを起こした時にセット
// bit1	Z	ゼロ	            演算結果が0の時にセット
// TODO: 他のbitは0にも1にも更新しなくて良い？
func (c *CPU) updateStatusRegister(result byte) {
	// bit1	Z
	if result == 0 {
		c.register.P = util.SetBit(c.register.P, 1)
	} else {
		c.register.P = util.ClearBit(c.register.P, 1)
	}

	// Bit7 N
	// Aの最上部bitの値とのORをとる
	if util.TestBit(result, 7) {
		c.register.P = util.SetBit(c.register.P, 7)
	} else {
		c.register.P = util.ClearBit(c.register.P, 7)
	}
	//fmt.Printf("result=%#02x,Z=%v,N=%v\n", result, testBit(c.register.P, 1), testBit(c.register.P, 7))
}

func (c *CPU) pushAddressToStack(address uint16) {
	l, h := address&0x00FF, address>>8
	c.write(0x0100+uint16(c.register.S), byte(h))
	c.register.S++
	c.write(0x0100+uint16(c.register.S), byte(l))
}

func (c *CPU) popAddressFromStack() uint16 {
	l := uint16(c.read(0x0100 + uint16(c.register.S)))
	c.register.S--
	h := uint16(c.read(0x0100 + uint16(c.register.S))) // l|h<<8
	return l | h<<8
}
