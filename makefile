build:
	go build -o bin/gones
run :
	go run main.go cpu.go ppu.go opecodes.go renderer.go
test:
	go test -v -run TestCPU_status
