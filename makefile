build:
	go build -o bin/gones
run :
	go run main.go cpu.go ppu.go opecodes.go canvas.go palette.go
test:
	go test -v -run TestCPU_status
