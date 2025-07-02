package ppu

import (
	"fmt"
	"image/color"
	"os"

	"github.com/yusukemisa/gones/canvas"
)

const (
	windowWidth, windowHeight = 256, 240
	// PPUメモリマップの定数
	patternTable0Start = 0x0000
	patternTable1Start = 0x1000
	nameTable0Start    = 0x2000
	nameTable1Start    = 0x2400
	nameTable2Start    = 0x2800
	nameTable3Start    = 0x2C00
	paletteStart       = 0x3F00
	// PPUの初期化待機時間（NTSCサイクル）
	InitWaitCyclesNTSC = 29658
	InitWaitCyclesPAL  = 33132
)

type register struct {
	// Control
	CTRL    byte // 0x2000 割り込み制御
	MASK    byte // 0x2001 レンダリング制御
	STATUS  byte // 0x2002 PPU状態
	OAMADDR byte // 0x2003 OAMアドレス
	OAMDATA byte // 0x2004 OAMデータ
	SCROLL  byte // 0x2005 スクロール
	ADDRESS byte // 0x2006 PPUアドレス
	DATA    byte // 0x2007 PPUデータ

	// スクロール関連の内部レジスタ
	scrollX, scrollY byte
	writeToggle      bool // スクロール書き込みのトグル
}

type AddressRegister struct {
	high, low byte
}

// Screen represents a screen to be displayed in a window.
// Window here means the window of this application viewed directly by the user on the monitor.
type Screen struct {
	tile  []Tile
	pixel []byte
}

type Tile struct {
	x, y int
	c    color.RGBA
}

// CPUInterfaceはPPUが必要とするCPUの機能を定義
type CPUInterface interface {
	TriggerNMI()
}

// PPUはPicture Processing Unitを表す
type PPU struct {
	cycle int
	line  int
	// CPUに読ませるのはこちらの内部バッファ.
	// 直接PPUメモリやROMから読んだ内容にアクセスさせない
	internalDataBuf byte
	address         *AddressRegister
	register        *register
	memory          []byte
	sprites         map[int][]byte
	tiles           []*Tile
	Canvas          canvas.Canvas
	oam             []byte
	secondaryOAM    []byte // セカンダリOAM
	debug           bool
	// スプライト関連
	spriteZeroHit  bool
	spriteOverflow bool
	// レンダリング状態
	renderingEnabled bool
	vblank           bool
	// 初期化関連
	initCycles         int
	initialized        bool
	InitWaitCyclesNTSC int // 初期化待機時間（NTSCサイクル）
	// 内部バス
	ioLatch     byte
	internalBus byte // 内部バス
	busDecay    int  // バス減衰カウンタ
	readBuffer  byte // 読み込みバッファ
	spriteCount int  // 現在のスプライト数
	nmi         bool // NMI割り込みフラグ
	scanline    int  // 現在のスキャンライン
	// 内部状態
	v     uint16       // VRAMアドレス
	t     uint16       // 一時VRAMアドレス
	x     byte         // ファインXスクロール
	w     bool         // 書き込みフラグ
	cpu   CPUInterface // CPUへの参照を追加
	frame uint64
}

