.PHONY: ios build install

BINARY_NAME=ws_wrapper/cmd/open_im_sdk_server
BIN_DIR=../../bin/
LAN_FILE=.go
GO_FILE:=${BINARY_NAME}${LAN_FILE}

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
install:
	make build
	mv ${BINARY_NAME} ${BIN_DIR}
clean:
	env GO111MODULE=on go clean -cache
	gomobile clean
	rm -fr build

reset_remote_branch:
	remote_branch=$(shell git rev-parse --abbrev-ref --symbolic-full-name @{u})
	git reset --hard $(remote_branch)
	git pull $(remote_branch)

ios:
	go get golang.org/x/mobile
	go mod download golang.org/x/exp
	GOARCH=arm64 gomobile bind -v -trimpath -ldflags "-s -w" -o build/OpenIMCore.xcframework -target=ios ./open_im_sdk/ ./open_im_sdk_callback/

android:
	go get golang.org/x/mobile
	go mod download golang.org/x/exp
	gomobile bind -v -trimpath -ldflags "-s -w" -o build/OpenIMCore.aar -target=android ./open_im_sdk/ ./open_im_sdk_callback/

