package rom

import (
	"io"
	"log"
	"os"
)

type Rom struct {
	PRG []byte
	CHR []byte
}

// NewRom creates `*Rom` from `nesFile`.
//
// 00000000  4e 45 53 1a 02 01 01 00  00 00 00 00 00 00 00 00  |NES.............|
// 0-3: Constant $4E $45 $53 $1A ("NES" followed by MS-DOS end-of-file)
// 4: Size of PRG ROM in 16 KB units
// 5: Size of CHR ROM in 8 KB units (Value 0 means the board uses CHR RAM)
func NewRom(nesFile *os.File) *Rom {
	sr := io.NewSectionReader(nesFile, 0, 0x10)
	buf := make([]byte, 0x10) // 16ByteのiNESヘッダ
	if _, err := sr.Read(buf); err != nil {
		log.Fatal("failed to read iNES header:", err)
	}

	sizeOfPRG, sizeOfCHR := int(buf[4]), int(buf[5])
	pr := io.NewSectionReader(nesFile, 0x10, int64(sizeOfPRG*0x4000))
	cr := io.NewSectionReader(nesFile, int64(0x10+sizeOfPRG*0x4000), int64(sizeOfCHR*0x2000))

	PRGROM, CHRROM := make([]byte, sizeOfPRG*0x4000), make([]byte, sizeOfCHR*0x2000)
	if _, err := pr.Read(PRGROM); err != nil {
		log.Fatal("failed to read PRGROM:", err)
	}

	if _, err := cr.Read(CHRROM); err != nil {
		log.Fatal(err)
	}
	return &Rom{
		PRG: PRGROM,
		CHR: CHRROM,
	}
}

func (r *Rom) ReadPRG(address uint16) byte {
	return r.PRG[address]
}
