GOFILES = $(shell find . -name '*.go' -not -path './vendor/*')
GOPACKAGES = $(shell go list ./...  | grep -v /vendor/)
WORKDIR = workdir
VERSION = `git describe --always --long HEAD`
GO111MODULE = on

LDFLAGS = -ldflags "-w -s -X main.ctmVersion=${VERSION}"

default: build

clean:
	rm -rf $(WORKDIR)

build: build-native

build-native: $(GOFILES)
	go build ${LDFLAGS} -o $(WORKDIR)/ctm .

build-linux-x64: $(GOFILES) 
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build ${LDFLAGS} -o $(WORKDIR)/ctm .

build-linux: build-linux-x64 $(GOFILES)

build-windows:
	GOOS=windows GOARCH=386 CGO_ENABLED=0 go build ${LDFLAGS} -o $(WORKDIR)/ctm.exe .

test: test-all

test-all:
	@go test -v $(GOPACKAGES)