// NewPPUは新しいPPUインスタンスを作成
func NewPPU(CHRROM []byte, debug bool, isPAL bool, cpu CPUInterface) *PPU {
	ppu := &PPU{
		address: &AddressRegister{},
		memory:  make([]byte, 0x4000), // 16KBのPPUメモリ
		register: &register{
			STATUS: 0x00, // 電源投入時はランダム
		},
		sprites:            make(map[int][]byte),
		tiles:              make([]*Tile, 0),
		oam:                make([]byte, 0x100), // OAMは256バイト
		secondaryOAM:       make([]byte, 32),    // セカンダリOAMは32バイト
		initCycles:         0,
		initialized:        false,
		InitWaitCyclesNTSC: 29658, // NTSCの初期化待機時間
		internalBus:        0x00,  // 内部バスの初期値
		busDecay:           0,
		readBuffer:         0x00,
		spriteCount:        0,
		nmi:                false,
		scanline:           0,
		cycle:              0,
		renderingEnabled:   false,
		cpu:                cpu,
		debug:              debug,
	}

	// CHRROMをパターンテーブルにコピー
	// パターンテーブル0に前半4KBをコピー
	copy(ppu.memory[patternTable0Start:patternTable1Start], CHRROM[0:0x1000])
	// パターンテーブル1に後半4KBをコピー
	copy(ppu.memory[patternTable1Start:nameTable0Start], CHRROM[0x1000:0x2000])

	// OAMの初期化（未定義の値で埋める）
	for i := 0; i < 0x100; i++ {
		ppu.oam[i] = 0xFF
	}

	// セカンダリOAMの初期化
	for i := 0; i < 32; i++ {
		ppu.secondaryOAM[i] = 0xFF
	}

	// パレットRAMの初期化（デフォルトパレットで初期化）
	for i := 0x3F00; i <= 0x3F1F; i++ {
		// ユニバーサルバックグラウンドカラー
		if i == 0x3F00 {
			ppu.memory[i] = 0x00 // 黒
			continue
		}
		// スプライトパレット0
		if i >= 0x3F10 && i <= 0x3F1F {
			ppu.memory[i] = byte((i - 0x3F10) % 4)
			continue
		}
		// バックグラウンドパレット0-3
		ppu.memory[i] = byte((i - 0x3F00) % 4)
	}

	// キャンバスの初期化
	ppu.Canvas = &canvas.SDL2Canvas{}
	if err := ppu.Canvas.Setup("gones", windowWidth, windowHeight); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to setup canvas: %v\n", err)
		os.Exit(1)
	}

	// 画面をクリア
	ppu.Canvas.Clear()
	ppu.Canvas.Present()

	// レンダリング設定を有効化
	ppu.register.MASK = 0x1E // 背景とスプライトの表示を有効化

	return ppu
}

func (ar *AddressRegister) set(data uint16) {
	ar.high = byte(data >> 8)
	ar.low = byte(data)
}

func (ar *AddressRegister) get() uint16 {
	return (uint16(ar.high) << 8) | uint16(ar.low)
}

func (ar *AddressRegister) increment() {
	addr := ar.get()
	addr++
	ar.set(addr)
}

func (ar *AddressRegister) increment32() {
	addr := ar.get()
	addr += 32
	ar.set(addr)
}

// RunはPPUの1サイクルを実行
func (p *PPU) Run() {
	// 初期化が完了していない場合は初期化処理を継続
	if !p.initialized {
		p.initCycles++
		if p.initCycles >= p.InitWaitCyclesNTSC {
			p.initialized = true
		}
		return
	}

	// スキャンラインとサイクルの更新
	p.cycle++
	if p.cycle > 339 {
		p.cycle = 0
		p.scanline++
		if p.scanline > 261 {
			p.scanline = -1 // プリレンダリングライン
			p.frame++
			// フレームの終了時に画面を更新
			p.Canvas.Present()
		}
	}

	// レンダリング処理
	if p.scanline >= -1 && p.scanline < 240 {
		// スプライト評価（スキャンライン0-239のサイクル0-63）
		if p.scanline >= 0 && p.cycle < 64 {
			p.evaluateSprites()
		}

		// 背景とスプライトのレンダリング（スキャンライン0-239のサイクル1-256）
		if p.cycle >= 1 && p.cycle <= 256 {
			if p.isBackgroundEnabled() {
				p.renderBackground()
			}
			if p.isSpritesEnabled() {
				p.renderSprites()
			}
		}
	}

	// VBlank処理
	if p.scanline == 241 && p.cycle == 1 {
		p.register.STATUS |= 0x80 // VBlankフラグを設定
		if p.register.CTRL&0x80 != 0 {
			p.cpu.TriggerNMI() // NMI割り込みを発生
		}
	} else if p.scanline == 261 && p.cycle == 1 {
		p.register.STATUS &= 0x7F // VBlankフラグをクリア
		p.spriteZeroHit = false
		p.spriteOverflow = false
	}
}

