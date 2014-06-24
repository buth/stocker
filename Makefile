GOPATH = $(shell pwd)/.tmp
GOBIN = ${GOPATH}/bin

install: all
	cp ${GOBIN}/stocker /usr/local/bin

all: ${GOBIN}/stocker

${GOBIN}/stocker: deps
	go install -v github.com/buth/stocker

deps: ${GOPATH}/src/github.com/buth/stocker
	go get -v -d github.com/buth/stocker

${GOPATH}/src/github.com/buth/stocker: ${GOPATH}
	mkdir -p ${GOPATH}/src/github.com/buth/stocker
	rsync -av --exclude /.tmp --exclude .git --exclude-from .gitignore ./ ${GOPATH}/src/github.com/buth/stocker/

${GOBIN}: ${GOPATH}
	mkdir -p ${GOBIN}

${GOPATH}:
	mkdir -p ${GOPATH}

clean:
	rm -rf .tmp
