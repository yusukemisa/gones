package util

func TestBit(x, n byte) bool {
	return x&(1<<n) != 0
}

func SetBit(x, n byte) byte {
	return x | (1 << n)
}

func ClearBit(x, n byte) byte {
	return x &^ (1 << n)
}
