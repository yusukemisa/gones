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
	pc       uint16
	nmi      bool
	debug    bool
}

func NewCPU(bus *bus.Bus, debug bool) *CPU {
	cpu := &CPU{
		register: &Register{
			P: 0b0010_0000,
		},
		bus:   bus,
		debug: debug,
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
	// リセットベクタ（0xFFFC-0xFFFD）から開始アドレスを読み込む
	l := uint16(c.bus.Read(0xFFFC))
	h := uint16(c.bus.Read(0xFFFD))
	if c.debug {
		fmt.Printf("Reading Reset Vector at fffc: %02x\n", l)
		fmt.Printf("Reading Reset Vector at fffd: %02x\n", h)
		fmt.Printf("Reset Vector: low=%02x, high=%02x\n", l, h)
	}
	c.register.PC = (h << 8) | l
	c.register.S = 0xFF // スタックポインタを初期化
	c.register.P = 0x34 // 割り込み禁止フラグとブレークフラグを設定

	if c.debug {
		fmt.Printf("CPU Reset: PC=%04x\n", c.register.PC)
	}
}

// Run is main processing in CPU
func (c *CPU) Run() int {
	code := c.fetch()
	if c.debug {
		//fmt.Printf("CPU: PC=%04x, opcode=%02x\n", c.register.PC, code)
	}

	inst, ok := opecodes[code]
	if !ok {
		log.Fatalf("opecode not found:%#02x", code)
	}

	c.exec(inst)
	return inst.cycle
}

func (c *CPU) fetch() byte {
	address := c.register.PC
	c.register.PC++
	return c.bus.Read(address)
}

func (c *CPU) exec(inst *instruction) {
	// fmt.Printf("%04X, %#v,\n", c.register.PC-1, inst)
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
	case "PHP":
		// ステータスのコピーをスタックに退避
		c.pushByteToStack(c.register.P)
	case "PHA":
		// アキュムレーターのコピーをスタックに退避
		c.pushByteToStack(c.register.A)
	case "PLA":
		// スタックからAにPull
		c.register.A = c.popByteFromStack()
		c.updateStatusRegister(c.register.A)
	case "PLP":
		// スタックからPにPull
		c.register.P = c.popByteFromStack()
	case "AND":
		if inst.mode == "Immediate" {
			c.register.A = c.register.A & c.fetch()
			c.updateStatusRegister(c.register.A)
		}
	case "CMP":
		if inst.mode == "Immediate" {
			result := c.register.A - c.fetch()
			if result >= 0 {
				c.register.P = util.SetBit(c.register.P, 0)
			}
			c.updateStatusRegister(result)
		}
	case "SEC":
		c.register.P = util.SetBit(c.register.P, 0)
	case "CLC":
		c.register.P = util.ClearBit(c.register.P, 0)
	case "CLD":
		// デシマルモードをOFF
		// bit3を消す
		c.register.P = util.ClearBit(c.register.P, 3)
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
	case "BMI":
		if inst.mode == "Relative" {
			relAddr := int8(c.fetch())
			if util.TestBit(c.register.P, 7) {
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
	case "BRK":
		// PCをインクリメント
		c.register.PC++

		// PCとステータスレジスタをスタックにプッシュ
		c.pushAddressToStack(c.register.PC)
		c.pushByteToStack(c.register.P | 0x10) // ブレークフラグを設定

		// 割り込みベクタ（0xFFFE-0xFFFF）からアドレスを読み込む
		l := uint16(c.bus.Read(0xFFFE))
		h := uint16(c.bus.Read(0xFFFF))
		c.register.PC = l | h<<8

		// 割り込み禁止フラグを設定
		c.register.P = util.SetBit(c.register.P, 2)
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

func (c *CPU) pushByteToStack(b byte) {
	c.write(0x0100+uint16(c.register.S), b)
	c.register.S++
}

func (c *CPU) pushAddressToStack(address uint16) {
	l, h := address&0x00FF, address>>8
	c.register.S++
	c.write(0x00FF+uint16(c.register.S), byte(h))
	c.register.S++
	c.write(0x00FF+uint16(c.register.S), byte(l))
	//fmt.Printf("pushAddressToStack:S=%#02x,l=%#02x,h=%#02x\n", c.register.S, l, h)
}

func (c *CPU) popAddressFromStack() uint16 {
	l := uint16(c.read(0x00FF + uint16(c.register.S)))
	c.register.S--
	h := uint16(c.read(0x00FF + uint16(c.register.S))) // l|h<<8
	c.register.S--
	//fmt.Printf("popAddressFromStack:S=%#02x,l=%#02x,h=%#02x\n", c.register.S, l, h)
	return l | h<<8
}

func (c *CPU) popByteFromStack() byte {
	b := c.read(0x0100 + uint16(c.register.S-1))
	c.register.S--
	return b
}

// TriggerNMIはNMI割り込みを発生させる
func (c *CPU) TriggerNMI() {
	// 現在のPCをスタックに退避
	c.pushAddressToStack(c.register.PC)
	// ステータスレジスタをスタックに退避
	c.pushByteToStack(c.register.P)
	// 割り込み禁止フラグを設定
	c.register.P = util.SetBit(c.register.P, 2)
	// NMIベクタ（0xFFFA-0xFFFB）から割り込みハンドラのアドレスを読み込む
	l := uint16(c.bus.Read(0xFFFA))
	h := uint16(c.bus.Read(0xFFFB))
	c.register.PC = (h << 8) | l
}
