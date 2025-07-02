package ppu

import (
	"fmt"
	"image/color"
	"testing"
)

type MockCanvas struct {
	pixels map[int]map[int]*color.RGBA
}

func NewMockCanvas() *MockCanvas {
	return &MockCanvas{
		pixels: make(map[int]map[int]*color.RGBA),
	}
}

func (m *MockCanvas) Setup(title string, width, height int) error {
	return nil
}

func (m *MockCanvas) SetPixel(x, y int, c *color.RGBA) {
	if m.pixels[y] == nil {
		m.pixels[y] = make(map[int]*color.RGBA)
	}
	m.pixels[y][x] = c
}

func (m *MockCanvas) Update() {
	// モックでは何もしない
}

func (m *MockCanvas) Render() {
	// モックでは何もしない
}

func (m *MockCanvas) Clear() {
	m.pixels = make(map[int]map[int]*color.RGBA)
}

func (m *MockCanvas) Present() {
	// モックでは何もしない
}

func (m *MockCanvas) Shutdown() {
	// モックでは何もしない
}

func (m *MockCanvas) GetPixel(x, y int) *color.RGBA {
	if row, ok := m.pixels[y]; ok {
		if pixel, ok := row[x]; ok {
			return pixel
		}
	}
	return nil
}

func setupTestPPU() *PPU {
	p := &PPU{
		memory:     make([]byte, 0x4000),
		register:   &register{},
		address:    &AddressRegister{},
		debug:      true,
		Canvas:     NewMockCanvas(),
		readBuffer: 0,
		oam:        make([]byte, 0x100),
	}

	// パレットの初期化
	for i := 0x3F00; i <= 0x3F1F; i++ {
		if i == 0x3F00 {
			p.memory[i] = 0x0F // ユニバーサルバックグラウンドカラー
		} else {
			p.memory[i] = byte((i - 0x3F00) % 4)
		}
	}

	// 最初の読み込みのためにバッファを初期化
	p.WriteRegister(0x2006, 0x20)
	p.WriteRegister(0x2006, 0x00)
	p.ReadRegister(0x2007) // ダミーリード

	return p
}

func TestNameTableWrite(t *testing.T) {
	t.Run("Single Byte Write and Read", func(t *testing.T) {
		p := setupTestPPU()

		// PPUADDRに0x2000を設定
		p.WriteRegister(0x2006, 0x20)
		p.WriteRegister(0x2006, 0x00)

		// 1バイトのデータを書き込む
		p.WriteRegister(0x2007, 0x30)

		// メモリの内容を直接確認
		memVal := p.memory[0x2000]
		t.Logf("Memory at 0x2000 after write: %02x", memVal)

		// PPUADDRを0x2000に戻す
		p.WriteRegister(0x2006, 0x20)
		p.WriteRegister(0x2006, 0x00)

		// バッファの初期状態を確認
		t.Logf("Read buffer before read: %02x", p.readBuffer)

		// データを読み込んで検証
		actual := p.ReadRegister(0x2007)
		t.Logf("First read value: %02x", actual)
		t.Logf("Read buffer after read: %02x", p.readBuffer)

		if actual != 0x30 {
			t.Errorf("Expected 0x30, got %02x", actual)
		}
	})

	t.Run("Multiple Bytes Write and Read", func(t *testing.T) {
		p := setupTestPPU()

		// PPUADDRに0x2000を設定
		p.WriteRegister(0x2006, 0x20)
		p.WriteRegister(0x2006, 0x00)

		// データを書き込む
		testData := []byte{0x30, 0x31, 0x32, 0x33}
		for i, data := range testData {
			p.WriteRegister(0x2007, data)
			// メモリの内容を直接確認
			memVal := p.memory[0x2000+i]
			t.Logf("Memory at 0x%04x after write: %02x", 0x2000+i, memVal)
		}

		// PPUADDRを0x2000に戻す
		p.WriteRegister(0x2006, 0x20)
		p.WriteRegister(0x2006, 0x00)

		// バッファの初期状態を確認
		t.Logf("Read buffer before reads: %02x", p.readBuffer)

		// データを読み込んで検証
		for i, expected := range testData {
			actual := p.ReadRegister(0x2007)
			t.Logf("Read %d: expected=%02x, actual=%02x, buffer=%02x", i, expected, actual, p.readBuffer)
			if actual != expected {
				t.Errorf("At index %d: Expected %02x, got %02x", i, expected, actual)
			}
		}
	})
}

func TestPPU_RenderBackground(t *testing.T) {
	p := setupTestPPU()

	// パターンテーブルにテストデータを設定
	patternData := []byte{
		0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, // タイルデータ（低位ビット）
		0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, 0x00, 0xFF, // タイルデータ（高位ビット）
	}
	copy(p.memory[0:16], patternData)

	// ネームテーブルにタイルインデックスを設定
	p.writeMemory(0x2000, 0x00) // 最初のタイルを使用

	// 属性テーブルにパレット情報を設定
	p.writeMemory(0x23C0, 0x00) // パレット0を使用

	// パレットデータを設定
	p.writeMemory(0x3F00, 0x0F) // ユニバーサルバックグラウンドカラー
	p.writeMemory(0x3F01, 0x01) // パレット0の色1
	p.writeMemory(0x3F02, 0x02) // パレット0の色2
	p.writeMemory(0x3F03, 0x03) // パレット0の色3

	// レンダリングを有効化
	p.register.MASK = 0x08    // 背景表示を有効化
	p.renderingEnabled = true // 明示的にレンダリングを有効化

	// PPUの状態を設定
	p.scanline = 0
	p.cycle = 1

	if p.debug {
		fmt.Printf("Before rendering - MASK: %02x, renderingEnabled: %v\n", p.register.MASK, p.renderingEnabled)
		fmt.Printf("Pattern data at 0x0000: %02x %02x %02x %02x %02x %02x %02x %02x\n",
			p.memory[0], p.memory[1], p.memory[2], p.memory[3],
			p.memory[4], p.memory[5], p.memory[6], p.memory[7])
	}

	// 直接renderBackgroundを呼び出す
	p.renderBackground()

	// レンダリング結果を検証
	mockCanvas := p.Canvas.(*MockCanvas)
	expectedPattern := []struct {
		x, y     int
		expected *color.RGBA
	}{
		{0, 0, &palette[0x01]},
		{1, 0, &palette[0x02]},
		{2, 0, &palette[0x01]},
		{3, 0, &palette[0x02]},
	}

	for _, test := range expectedPattern {
		actual := mockCanvas.GetPixel(test.x, test.y)
		if actual == nil {
			t.Errorf("No pixel at (%d,%d)", test.x, test.y)
			continue
		}
		if *actual != *test.expected {
			t.Errorf("At (%d,%d): expected %v, got %v", test.x, test.y, test.expected, actual)
		}
	}
}

