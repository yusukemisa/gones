package main

import (
	"bufio"
	"log"
	"os"
)

func main() {
	f, err := os.Open("sample1.nes")
	if err != nil {
		log.Fatal(err)
	}

	r := bufio.NewReader(f)
	buf := make([]byte, 0x10) // 16ByteのiNESヘッダ
	_, err = r.Read(buf)
	if err != nil {
		log.Fatal(err)
	}

	// 00000000  4e 45 53 1a 02 01 01 00  00 00 00 00 00 00 00 00  |NES.............|
	// 0-3: Constant $4E $45 $53 $1A ("NES" followed by MS-DOS end-of-file)
	//   4: Size of PRG ROM in 16 KB units
	//   5: Size of CHR ROM in 8 KB units (Value 0 means the board uses CHR RAM)
	sizeOfPRG, sizeOfCHR := int(buf[4]), int(buf[5])
	PRGROM, CHRROM := make([]byte, sizeOfPRG*0x4000), make([]byte, sizeOfCHR*0x2000)

	_, err = r.Read(PRGROM)
	if err != nil {
		log.Fatal(err)
	}
	_, err = r.Read(CHRROM)
	if err != nil {
		log.Fatal(err)
	}
}