// renderBackgroundは背景をレンダリング
func (p *PPU) renderBackground() {
	if !p.isBackgroundEnabled() {
		return
	}

	// 現在のスキャンラインとサイクルから描画位置を計算
	x := p.cycle - 1
	if x < 0 || x >= 256 {
		return
	}
	y := p.scanline

	// スクロール位置を考慮したタイルの位置を計算
	tileX := (x + int(p.register.scrollX)) / 8
	tileY := (y + int(p.register.scrollY)) / 8

	// ネームテーブルのベースアドレスを選択
	var baseNameTable uint16 = nameTable0Start
	if p.register.CTRL&0x03 != 0 {
		// ネームテーブルの選択（CTRL レジスタのビット0-1）
		switch p.register.CTRL & 0x03 {
		case 1:
			baseNameTable = nameTable1Start
		case 2:
			baseNameTable = nameTable2Start
		case 3:
			baseNameTable = nameTable3Start
		}
	}

	// ネームテーブルのアドレスを計算
	nameTableAddr := baseNameTable + uint16(tileY%30*32+tileX%32)

	// タイルのインデックスを取得
	tileIndex := p.readMemory(nameTableAddr)

	// パターンテーブルのベースアドレスを選択（CTRL レジスタのビット4）
	var patternTableBase uint16 = patternTable0Start
	if p.register.CTRL&0x10 != 0 {
		patternTableBase = patternTable1Start
	}

	// タイル内のピクセル位置を計算
	pixelX := (x + int(p.register.scrollX)) % 8
	pixelY := (y + int(p.register.scrollY)) % 8

	// パターンテーブルのアドレスを計算
	patternAddr := patternTableBase + uint16(tileIndex)*16 + uint16(pixelY)

	// タイルデータを取得
	lowByte := p.readMemory(patternAddr)
	highByte := p.readMemory(patternAddr + 8)

	// ピクセルの色インデックスを計算
	colorIndex := ((highByte >> (7 - uint(pixelX))) & 0x01) << 1
	colorIndex |= (lowByte >> (7 - uint(pixelX))) & 0x01

	// アトリビュートテーブルのアドレスを計算
	attrTableAddr := baseNameTable + 0x3C0 + uint16((tileY/4)*8+(tileX/4))
	attrByte := p.readMemory(attrTableAddr)

	// 4x4タイル内での位置に基づいてパレット番号を取得
	attrShift := ((tileY%4)/2)*4 + ((tileX%4)/2)*2
	paletteIndex := (attrByte >> attrShift) & 0x03

	// パレットのアドレスを計算
	paletteAddr := uint16(0x3F00) + uint16(paletteIndex<<2) + uint16(colorIndex)
	colorValue := p.readMemory(paletteAddr) & 0x3F
	if paletteAddr == 0x3F00 {
		fmt.Printf("DEBUG: $3F00 = %02X (used in palette rendering)\n", colorValue)
	}
	// 色を取得して描画
	if colorValue < uint8(len(palette)) {
		p.Canvas.SetPixel(x, y, &palette[colorValue])
	}
}

// renderSpritesはスプライトをレンダリング
func (p *PPU) renderSprites() {
	if !p.isSpritesEnabled() {
		return
	}

	// スプライトのサイズを取得（8x8 or 8x16）
	spriteHeight := 8
	if p.register.CTRL&0x20 != 0 {
		spriteHeight = 16
	}

	// セカンダリOAMからスプライトを取得（後ろから処理して優先順位を正しく扱う）
	for i := 7; i >= 0; i-- {
		y := p.secondaryOAM[i*4]
		x := p.secondaryOAM[i*4+3]
		tileIndex := p.secondaryOAM[i*4+1]
		attributes := p.secondaryOAM[i*4+2]

		// スプライトが現在のスキャンラインに表示されるかチェック
		if y <= byte(p.scanline) && byte(p.scanline) < y+byte(spriteHeight) {
			// スプライト0ヒットのチェック
			if i == 0 && !p.spriteZeroHit && p.isBackgroundEnabled() {
				if x <= byte(p.cycle-1) && byte(p.cycle-1) < x+8 {
					p.spriteZeroHit = true
					p.register.STATUS |= 0x40
				}
			}

			// スプライトの描画
			p.renderSprite(x, y, tileIndex, attributes, spriteHeight)
		}
	}
}

