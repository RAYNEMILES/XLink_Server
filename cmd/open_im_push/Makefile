.PHONY: all build run gotool install clean help

BINARY_NAME=open_im_push
BIN_DIR=../../bin/

all: gotool build

build:
    ifeq ($(shell arch),arm64)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++ go build -ldflags="-w -s"

else	ifeq ($(shell arch),aarch64)
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 go build -ldflags="-w -s"

else	ifeq ($(shell arch),i386)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 CC=x86_64-linux-musl-gcc  CXX=x86_64-linux-musl-g++ go build -ldflags="-w -s"
	
else
	$(warning  $(shell arch))
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s"
endif

run:
	@go run ./

gotool:
	go fmt ./
	go vet ./

install:
	make build
	mv ${BINARY_NAME} ${BIN_DIR}

clean:
	@if [ -f ${BINARY_NAME} ] ; then rm ${BINARY_NAME} ; fi


