BINARY_NAME=rettiwt

all: build test

build: 
	go build -o ${BINARY_NAME} ./

test:
	go test -v ./

run:
	go run ./

clean:
	go clean
	rm ${BINARY_NAME}