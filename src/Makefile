BINARY_NAME=main.out

deps:
	go get github.com/libp2p/go-libp2p
	go get github.com/multiformats/go-multiaddr

all: build test

build:
	go build -o ${BINARY_NAME} main.go

test:
	go test -v main.go

clean:
	go clean
	rm ${BINARY_NAME}