// renderSpriteはスプライトを描画
func (p *PPU) renderSprite(x, y byte, tileIndex byte, attributes byte, spriteHeight int) {
	// スプライトのパターンテーブルアドレスを計算
	var patternAddr uint16
	if spriteHeight == 16 {
		// 8x16モードでは、タイルインデックスの最下位ビットで
		// パターンテーブルを選択
		patternAddr = (uint16(tileIndex&0xFE) * 16) |
			(uint16(tileIndex&0x01) * 0x1000)
	} else {
		// 8x8モードでは、CTRLレジスタのビット3で
		// スプライトパターンテーブルを選択
		patternAddr = uint16(tileIndex) * 16
		if p.register.CTRL&0x08 != 0 {
			patternAddr += 0x1000
		}
	}

	// スプライトの背面/前面の優先順位
	behindBackground := (attributes & 0x20) != 0

	// スプライトの描画
	for row := 0; row < spriteHeight; row++ {
		// 垂直反転の処理
		patternRow := row
		if attributes&0x80 != 0 {
			patternRow = spriteHeight - 1 - row
		}

		lowByte := p.readMemory(patternAddr + uint16(patternRow))
		highByte := p.readMemory(patternAddr + uint16(patternRow) + 8)

		for col := 0; col < 8; col++ {
			// 水平反転の処理
			patternCol := col
			if attributes&0x40 != 0 {
				patternCol = 7 - col
			}

			// ピクセルの色を計算
			pixel := ((highByte >> (7 - patternCol)) & 1) << 1
			pixel |= (lowByte >> (7 - patternCol)) & 1

			// 透明色（0）の場合はスキップ
			if pixel == 0 {
				continue
			}

			// 画面座標の計算
			screenX := int(x) + col
			screenY := int(y) + row

			// 画面外のピクセルはスキップ
			if screenX < 0 || screenX >= 256 || screenY < 0 || screenY >= 240 {
				continue
			}

			// パレットから色を取得
			paletteIndex := ((attributes & 0x03) << 2) | pixel
			color := p.getPaletteColor(1, paletteIndex).(*color.RGBA)

			// 背景との優先順位を考慮して描画
			if !behindBackground {
				p.Canvas.SetPixel(screenX, screenY, color)
			}
		}
	}
}

// getPaletteColorはパレットから色を取得
func (p *PPU) getPaletteColor(table byte, index byte) color.Color {
	// パレットアドレスを計算
	addr := 0x3F00
	if table == 0 {
		addr += int(index & 0x0F) // 下位4ビットのみ使用
	} else {
		addr += 0x10 + int(index&0x0F) // 下位4ビットのみ使用
	}

	// パレットから色を取得
	paletteIndex := p.readMemory(uint16(addr)) & 0x3F // 上位2ビットをマスク

	// グレースケールモードの処理
	if p.register.MASK&0x01 != 0 {
		paletteIndex &= 0x30 // 上位2ビットのみ保持
	}

	// パレットインデックスの範囲チェック
	if paletteIndex >= uint8(len(palette)) {
		paletteIndex = 0
	}

	// 基本色を取得
	baseColor := palette[paletteIndex]

	// カラー強調の処理
	r := uint8(baseColor.R)
	g := uint8(baseColor.G)
	b := uint8(baseColor.B)

	// 赤強調
	if p.register.MASK&0x20 != 0 {
		r = uint8(float64(r) * 1.1)
	}
	// 緑強調
	if p.register.MASK&0x40 != 0 {
		g = uint8(float64(g) * 1.1)
	}
	// 青強調
	if p.register.MASK&0x80 != 0 {
		b = uint8(float64(b) * 1.1)
	}

	return &color.RGBA{r, g, b, 0xFF}
}

// ReadRegisterはPPUのレジスタを読み込む
func (p *PPU) ReadRegister(addr uint16) byte {
	switch addr {
	case 0x2002: // PPUSTATUS
		status := p.register.STATUS
		// VBlankフラグをクリア
		p.register.STATUS &^= 0x80
		// アドレスラッチをリセット
		p.register.writeToggle = false
		return status
	case 0x2007: // PPUDATA
		currentAddr := p.address.get()

		// 現在のバッファの値を保存
		data := p.readBuffer

		// アドレスを増分
		if p.register.CTRL&0x04 != 0 {
			p.address.increment32()
		} else {
			p.address.increment()
		}

		// パレットデータの場合は直接読み込む
		if currentAddr >= 0x3F00 && currentAddr <= 0x3FFF {
			p.readBuffer = p.readMemory(currentAddr)
			if p.debug {
				fmt.Printf("Reading palette data from PPU memory at %04x: %02x, buffer updated to: %02x\n", currentAddr, data, p.readBuffer)
			}
			return data
		}

		// 通常のメモリ読み込みの場合
		// 次のデータをバッファに読み込む
		p.readBuffer = p.readMemory(p.address.get())

		if p.debug {
			fmt.Printf("Reading from PPU memory at %04x: %02x, next buffer: %02x\n", currentAddr, data, p.readBuffer)
		}
		return data
	}
	return 0
}

