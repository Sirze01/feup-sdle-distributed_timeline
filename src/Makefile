BINARY_NAME=rettiwt

all: build test

build: 
	go build -o ${BINARY_NAME} *.go

test:
	go test -v *.go

run:
	go run *.go

clean_nodes:
	rm -rf nodes/*

clean:
	go clean
	rm ${BINARY_NAME}