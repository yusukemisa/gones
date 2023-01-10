package main

var opecodes = map[byte]*instruction{
	0x4C: {
		code:        0x4C,
		name:        "JMP",
		mode:        "Absolute",
		description: "PCをIM16へジャンプ",
		cycle:       3,
		// Z: not affected
		// N: not affected
	},
	0x78: {
		code:  0x78,
		name:  "SEI",
		mode:  "Implied",
		cycle: 2,
		// Z: not affected
		// I: set to 1
		// N: not affected
	},
	0x8D: {
		code:        0x8D,
		name:        "STA",
		mode:        "Absolute",
		description: "Aの内容をアドレス「IM16」に書き込む",
		cycle:       4,
		// Z: not affected
		// N: not affected
	},
	0x9A: {
		code:        0x9A,
		name:        "TXS",
		mode:        "Implied",
		description: "XをSへコピー",
		cycle:       2,
		// Z: not affected
		// N: not affected
	},
	0xA0: {
		code:        0xA0,
		name:        "LDY",
		mode:        "Immediate",
		description: "次アドレスの即値をYにロード",
		cycle:       2,
		// Z:Set if Y = 0
		// N:Set if bit 7 of Y is set
	},
	0xA2: {
		code:        0xA2,
		name:        "LDX",
		mode:        "Immediate",
		description: "次アドレスの即値をXにロード",
		cycle:       2,
		// Z:Set if X = 0
		// N:Set if bit 7 of X is set
	},
	0xA9: {
		code:        0xA9,
		name:        "LDA",
		mode:        "Immediate",
		description: "次アドレスの即値をAにロード",
		cycle:       2,
		// Z:Set if A = 0
		// N:Set if bit 7 of A is set
	},
	0xBD: {
		code:        0xBD,
		name:        "LDA",
		mode:        "AbsoluteX",
		description: "アドレス「IM16 + X」の8bit値をAにロード",
		cycle:       4,
		// Bytes:3
		// Z:Set if A = 0
		// N:Set if bit 7 of A is set
	},
	0xD0: {
		code:        0xD0,
		name:        "BNE",
		mode:        "Relative",
		description: "Branch on not equal 0. ステータスレジスタのZがクリアされている場合アドレス「PC + IM8」へジャンプ",
		cycle:       3, // 4の場合もある
		// Z: not affected
		// N: not affected
	},
	0xE8: {
		code:        0xE8,
		name:        "INX",
		mode:        "Implied",
		description: "Xをインクリメント",
		cycle:       2,
		// Z:Set if X = 0
		// N:Set if bit 7 of X is set
	},
	0x88: {
		code:        0x88,
		name:        "DEY",
		mode:        "Implied",
		description: "Yをデクリメント",
		cycle:       2,
		// Z:Set if Y = 0
		// N:Set if bit 7 of Y is set
	},
	0xF0: {
		code:        0xF0,
		name:        "BEQ",
		mode:        "Relative",
		description: "Branch on equal 0. ステータスレジスタのZがセットされている時に分岐.",
		cycle:       3, // 4の場合もある
		// Z: not affected
		// N: not affected
	},
	0xF6: {
		code:        0xF6,
		name:        "INC",
		mode:        "ZeroPageX",
		description: "Increment Memory by One. アドレス「IM8 + X」の値をインクリメント.",
		cycle:       6,
		// Z:Set if result = 0
		// N:Set if bit 7 of result is set
	},
}