// WriteRegisterはPPUのレジスタに書き込む
func (p *PPU) WriteRegister(addr uint16, data byte) {
	if p.debug {
		fmt.Printf("PPU Write Register: addr=%04x, data=%02x\n", addr, data)
	}

	switch addr {
	case 0x2000: // PPUCTRL
		oldCtrl := p.register.CTRL
		p.register.CTRL = data | 0x80 // NMIを常に有効にする
		// NMIが無効から有効に変更された場合、VBlank中であればNMIを生成
		if oldCtrl&0x80 == 0 && p.register.CTRL&0x80 != 0 && p.register.STATUS&0x80 != 0 {
			p.generateNMI()
		}
	case 0x2001: // PPUMASK
		p.register.MASK = data
	case 0x2002: // PPUSTATUS
		// 読み取り専用
	case 0x2003: // OAMADDR
		p.register.OAMADDR = data
	case 0x2004: // OAMDATA
		p.register.OAMDATA = data
		p.oam[p.register.OAMADDR] = data
		p.register.OAMADDR++
	case 0x2005: // PPUSCROLL
		p.register.SCROLL = data
		if p.register.writeToggle {
			p.register.scrollY = data
		} else {
			p.register.scrollX = data
		}
		p.register.writeToggle = !p.register.writeToggle
	case 0x2006: // PPUADDR
		if p.register.writeToggle {
			p.address.low = data
			p.register.writeToggle = false
		} else {
			p.address.high = data & 0x3F
			p.register.writeToggle = true
		}
	case 0x2007: // PPUDATA
		addr := p.address.get()
		if addr >= 0x3F00 && addr <= 0x3F1F {
			// パレットデータの書き込み
			p.memory[addr] = data
			// ミラーリング（0x3F10, 0x3F14, 0x3F18, 0x3F1C）
			if addr == 0x3F00 {
				p.memory[0x3F10] = data
			} else if addr == 0x3F04 {
				p.memory[0x3F14] = data
			} else if addr == 0x3F08 {
				p.memory[0x3F18] = data
			} else if addr == 0x3F0C {
				p.memory[0x3F1C] = data
			} else if addr == 0x3F10 {
				p.memory[0x3F00] = data
			} else if addr == 0x3F14 {
				p.memory[0x3F04] = data
			} else if addr == 0x3F18 {
				p.memory[0x3F08] = data
			} else if addr == 0x3F1C {
				p.memory[0x3F0C] = data
			}
		} else {
			p.memory[addr] = data
		}
		p.address.increment()
	}
}

// readMemoryはPPUのメモリを読み込む
func (p *PPU) readMemory(addr uint16) byte {
	// アドレスを0x3FFFまでに制限
	addr &= 0x3FFF

	// ネームテーブルのミラーリング（0x2000-0x2EFF）
	if addr >= 0x2000 && addr < 0x3000 {
		// ベースアドレスを計算（0x2000-0x23FF）
		baseAddr := addr & 0x23FF
		return p.memory[baseAddr]
	}

	// 0x3000-0x3EFFは0x2000-0x2EFFのミラー
	if addr >= 0x3000 && addr < 0x3F00 {
		return p.readMemory(addr - 0x1000)
	}

	// パレットのミラーリング（0x3F00-0x3FFF）
	if addr >= 0x3F00 {
		addr &= 0x3F1F
		if addr&0x13 == 0x10 {
			addr &= 0x0F
		}
		return p.memory[addr]
	}

	return p.memory[addr]
}

