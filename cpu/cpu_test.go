package cpu

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/yusukemisa/gones/rom"
)

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

		//"INC_Implied(0xF6)": {
		//			in: &CPU{
		//				register: &Register{
		//					X: 0x03,
		//				},
		//				memory: []byte{0xF6, 0x02, 0x00, 0xAA, 0xFE, 0x00},
		//			},
		//			out: &CPU{
		//				register: &Register{
		//					X:  0x03,
		//					PC: 0x01,
		//					P:  0b10000000,
		//				},
		//				memory: []byte{0xF6, 0x02, 0x00, 0xAA, 0xFF, 0x00},
		//			},
		//		},
		//		"BNE_Relative(0xD0)_jump": {
		//			in: &CPU{
		//				register: &Register{
		//					X:  0x03,
		//					PC: 0x03,
		//					P:  0b11111101, // N,Z: not affected
		//				},
		//				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
		//			},
		//			out: &CPU{
		//				register: &Register{
		//					X:  0x03,
		//					PC: 0x02,
		//					P:  0b11111101,
		//				},
		//				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
		//			},
		//		},
		//		"BNE_Relative(0xD0)_no jump": {
		//			in: &CPU{
		//				register: &Register{
		//					X:  0x03,
		//					PC: 0x03,
		//					P:  0b11111111, // N,Z: not affected
		//				},
		//				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
		//			},
		//			out: &CPU{
		//				register: &Register{
		//					X:  0x03,
		//					PC: 0x05,
		//					P:  0b11111111,
		//				},
		//				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
		//			},
		//		},
	} {
		tt := tt
		t.Run(fmt.Sprintf("code=%#02x:%s", tt.opecode, tt.name), func(t *testing.T) {
			rom := &rom.Rom{PRG: tt.param}
			cpu := NewCPU(rom, true)
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
