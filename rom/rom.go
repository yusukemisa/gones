package rom

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os"
)

type Rom struct {
	PRGROM  []byte
	CHRROM  []byte
	PRGSize int // Size of PRG ROM in bytes
}

// NewRom creates `*Rom` from `nesFile`.
//
// 00000000  4e 45 53 1a 02 01 01 00  00 00 00 00 00 00 00 00  |NES.............|
// 4: Size of PRG ROM in 16 KB units
// 5: Size of CHR ROM in 8 KB units (Value 0 means the board uses CHR RAM)
func NewRom(nesFile *os.File, debug bool) *Rom {
	// ファイルの内容を読み込む
	data, err := io.ReadAll(nesFile)
	if err != nil {
		log.Fatal("failed to read NES file:", err)
	}

	// iNESヘッダーの検証
	if !bytes.Equal(data[0:4], []byte{0x4E, 0x45, 0x53, 0x1A}) {
		log.Fatal("invalid NES header")
	}

	// PRGROMとCHRROMのサイズを取得
	sizeOfPRG := int(data[4]) // PRGROMのサイズ（16KB単位） 2
	sizeOfCHR := int(data[5]) // CHRROMのサイズ（8KB単位）1

	fmt.Printf("NewRom: PRG size=%d (0x%x), CHR size=%d (0x%x)\n",
		sizeOfPRG*0x4000, sizeOfPRG*0x4000,
		sizeOfCHR*0x2000, sizeOfCHR*0x2000)

	// PRGROMの読み込み位置
	prgOffset := 0x10 // iNESヘッダーの直後
	PRGROM := make([]byte, sizeOfPRG*0x4000)
	copy(PRGROM, data[prgOffset:prgOffset+sizeOfPRG*0x4000])

	// CHRROMの読み込み位置を修正
	chrOffset := 0x10 + sizeOfPRG*0x4000 // PRG-ROMの直後
	CHRROM := make([]byte, sizeOfCHR*0x2000)
	copy(CHRROM, data[chrOffset:chrOffset+sizeOfCHR*0x2000])

	return &Rom{
		PRGROM:  PRGROM,
		CHRROM:  CHRROM,
		PRGSize: sizeOfPRG * 0x4000,
	}
}

func (r *Rom) ReadPRG(address uint16) byte {
	return r.PRGROM[address]
}