// writeMemoryはPPUのメモリに書き込む
func (p *PPU) writeMemory(addr uint16, data byte) {
	// アドレスを0x3FFFまでに制限
	addr &= 0x3FFF

	// ネームテーブルのミラーリング（0x2000-0x2EFF）
	if addr >= 0x2000 && addr < 0x3000 {
		// ベースアドレスを計算（0x2000-0x23FF）
		baseAddr := addr & 0x23FF
		p.memory[baseAddr] = data
		return
	}

	// 0x3000-0x3EFFは0x2000-0x2EFFのミラー
	if addr >= 0x3000 && addr < 0x3F00 {
		p.writeMemory(addr-0x1000, data)
		return
	}

	// パレットのミラーリング（0x3F00-0x3FFF）
	if addr >= 0x3F00 {
		addr &= 0x3F1F
		if addr&0x13 == 0x10 {
			addr &= 0x0F
		}
		p.memory[addr] = data
		return
	}

	p.memory[addr] = data
}

// OAMDMAはスプライトデータのDMA転送を実行
func (p *PPU) OAMDMA(data []byte) {
	// OAMのアドレスから開始
	addr := p.register.OAMADDR
	// 256バイトのデータを転送
	for i := 0; i < 256; i++ {
		p.oam[addr] = data[i]
		addr++
	}
}

// スクロール位置の更新
func (p *PPU) updateScroll() {
	// ネームテーブルの選択（CTRL レジスタのビット0-1）
	baseNameTable := uint16(p.register.CTRL&0x03)*0x400 + 0x2000

	// スクロール位置に基づいてネームテーブルのアドレスを更新
	x := uint16(p.register.scrollX)
	y := uint16(p.register.scrollY)

	// X方向のスクロールがネームテーブルの境界を超える場合
	if x >= 256 {
		x -= 256
		baseNameTable ^= 0x0400 // 水平方向のネームテーブルを切り替え
	}

	// Y方向のスクロールがネームテーブルの境界を超える場合
	if y >= 240 {
		y -= 240
		baseNameTable ^= 0x0800 // 垂直方向のネームテーブルを切り替え
	}

	// スクロール位置を更新
	p.address.set(baseNameTable + (y/8)*32 + x/8)
}

// NMI割り込みを生成
func (p *PPU) generateNMI() {
	p.nmi = true
	p.register.STATUS |= 0x80 // VBlankフラグを設定
	if p.cpu != nil {
		p.cpu.TriggerNMI()
	}
}

// スプライト評価を実行
func (p *PPU) evaluateSprites() {
	// セカンダリOAMをクリア
	for i := 0; i < 32; i++ {
		p.secondaryOAM[i] = 0xFF
	}

	secondaryOAMIndex := 0
	spriteCount := 0
	spriteHeight := 8
	if p.register.CTRL&0x20 != 0 {
		spriteHeight = 16
	}

	// プライマリOAMをスキャン
	for i := 0; i < 64; i++ {
		y := p.oam[i*4]
		// スプライトが現在のスキャンラインに表示されるかチェック
		if y <= byte(p.scanline) && byte(p.scanline) < y+byte(spriteHeight) {
			if spriteCount < 8 {
				// セカンダリOAMにコピー
				p.secondaryOAM[secondaryOAMIndex*4] = y
				p.secondaryOAM[secondaryOAMIndex*4+1] = p.oam[i*4+1]
				p.secondaryOAM[secondaryOAMIndex*4+2] = p.oam[i*4+2]
				p.secondaryOAM[secondaryOAMIndex*4+3] = p.oam[i*4+3]
				secondaryOAMIndex++
			}
			spriteCount++
			if spriteCount > 8 {
				// スプライトオーバーフロー
				p.spriteOverflow = true
				p.register.STATUS |= 0x20
				break
			}
		}
	}
}

// DebugPrintRegistersはPPUレジスタの状態を表示する
func (p *PPU) DebugPrintRegisters() {
	fmt.Printf("PPU Registers:\n")
	fmt.Printf("  PPUCTRL  (0x2000): %08b\n", p.register.CTRL)
	fmt.Printf("  PPUMASK  (0x2001): %08b\n", p.register.MASK)
	fmt.Printf("  PPUSTATUS(0x2002): %08b\n", p.register.STATUS)
	fmt.Printf("  OAMADDR  (0x2003): %08b\n", p.register.OAMADDR)
	fmt.Printf("  OAMDATA  (0x2004): %08b\n", p.register.OAMDATA)
	fmt.Printf("  PPUSCROLL(0x2005): %08b\n", p.register.SCROLL)
	fmt.Printf("  PPUADDR  (0x2006): %08b\n", p.register.ADDRESS)
	fmt.Printf("  PPUDATA  (0x2007): %08b\n", p.register.DATA)
	fmt.Printf("  Scroll X: %d, Y: %d\n", p.register.scrollX, p.register.scrollY)
	fmt.Printf("  Write Toggle: %v\n", p.register.writeToggle)
}

