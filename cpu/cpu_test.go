package cpu

import (
	"reflect"
	"testing"
)

func TestCPU_status(t *testing.T) {
	t.Parallel()
	for title, tt := range map[string]struct {
		in  *CPU
		out *CPU
	}{
		"LDA(0xA9) set N": {
			in: &CPU{memory: []byte{0xA9, 0xFF}},
			out: &CPU{
				register: &Register{
					A:  0xFF,
					PC: 0x02,
					P:  0b10000000,
				},
				memory: []byte{0xA9, 0xFF},
			},
		},
		"LDA(0xA9) set Z": {
			in: &CPU{memory: []byte{0xA9, 0x00}},
			out: &CPU{
				register: &Register{
					A:  0x00,
					PC: 0x02,
					P:  0b00000010,
				},
				memory: []byte{0xA9, 0x00},
			},
		},
		"LDA_AbsoluteX(0xBD) set N": {
			in: &CPU{
				register: &Register{
					X: 0x01,
				},
				memory: []byte{0xBD, 0x03, 0x00, 0x00, 0xFF, 0x00},
			},
			out: &CPU{
				register: &Register{
					A:  0xFF,
					X:  0x01,
					PC: 0x03,
					P:  0b10000000,
				},
				memory: []byte{0xBD, 0x03, 0x00, 0x00, 0xFF, 0x00},
			},
		},
		"LDA_AbsoluteX(0xBD) set Z": {
			in: &CPU{
				register: &Register{
					X: 0x02,
				},
				memory: []byte{0xBD, 0x03, 0x00, 0x00, 0xFF, 0x00},
			},
			out: &CPU{
				register: &Register{
					A:  0x00,
					X:  0x02,
					PC: 0x03,
					P:  0b00000010,
				},
				memory: []byte{0xBD, 0x03, 0x00, 0x00, 0xFF, 0x00},
			},
		},
		"STA_Absolute(0x8D)": {
			in: &CPU{
				register: &Register{
					A: 0xAA,
					X: 0x02,
				},
				memory: []byte{0x8D, 0x03, 0x00, 0xFF, 0x00},
			},
			out: &CPU{
				register: &Register{
					A:  0xAA,
					X:  0x02,
					PC: 0x03,
					P:  0b00000000, //N,Z: not affected
				},
				memory: []byte{0x8D, 0x03, 0x00, 0xAA, 0x00},
			},
		},
		"TXS_Implied(0x9A)": {
			in: &CPU{
				register: &Register{
					X: 0xFF,
				},
				memory: []byte{0x9A, 0x03, 0x00, 0xFF, 0x00},
			},
			out: &CPU{
				register: &Register{
					X:  0xFF,
					S:  0xFF,
					PC: 0x01,
					P:  0b00000000, //N,Z: not affected
				},
				memory: []byte{0x9A, 0x03, 0x00, 0xFF, 0x00},
			},
		},
		"LDY_Immediate(0xA0)": {
			in: &CPU{
				register: &Register{
					X: 0x03,
					Y: 0x03,
				},
				memory: []byte{0xA0, 0xFF},
			},
			out: &CPU{
				register: &Register{
					X:  0x03,
					Y:  0xFF,
					PC: 0x02,
					P:  0b10000000,
				},
				memory: []byte{0xA0, 0xFF},
			},
		},
		"LDX_Immediate(0xA2)": {
			in: &CPU{
				register: &Register{
					X: 0x03,
					Y: 0x03,
				},
				memory: []byte{0xA2, 0x00},
			},
			out: &CPU{
				register: &Register{
					X:  0x00,
					Y:  0x03,
					PC: 0x02,
					P:  0b00000010,
				},
				memory: []byte{0xA2, 0x00},
			},
		},
		"INX_Implied(0xE8)": {
			in: &CPU{
				register: &Register{
					X: 0xFE,
					Y: 0x03,
				},
				memory: []byte{0xE8},
			},
			out: &CPU{
				register: &Register{
					X:  0xFF,
					Y:  0x03,
					PC: 0x01,
					P:  0b10000000,
				},
				memory: []byte{0xE8},
			},
		},
		"INY_Implied(0x88)": {
			in: &CPU{
				register: &Register{
					X: 0x03,
					Y: 0x01,
				},
				memory: []byte{0x88},
			},
			out: &CPU{
				register: &Register{
					X:  0x03,
					Y:  0x00,
					PC: 0x01,
					P:  0b00000010,
				},
				memory: []byte{0x88},
			},
		},
		"INC_Implied(0xF6)": {
			in: &CPU{
				register: &Register{
					X: 0x03,
				},
				memory: []byte{0xF6, 0x02, 0x00, 0xAA, 0xFE, 0x00},
			},
			out: &CPU{
				register: &Register{
					X:  0x03,
					PC: 0x01,
					P:  0b10000000,
				},
				memory: []byte{0xF6, 0x02, 0x00, 0xAA, 0xFF, 0x00},
			},
		},
		"BNE_Relative(0xD0)_jump": {
			in: &CPU{
				register: &Register{
					X:  0x03,
					PC: 0x03,
					P:  0b11111101, // N,Z: not affected
				},
				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
			},
			out: &CPU{
				register: &Register{
					X:  0x03,
					PC: 0x02,
					P:  0b11111101,
				},
				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
			},
		},
		"BNE_Relative(0xD0)_no jump": {
			in: &CPU{
				register: &Register{
					X:  0x03,
					PC: 0x03,
					P:  0b11111111, // N,Z: not affected
				},
				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
			},
			out: &CPU{
				register: &Register{
					X:  0x03,
					PC: 0x05,
					P:  0b11111111,
				},
				memory: []byte{0x00, 0x00, 0x00, 0xD0, 0xFD, 0x00, 0xAA, 0xBB, 0xCC},
			},
		},
	} {
		tt := tt
		tt.in.debug, tt.out.debug = true, true
		if tt.in.register == nil {
			tt.in.register = &Register{}
		}
		t.Run(title, func(t *testing.T) {
			opecode := tt.in.fetch()
			tt.in.exec(opecodes[opecode])

			if !reflect.DeepEqual(tt.in, tt.out) {
				t.Errorf("\nwant=%#v\n got=%#v", tt.out, tt.in)
			}
		})
	}
}
