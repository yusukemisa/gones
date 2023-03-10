package cpu

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"

	"github.com/yusukemisa/gones/bus"
	"github.com/yusukemisa/gones/rom"
)

func TestCPU_memory(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		opecode      byte
		name         string
		param        []byte
		init         func(cpu *CPU)
		wantRegister *Register
		address      []uint16
		wantData     []byte
	}{
		{
			opecode: 0x20,
			name:    "JSR",
			param:   []byte{0x10, 0x80},
			init: func(cpu *CPU) {
				cpu.register.PC = 0x8000
				cpu.register.P = 0b0000_0000
			},
			wantRegister: &Register{
				PC: 0x8010,
				S:  0x02,
			},
			address:  []uint16{0x0100, 0x0101},
			wantData: []byte{0x80, 0x02},
		},
		{
			opecode: 0x60,
			name:    "RTS",
			param:   []byte{},
			init: func(cpu *CPU) {
				cpu.register.PC = 0x8100
				cpu.register.P = 0b0000_0000
				cpu.pushAddressToStack(0x8010)
			},
			wantRegister: &Register{
				PC: 0x8010,
			},
			address:  []uint16{0x0100, 0x0101},
			wantData: []byte{0x80, 0x10},
		},
		{
			opecode: 0x08,
			name:    "PHP", // ステータスのコピーをスタックに退避
			param:   []byte{},
			init: func(cpu *CPU) {
				cpu.register.PC = 0x8000
				cpu.register.P = 0b0101_0101
			},
			wantRegister: &Register{
				PC: 0x8000,
				P:  0b0101_0101,
				S:  0x01,
			},
			address:  []uint16{0x0100},
			wantData: []byte{0b0101_0101},
		},
		{
			opecode: 0x48,
			name:    "PHA", // アキュムレーターのコピーをスタックに退避
			param:   []byte{},
			init: func(cpu *CPU) {
				cpu.register.PC = 0x8000
				cpu.register.A = 0x10
				cpu.register.P = 0b0000_0000
			},
			wantRegister: &Register{
				PC: 0x8000,
				A:  0x10,
				S:  0x01,
			},
			address:  []uint16{0x0100},
			wantData: []byte{0x10},
		},
		{
			opecode: 0x68,
			name:    "PLA", // スタックからAにpull
			param:   []byte{},
			init: func(cpu *CPU) {
				cpu.register.PC = 0x8000
				cpu.register.A = 0xFF
				cpu.register.P = 0b0000_0000
				cpu.register.S = 0x10
				cpu.pushByteToStack(0x20)
			},
			wantRegister: &Register{
				PC: 0x8000,
				A:  0x20,
				S:  0x10,
			},
			address:  []uint16{0x0110},
			wantData: []byte{0x20},
		},
		{
			opecode: 0x28,
			name:    "PLP", // スタックからPにpull
			param:   []byte{},
			init: func(cpu *CPU) {
				cpu.register.PC = 0x8000
				cpu.register.P = 0b0000_0000
				cpu.register.S = 0x10
				cpu.pushByteToStack(0b1111_0000)
			},
			wantRegister: &Register{
				PC: 0x8000,
				P:  0b1111_0000,
				S:  0x10,
			},
			address:  []uint16{0x0110},
			wantData: []byte{0b1111_0000},
		},
	} {
		tt := tt
		t.Run(fmt.Sprintf("code=%#02x:%s", tt.opecode, tt.name), func(t *testing.T) {
			rom := &rom.Rom{PRG: tt.param}

			cpu := NewCPU(bus.NewBus(rom, nil))
			tt.init(cpu)

			cpu.exec(opecodes[tt.opecode])
			if diff := cmp.Diff(tt.wantRegister, cpu.register); diff != "" {
				t.Errorf("register mismatch (-want +got):\n%s", diff)
			}

			if len(tt.address) != 0 {
				for i, addr := range tt.address {
					if diff := cmp.Diff(tt.wantData[i], cpu.read(addr)); diff != "" {
						t.Errorf("memory data mismatch (-want +got):\n%s", diff)
					}
				}
			}
		})
	}
}