// isSpritesEnabledはスプライトの描画が有効かどうかを返す
func (p *PPU) isSpritesEnabled() bool {
	return p.register.MASK&0x10 != 0
}

// isBackgroundEnabledは背景の描画が有効かどうかを返す
func (p *PPU) isBackgroundEnabled() bool {
	return p.register.MASK&0x08 != 0
}

func (p *PPU) read(addr uint16) byte {
	return p.readMemory(addr)
}

func (p *PPU) getColorFromPalette(paletteNum byte, colorIndex byte) *color.RGBA {
	// パレットインデックスを0x3F00-0x3F1Fの範囲に制限
	addr := uint16(0x3F00) + uint16(paletteNum*4) + uint16(colorIndex&0x03)
	addr &= 0x3F1F

	// パレットデータを読み込む
	paletteIndex := p.readMemory(addr) & 0x3F

	// グレースケールモードの処理
	if p.register.MASK&0x01 != 0 {
		paletteIndex &= 0x30
	}

	return &palette[paletteIndex]
}

func (p *PPU) getAttributeByte(tileX, tileY uint16) byte {
	// アトリビュートテーブルのベースアドレスを計算
	attrAddr := uint16(0x23C0) | (tileY&0x1C)<<1 | (tileX >> 2)

	// アトリビュートバイトを読み込む
	attrByte := p.readMemory(attrAddr)

	// タイルの位置に応じてパレット番号を抽出
	shift := ((tileY & 0x02) << 1) | (tileX & 0x02)
	return (attrByte >> shift) & 0x03
}

// renderPixelはスクリーン上の1ピクセルを描画
func (p *PPU) renderPixel() {
	if !p.renderingEnabled || p.line < 0 || p.line >= 240 || p.cycle < 1 || p.cycle > 256 {
		return
	}

	x := p.cycle - 1
	y := p.scanline // lineではなくscanlineを使用

	// タイルの位置を計算
	tileX := x / 8
	tileY := y / 8

	// タイル内のピクセル位置を計算
	pixelX := x % 8
	pixelY := y % 8

	// ネームテーブルからタイルインデックスを取得
	nameTableAddr := nameTable0Start + uint16(tileY*32+tileX)
	tileIndex := p.memory[nameTableAddr]

	// パターンテーブルのベースアドレスを選択
	var patternTableBase uint16 = patternTable0Start
	if p.register.CTRL&0x10 != 0 {
		patternTableBase = patternTable1Start
	}

	// パターンテーブルのアドレスを計算
	patternAddr := patternTableBase + uint16(tileIndex)*16 + uint16(pixelY) + 9

	// タイルデータを取得
	lowByte := p.memory[patternAddr]
	highByte := p.memory[patternAddr+8]

	// ピクセルの色インデックスを計算
	colorIndex := ((highByte >> (7 - pixelX)) & 0x01) << 1
	colorIndex |= (lowByte >> (7 - pixelX)) & 0x01

	// 属性テーブルのアドレスを計算
	attrTableAddr := nameTableAddr + 0x3C0 + uint16((tileY/4)*8+(tileX/4))

	// パレットインデックスを取得
	attrByte := p.memory[attrTableAddr]
	attrShift := ((tileY%4)/2)*4 + ((tileX%4)/2)*2
	paletteIndex := (attrByte >> attrShift) & 0x03

	// パレットのアドレスを計算
	paletteAddr := uint16(0x3F00) + uint16(paletteIndex<<2) + uint16(colorIndex)
	colorValue := p.readMemory(paletteAddr) & 0x3F

	// 色を取得
	if colorValue < uint8(len(palette)) {
		p.Canvas.SetPixel(x, y, &palette[colorValue])
	}
}

func (p *PPU) setPixel(x, y int, color *color.RGBA) {
	if p.debug {
		// fmt.Printf("SetPixel: x=%d, y=%d, color=%v\n", x, y, color)
	}
	p.Canvas.SetPixel(x, y, color)
}
