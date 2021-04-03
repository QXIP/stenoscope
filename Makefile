default: build

go-get:
	go get ./...

compile-cmd:
	go build -o SSTableKeys src/SSTableKeys.go

compile-go:
	go build -buildmode=c-shared -o sstablekeys.so src/SSTableKeysNode.go

build: go-get
	compile-cmd
	compile-go
	npm install
	npm run build
