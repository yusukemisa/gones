package bus

import (
	"fmt"
	"testing"
)

func TestBus_Read(t *testing.T) {
	t.Parallel()

	bus := NewBus(true)
	for i := 0; i < 0x0800; i++ {
		bus.cpuRAM[i] = byte(i & 0b11111111)
	}

	for _, tt := range []struct {
		address uint16
		want    byte
	}{
		{0x0001, 0x01},
		{0x0110, 0x10},
		{0x07FF, 0xFF},
	} {
		tt := tt
		t.Run(fmt.Sprintf("Read:%#04x", tt.address), func(t *testing.T) {
			if want, got := tt.want, bus.Read(tt.address); want != got {
				t.Errorf("want=%v, got=%v", want, got)
			}
		})
	}
}

func TestBus_Write(t *testing.T) {
	t.Parallel()

	bus := NewBus(true)
	for _, tt := range []struct {
		address uint16
		data    byte
		want    byte
	}{
		{0x0001, 0x01, 0x01},
		{0x0110, 0x10, 0x10},
		{0x07FF, 0xFF, 0xFF},
	} {
		tt := tt
		t.Run(fmt.Sprintf("Write:address=%#04x,data=%#02x", tt.address, tt.data), func(t *testing.T) {
			if want, got := byte(0), bus.Read(tt.address); want != got {
				t.Errorf("want=%v, got=%v", want, got)
			}
			bus.write(tt.address, tt.data)
			if want, got := tt.want, bus.Read(tt.address); want != got {
				t.Errorf("want=%v, got=%v", want, got)
			}
		})
	}
}
