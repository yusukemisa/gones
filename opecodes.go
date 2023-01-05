package main

var opecodes = map[byte]*instruction{
	0x78: {
		code: 0x78,
		name: "SEI",
		mode: "Implied",
	},
	0x88: {
		code:        0x88,
		name:        "DEY",
		mode:        "Implied",
		description: "Yをデクリメント",
	},
	0x8D: {
		code:        0x8D,
		name:        "STA",
		mode:        "Absolute",
		description: "アドレス「IM16」の8bit値をAにストア",
	},
	0x9A: {
		code:        0x9A,
		name:        "TXS",
		mode:        "Implied",
		description: "XをSへコピー",
	},
	0xA0: {
		code:        0xA0,
		name:        "LDY",
		mode:        "Immediate",
		description: "次アドレスの即値をYにロード",
	},
	0xA2: {
		code:        0xA2,
		name:        "LDX",
		mode:        "Immediate",
		description: "次アドレスの即値をXにロード",
	},
	0xA9: {
		code:        0xA9,
		name:        "LDA",
		mode:        "Immediate",
		description: "次アドレスの即値をAにロード",
	},
	0xBD: {
		code:        0xBD,
		name:        "LDA",
		mode:        "AbsoluteX",
		description: "アドレス「IM16 + X」の8bit値をAにロード",
	},
	0xD0: {
		code:        0xD0,
		name:        "BNE",
		mode:        "Relative",
		description: "Branch on not equal 0. ステータスレジスタのZがクリアされている時に分岐.",
	},
	0xE8: {
		code:        0xE8,
		name:        "INX",
		mode:        "Implied",
		description: "Xをインクリメント",
	},
}
