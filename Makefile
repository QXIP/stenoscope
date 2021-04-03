default: build

compile-go:
	go build -buildmode=c-shared -o sstablekeys.so src/SSTableKeysNode.go

build: compile-go
	npm run build
