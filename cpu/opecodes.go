package cpu

var opecodes = map[byte]*instruction{
	0x00: {
		code: 0x00,
		name: "BRK",
		mode: "Implied",
		description: "BRK命令は、強制的に割り込み要求を発生させます。プログラムカウンタとプロセッサステータスがスタックにプッシュされ、" +
			"$FFFE/FのIRQ割り込みベクタがPCにロードされ、ステータス内のブレークフラグが1にセットされます。",
		cycle: 7,
		// Z: not affected
		// N: not affected
		// B: Set to 1
	},
	0x08: {
		code:        0x08,
		name:        "PHP", // Push Processor Status
		mode:        "Implied",
		description: "Pushes a copy of the status flags on to the stack.",
		cycle:       3,
		// Z: not affected
		// N: not affected
		// bytes:1
	},
	0x68: {
		code:        0x68,
		name:        "PLA", // Pull Accumulator
		mode:        "Implied",
		description: "Pulls an 8 bit value from the stack and into the accumulator. The zero and negative flags are set as appropriate.",
		cycle:       4,
		// Z: Set if A = 0
		// N: Set if bit 7 of A is set
		// bytes:1
	},
	0x10: {
		code:        0x10,
		name:        "BPL", // Branch if Positive
		mode:        "Relative",
		description: "ステータスレジスタのNがクリアされている場合アドレス「PC + IM8」へジャンプ",
		cycle:       2, // 2 (+1 if branch succeeds +2 if to a new page)
		// Z: not affected
		// N: not affected
		// bytes:2
	},
	0x18: {
		code:        0x18,
		name:        "CLC", // Clear carry flag
		mode:        "Implied",
		description: "Set the carry flag to 0",
		cycle:       2,
		// Z: not affected
		// N: not affected
		// C: set to 0
		// bytes:1
	},
	0x20: {
		code:        0x20,
		name:        "JSR", // Jump to subroutine
		mode:        "Absolute",
		description: "サブルーチンを呼び出し",
		cycle:       6,
		// Z: not affected
		// N: not affected
		// bytes:3
	},
	0x60: {
		code:        0x60,
		name:        "RTS", // Return from Subroutine
		mode:        "Implied",
		description: "サブルーチンから復帰",
		cycle:       6,
		// Z: not affected
		// N: not affected
		// bytes:1
	},
	0x24: {
		code:        0x24,
		name:        "BIT", // Bit Test
		mode:        "ZeroPage",
		description: "Aと0x00IM8番地の値をビット比較演算します",
		cycle:       3,
		// Z: Set if the result if the AND is zero
		// V: Set to bit 6 of the memory value
		// N: Set to bit 7 of the memory value
		// bytes:2
	},
	0x38: {
		code:        0x38,
		name:        "SEC",
		mode:        "Implied",
		description: "Set carry flag",
		cycle:       6,
		// Z: not affected
		// N: not affected
		// bytes:1
	},
	0x4C: {
		code:        0x4C,
		name:        "JMP",
		mode:        "Absolute",
		description: "PCをIM16へジャンプ",
		cycle:       3,
		// Z: not affected
		// N: not affected
	},
	0x50: {
		code:        0x50,
		name:        "BVC", // Branch if Overflow Clear
		mode:        "Relative",
		description: "ステータスレジスタのVがクリアされている場合アドレス「PC + IM8」へジャンプ",
		cycle:       2, // 2 (+1 if branch succeeds +2 if to a new page)
		// Z: not affected
		// N: not affected
		// bytes:2
	},
	0x70: {
		code:        0x70,
		name:        "BVS", // Branch if Overflow Set
		mode:        "Relative",
		description: "ステータスレジスタのVがセットされている場合アドレス「PC + IM8」へジャンプ",
		cycle:       2, // 2 (+1 if branch succeeds +2 if to a new page)
		// Z: not affected
		// N: not affected
		// bytes:2
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
	0x86: {
		code:        0x86,
		name:        "STX",
		mode:        "ZeroPage", // 0x00を上位アドレス、PCに格納された値を下位アドレスとした番地を演算対象とする
		description: "Stores the contents of the X register into memory",
		cycle:       3,
		// Z: not affected
		// N: not affected
		// bytes:2
	},
	0x90: {
		code:        0x90,
		name:        "BCC", // Branch if Carry Clear
		mode:        "Relative",
		description: "If the carry flag is clear then add the relative displacement to the program counter to cause a branch to a new location.",
		cycle:       2, // 2 (+1 if branch succeeds +2 if to a new page)
		// Z: not affected
		// N: not affected
		// bytes:2
	},
	0x85: {
		code:        0x85,
		name:        "STA",
		mode:        "ZeroPage",
		description: "Aの内容をアドレス「MI8 | 0x00<<8 」に書き込む",
		cycle:       3,
		// Z: not affected
		// N: not affected
		// bytes:2
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
	0xB0: {
		code:        0xB0,
		name:        "BCS", // Branch if Carry Set
		mode:        "Relative",
		description: "If the carry flag is set then add the relative displacement to the program counter to cause a branch to a new location",
		cycle:       2,
		// Z: not affected
		// N: not affected
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
	0xEA: {
		code:        0xEA,
		name:        "NOP",
		mode:        "Implied",
		description: "No operation",
		cycle:       2,
		// Z: not affected
		// N: not affected
		// Bytes:1
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
		description: "Branch on equal 0. ステータスレジスタのZがセットされている場合アドレス「PC + IM8」へジャンプ",
		cycle:       3, // 2 (+1 if branch succeeds +2 if to a new page)
		// Z: not affected
		// N: not affected
		// bytes:2
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
	0xF8: {
		code:        0xF8, // Set Decimal Flag
		name:        "SED",
		mode:        "Implied",
		description: "Set the decimal mode flag to one.",
		cycle:       2,
		// Z: not affected
		// N: not affected
		// D: Set to 1
		// bytes:1
	},
}
