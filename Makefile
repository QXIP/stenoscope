default: build

go-get:
	go get ./...

compile-cmd:
	go build -o SSTableKeys src/SSTableKeys.go
	go build -buildmode=c-shared -o sstablekeys.so src/SSTableKeysNode.go

compile-go:
	go build -buildmode=c-shared -o sstablekeys.so src/SSTableKeysNode.go

build: compile-cmd
	npm install
	npm run build