func TestPPU_ReadMemory(t *testing.T) {
	p := setupTestPPU()

	// テストデータを直接メモリに書き込む
	testAddr := uint16(0x2000)
	testData := byte(0x42)
	p.memory[testAddr] = testData

	// メモリから読み込んで検証
	result := p.readMemory(testAddr)
	if result != testData {
		t.Errorf("Expected %02x at address %04x, got %02x", testData, testAddr, result)
	}
}

func TestPPU_WriteMemory(t *testing.T) {
	p := setupTestPPU()

	// テストデータをメモリに書き込む
	testAddr := uint16(0x2000)
	testData := byte(0x42)
	p.writeMemory(testAddr, testData)

	// メモリから直接読み込んで検証
	result := p.memory[testAddr]
	if result != testData {
		t.Errorf("Expected %02x at address %04x, got %02x", testData, testAddr, result)
	}
}

func TestPPU_AddressRegister(t *testing.T) {
	t.Run("アドレスレジスタの設定と取得", func(t *testing.T) {
		p := setupTestPPU()

		// アドレスを設定
		p.WriteRegister(0x2006, 0x20) // 上位バイト
		p.WriteRegister(0x2006, 0x00) // 下位バイト

		// アドレスの取得を確認
		addr := p.address.get()
		if addr != 0x2000 {
			t.Errorf("アドレスが正しく設定されていません: 期待値=0x2000, 実際の値=0x%04x", addr)
		}
	})

	t.Run("アドレスレジスタのインクリメント", func(t *testing.T) {
		p := setupTestPPU()

		// アドレスを設定
		p.WriteRegister(0x2006, 0x20)
		p.WriteRegister(0x2006, 0x00)

		// アドレスをインクリメント
		p.address.increment()

		// インクリメント後のアドレスを確認
		addr := p.address.get()
		if addr != 0x2001 {
			t.Errorf("アドレスが正しくインクリメントされていません: 期待値=0x2001, 実際の値=0x%04x", addr)
		}
	})
}

func TestPPU_ReadBuffer(t *testing.T) {
	t.Run("バッファの初期化", func(t *testing.T) {
		p := setupTestPPU()

		// メモリにデータを書き込む
		p.memory[0x2000] = 0x30

		// アドレスを設定
		p.WriteRegister(0x2006, 0x20)
		p.WriteRegister(0x2006, 0x00)

		// バッファの初期状態を確認
		if p.readBuffer != 0x30 {
			t.Errorf("バッファが正しく初期化されていません: 期待値=0x30, 実際の値=0x%02x", p.readBuffer)
		}
	})

	t.Run("バッファの更新", func(t *testing.T) {
		p := setupTestPPU()

		// メモリにデータを書き込む
		p.memory[0x2000] = 0x30
		p.memory[0x2001] = 0x31

		// アドレスを設定
		p.WriteRegister(0x2006, 0x20)
		p.WriteRegister(0x2006, 0x00)

		// 最初の読み込み
		firstRead := p.ReadRegister(0x2007)
		t.Logf("最初の読み込み: %02x, バッファ: %02x", firstRead, p.readBuffer)

		// 2回目の読み込み
		secondRead := p.ReadRegister(0x2007)
		t.Logf("2回目の読み込み: %02x, バッファ: %02x", secondRead, p.readBuffer)

		if firstRead != 0x30 || secondRead != 0x31 {
			t.Errorf("読み込み値が正しくありません: 1回目=0x%02x, 2回目=0x%02x", firstRead, secondRead)
		}
	})
}

func TestPPU_MemoryMirroring(t *testing.T) {
	t.Run("ネームテーブルのミラーリング", func(t *testing.T) {
		p := setupTestPPU()

		// メモリにデータを書き込む
		p.memory[0x2000] = 0x30

		// ミラーリングされたアドレスから読み込む
		value := p.readMemory(0x3000)
		if value != 0x30 {
			t.Errorf("ミラーリングが正しく機能していません: 期待値=0x30, 実際の値=0x%02x", value)
		}
	})

	t.Run("パレットのミラーリング", func(t *testing.T) {
		p := setupTestPPU()

		// パレットにデータを書き込む
		p.memory[0x3F00] = 0x30

		// ミラーリングされたアドレスから読み込む
		value := p.readMemory(0x3F20)
		if value != 0x30 {
			t.Errorf("パレットのミラーリングが正しく機能していません: 期待値=0x30, 実際の値=0x%02x", value)
		}
	})
}
