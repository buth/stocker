VERSION = $(shell cat VERSION)
GOPATH = $(shell pwd)/.gopath
GOBIN = ${GOPATH}/bin
BUILDS = $(shell pwd)/.builds

install: all
	cp ${GOBIN}/stocker /usr/local/bin

all: ${GOBIN}/stocker

${GOBIN}/stocker: dependencies
	go install -v github.com/buth/stocker

test: dependencies
	go test -v github.com/buth/stocker/...

release: dependencies ${BUILDS}/stocker-${VERSION}
	gox -output="${BUILDS}/stocker-${VERSION}-{{.OS}}-{{.Arch}}/bin/stocker" -osarch="linux/arm linux/386 linux/amd64 darwin/amd64" github.com/buth/stocker

dependencies: ${GOPATH}/src/github.com/buth/stocker
	go get -v -d github.com/buth/stocker

${BUILDS}/stocker-${VERSION}: ${GOPATH}/src/github.com/buth/stocker ${BUILDS}
	cp -r ${GOPATH}/src/github.com/buth/stocker ${BUILDS}/stocker-${VERSION}

${GOPATH}/src/github.com/buth/stocker: ${GOPATH}
	mkdir -p ${GOPATH}/src/github.com/buth/stocker
	rsync -av --exclude .git --exclude-from .gitignore ./ ${GOPATH}/src/github.com/buth/stocker/

${GOBIN}: ${GOPATH}
	mkdir -p ${GOBIN}

${GOPATH}:
	mkdir -p ${GOPATH}

${BUILDS}:
	mkdir -p ${BUILDS}

clean:
	rm -rf ${GOPATH} ${BUILDS}
