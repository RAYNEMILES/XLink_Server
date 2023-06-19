.PHONY: all build run gotool install clean help mac

BINARY_NAME=main
BIN_DIR=./
LAN_FILE=.go
GO_FILE:=./

all: gotool build

build:
    ifeq ($(shell arch),arm64)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++ go build -ldflags="-w -s" -o  ${BINARY_NAME}  ${GO_FILE}

else	ifeq ($(shell arch),aarch64)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s" -o  ${BINARY_NAME}  ${GO_FILE}

else	ifeq ($(shell arch),i386)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++ go build -ldflags="-w -s" -o  ${BINARY_NAME}  ${GO_FILE}
	
else
	$(warning  $(shell arch))
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o  ${BINARY_NAME}  ${GO_FILE}
endif

run:
	@go run ./

gotool:
	go fmt ./
	go vet ./

install:
	make build
	mv main press_test

clean:
	@if [ -f ${BINARY_NAME} ] ; then rm ${BINARY_NAME} ; fi

mac:
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags="-w -s" -o  ${BINARY_NAME}  ${GO_FILE}
	mv main press_test_mac