func TestCPU_exec(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		opecode      byte
		name         string
		param        []byte
		orgRegister  *Register
		wantRegister *Register
		address      uint16
		wantData     byte
	}{
		{
			opecode:      0x4C,
			name:         "JMP",
			param:        []byte{0xFF, 0x80},
			orgRegister:  &Register{PC: 0x8000},
			wantRegister: &Register{PC: 0x80FF},
		},
		{
			opecode: 0x29,
			name:    "AND",
			param:   []byte{0b1111_0000},
			orgRegister: &Register{
				PC: 0x8000,
				A:  0xFF,
			},
			wantRegister: &Register{
				PC: 0x8001,
				A:  0b1111_0000,
				P:  0b10000000,
			},
		},
		{
			opecode: 0xC9,
			name:    "CMP",
			param:   []byte{0x01},
			orgRegister: &Register{
				PC: 0x8000,
				A:  0x01,
			},
			wantRegister: &Register{
				PC: 0x8001,
				A:  0x01,
				P:  0b00000011,
			},
		},
		{
			opecode: 0x18,
			name:    "CLC",
			param:   []byte{},
			orgRegister: &Register{
				PC: 0x8000,
				P:  0b00000001,
			},
			wantRegister: &Register{
				PC: 0x8000,
				P:  0b00000000,
			},
		},
		{
			opecode:     0x38,
			name:        "SEC",
			param:       []byte{},
			orgRegister: &Register{PC: 0x8000},
			wantRegister: &Register{
				PC: 0x8000,
				P:  0b00000001,
			},
		},
		{
			opecode: 0xD8,
			name:    "CLD",
			param:   []byte{},
			orgRegister: &Register{
				PC: 0x8000,
				P:  0b00001000,
			},
			wantRegister: &Register{
				PC: 0x8000,
				P:  0b00000000,
			},
		},
		{
			opecode:     0xF8,
			name:        "SED",
			param:       []byte{},
			orgRegister: &Register{PC: 0x8000},
			wantRegister: &Register{
				PC: 0x8000,
				P:  0b00001000,
			},
		},
		{
			opecode:     0x78,
			name:        "SEI",
			param:       []byte{},
			orgRegister: &Register{PC: 0x8000},
			wantRegister: &Register{
				PC: 0x8000,
				P:  0b00000100,
			},
		},
		{
			opecode: 0x24,
			name:    "BIT", // TODO: 正しいか怪しい・・・
			param:   []byte{0x01},
			orgRegister: &Register{
				PC: 0x8000,
				A:  0b11111111,
				P:  0b10100100,
			},
			wantRegister: &Register{
				PC: 0x8001,
				A:  0b1111_1111,
				P:  0b00100110,
			},
		}, // 10100100
		{
			opecode:     0xA9,
			name:        "LDA(set N)",
			param:       []byte{0xFF},
			orgRegister: &Register{PC: 0x8000},
			wantRegister: &Register{
				A:  0xFF,
				PC: 0x8001,
				P:  0b10000000,
			},
		},
		{
			opecode:     0xA9,
			name:        "LDA(set zero flag)",
			param:       []byte{0x00},
			orgRegister: &Register{PC: 0x8000},
			wantRegister: &Register{
				A:  0x00,
				PC: 0x8001,
				P:  0b00000010,
			},
		},
		{
			opecode: 0xBD, // アドレス「IM16 + X」の8bit値をAにロード
			name:    "LDA_AbsoluteX(set N)",
			param:   []byte{0x03, 0x80, 0x00, 0x00, 0xFF},
			orgRegister: &Register{
				PC: 0x8000,
				X:  0x01,
			},
			wantRegister: &Register{
				A:  0xFF,
				X:  0x01,
				PC: 0x8002,
				P:  0b10000000,
			},
		},
		{
			opecode: 0xBD, // アドレス「IM16 + X」の8bit値をAにロード
			name:    "LDA_AbsoluteX(set Z)",
			param:   []byte{0x03, 0x80, 0x00, 0xFF, 0x00},
			orgRegister: &Register{
				PC: 0x8000,
				X:  0x01,
			},
			wantRegister: &Register{
				A:  0x00,
				X:  0x01,
				PC: 0x8002,
				P:  0b00000010,
			},
		},
		{
			opecode: 0x8D, // Aの内容をアドレス「IM16」に書き込む.
			name:    "STA_Absolute",
			param:   []byte{0x03, 0x00},
			orgRegister: &Register{
				PC: 0x8000,
				A:  0xAA,
			},
			wantRegister: &Register{
				A:  0xAA,
				PC: 0x8002,
				P:  0b00000000, //N,Z: not affected
			},
			address:  0x0003,
			wantData: 0xAA,
		},
		{
			opecode: 0x85, // Aの内容をアドレス「MI8 | 0x00<<8 」に書き込む.
			name:    "STA_ZeroPage",
			param:   []byte{0x03},
			orgRegister: &Register{
				PC: 0x8000,
				A:  0xAA,
			},
			wantRegister: &Register{
				A:  0xAA,
				PC: 0x8001,
				P:  0b00000000, //N,Z: not affected
			},
			address:  0x0003,
			wantData: 0xAA,
		},
		{
			opecode: 0x86, // Xの内容をアドレス「MI8 | 0x00<<8 」に書き込む.
			name:    "STX_ZeroPage",
			param:   []byte{0x86},
			orgRegister: &Register{
				PC: 0x8000,
				X:  0xFF,
			},
			wantRegister: &Register{
				X:  0xFF,
				PC: 0x8001,
				P:  0b00000000, //N,Z: not affected
			},
			address:  0x0086,
			wantData: 0xFF,
		},
		{
			opecode: 0x9A, // XをSへコピー
			name:    "TXS_Implied",
			param:   []byte{},
			orgRegister: &Register{
				PC: 0x8000,
				X:  0xFF,
			},
			wantRegister: &Register{
				X:  0xFF,
				S:  0xFF,
				PC: 0x8000,
				P:  0b00000000, //N,Z: not affected
			},
		},
		{
			opecode: 0xA0, // 次アドレスの即値をYにロード
			name:    "LDY_Immediate",
			param:   []byte{0xFF},
			orgRegister: &Register{
				PC: 0x8000,
			},
			wantRegister: &Register{
				Y:  0xFF,
				PC: 0x8001,
				P:  0b10000000,
			},
		},
		{
			opecode: 0xA2, // 次アドレスの即値をXにロード
			name:    "LDX_Immediate",
			param:   []byte{0x00},
			orgRegister: &Register{
				PC: 0x8000,
			},
			wantRegister: &Register{
				X:  0x00,
				PC: 0x8001,
				P:  0b00000010,
			},
		},
		{
			opecode: 0xE8, // Xをインクリメント
			name:    "INX_Implied",
			param:   []byte{},
			orgRegister: &Register{
				X:  0xF0,
				PC: 0x8000,
			},
			wantRegister: &Register{
				X:  0xF1,
				PC: 0x8000,
				P:  0b10000000,
			},
		},
		{
			opecode: 0x88, // Yをデクリメント
			name:    "DEY_Implied",
			param:   []byte{},
			orgRegister: &Register{
				Y:  0x00,
				PC: 0x8000,
			},
			wantRegister: &Register{
				Y:  0xFF,
				PC: 0x8000,
				P:  0b10000000,
			},
		},
		{
			opecode: 0xF6, // Increment Memory by One. アドレス「IM8 + X」の値をインクリメント.
			name:    "INC_ZeroPageX",
			param:   []byte{},
			orgRegister: &Register{
				X: 0x01,
			},
			wantRegister: &Register{
				X:  0x01,
				PC: 0x0000,
			},
			address:  0x0001,
			wantData: 0x01,
		},
		{
			opecode: 0xD0, // ステータスレジスタのZがクリアされている場合アドレス「PC + IM8」へジャンプ
			name:    "BNE_Relative",
			param:   []byte{0x0A},
			orgRegister: &Register{
				P:  0b11111101, // only Z cleared
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b11111101,
				PC: 0x800B,
			},
		},
		{
			opecode: 0xD0, // ステータスレジスタのZがクリアされている場合アドレス「PC + IM8」へジャンプ
			name:    "BNE_Relative_no_jump",
			param:   []byte{0x0A},
			orgRegister: &Register{
				P:  0b11111111,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b11111111,
				PC: 0x8001,
			},
		},
		{
			opecode: 0xB0, // キャリーフラグがセットされている場合アドレス「PC + IM8」へ分岐
			name:    "BCS_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b11111111,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b11111111,
				PC: 0x8011,
			},
		},
		{
			opecode: 0xB0, // キャリーフラグがセットされている場合アドレス「PC + IM8」へ分岐
			name:    "BCS_not_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b00000000,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b00000000,
				PC: 0x8001,
			},
		},
		{
			opecode: 0x90, // キャリーフラグがクリアされている場合アドレス「PC + IM8」へ分岐
			name:    "BCC_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b00000000,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b00000000,
				PC: 0x8011,
			},
		},
		{
			opecode: 0x90,
			name:    "BCC_not_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b11111111,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b11111111,
				PC: 0x8001,
			},
		},
		{
			opecode: 0xF0, // ステータスレジスタのZがセットされている場合アドレス「PC + IM8」へジャンプ",
			name:    "BEQ_Relative_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b0000_0010,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b0000_0010,
				PC: 0x8011,
			},
		},
		{
			opecode: 0xF0,
			name:    "BEQ_Relative_not_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b0000_0000,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b0000_0000,
				PC: 0x8001,
			},
		},
		{
			opecode: 0x70, // ステータスレジスタのVがセットされている場合アドレス「PC + IM8」へジャンプ",
			name:    "BVS_Relative_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b0100_0000,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b0100_0000,
				PC: 0x8011,
			},
		},
		{
			opecode: 0x70, // ステータスレジスタのVがセットされている場合アドレス「PC + IM8」へジャンプ",
			name:    "BVS_Relative_not_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b0000_0000,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b0000_0000,
				PC: 0x8001,
			},
		},
		{
			opecode: 0x30, // Branch if Minus
			name:    "BMI_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b1000_0000,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b1000_0000,
				PC: 0x8011,
			},
		},
		{
			opecode: 0x10, // Branch if Positive
			name:    "BPL_branch",
			param:   []byte{0x10},
			orgRegister: &Register{
				P:  0b0000_0000,
				PC: 0x8000,
			},
			wantRegister: &Register{
				P:  0b0000_0000,
				PC: 0x8011,
			},
		},
	} {
		tt := tt
		t.Run(fmt.Sprintf("code=%#02x:%s", tt.opecode, tt.name), func(t *testing.T) {
			rom := &rom.Rom{PRG: tt.param}

			cpu := NewCPU(bus.NewBus(rom, nil))
			cpu.register = tt.orgRegister

			cpu.exec(opecodes[tt.opecode])
			if diff := cmp.Diff(tt.wantRegister, cpu.register); diff != "" {
				t.Errorf("register mismatch (-want +got):\n%s", diff)
			}

			if tt.address != 0 && tt.wantData != 0 {
				if diff := cmp.Diff(tt.wantData, cpu.read(tt.address)); diff != "" {
					t.Errorf("memory data mismatch (-want +got):\n%s", diff)
				}
			}
		})
	}
}
