.PHONY: build

APPNAME := shortener
VERSION := $(shell echo "v1.0.1")
BUILD_DATE := $(shell date +'%Y/%m/%d %H:%M:%S')
COMMIT := $(shell echo "#1")

MAIN_PACKAGE_PATH := main

LDFLAGS := -ldflags "-w -s \
	-X $(MAIN_PACKAGE_PATH).BuildVersion=$(VERSION) \
	-X '$(MAIN_PACKAGE_PATH).BuildDate=$(BUILD_DATE)' \
	-X $(MAIN_PACKAGE_PATH).BuildCommit=$(COMMIT)"

build:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -buildvcs=false $(LDFLAGS) -o bin/$(APPNAME) ./cmd/$(APPNAME)/main.go

clean:
	rm -f bin/$(APPNAME)