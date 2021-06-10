PKG := github.com/wtsi-ssg/wrstat
VERSION := $(shell git describe --tags --always --long --dirty)
TAG := $(shell git describe --abbrev=0 --tags)
LDFLAGS = -ldflags "-X ${PKG}/cmd.Version=${VERSION}"
export GOPATH := $(shell go env GOPATH)
PATH := $(PATH):${GOPATH}/bin

default: install

build: export CGO_ENABLED = 0
build:
	go build -tags netgo ${LDFLAGS}

install: export CGO_ENABLED = 0
install:
	@rm -f ${GOPATH}/bin/wrstat
	@go install -tags netgo ${LDFLAGS}
	@echo installed to ${GOPATH}/bin/wrstat

test: export CGO_ENABLED = 0
test:
	@go test -tags netgo --count 1 ./...

race: export CGO_ENABLED = 1
race:
	go test -tags netgo -race --count 1 ./...

bench: export CGO_ENABLED = 1
bench:
	go test -tags netgo --count 1 -run Bench -bench=. ./...

# curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.39.0
lint:
	@golangci-lint run

clean:
	@rm -f ./wrstat
	@rm -f ./dist.zip

dist: export CGO_ENABLED = 0
# go get -u github.com/gobuild/gopack
# go get -u github.com/aktau/github-release
dist:
	gopack pack --os linux --arch amd64 -o linux-dist.zip
	github-release release --tag ${TAG} --pre-release
	github-release upload --tag ${TAG} --name wr-linux-x86-64.zip --file linux-dist.zip
	@rm -f wr linux-dist.zip

.PHONY: test race bench lint build install clean dist