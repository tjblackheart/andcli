GOVER=$(shell go version | sed 's/^.*go\([0-9.]*\).*/\1/')
COMMIT=$(shell git rev-parse --short HEAD)
# macOS/BSD compatible equivalent of `date --rfc-3339=seconds`
NOW=$(shell date "+%F %T%:z")
FLAGS=-s -w -X 'main.commit=$(COMMIT)' -X 'main.gover=$(GOVER)' -X 'main.date=$(NOW)'

# set local vars without pipeline access
TAG=$(shell git describe --tags --abbrev=0)
ARCH=$(shell go env GOARCH)

build: clean
	go build -ldflags="$(FLAGS) -X 'main.tag=$(TAG)' -X 'main.arch=$(ARCH)'" -o bin/andcli ./...

ci:
	go build -ldflags="$(FLAGS) -X 'main.tag=$(CI_TAG)' -X 'main.arch=$(GOARCH)'" -o bin/andcli_$(RELEASE) ./...

compress: build
	upx bin/andcli

clean:
	rm -rf bin/*

docs:
	export ANDCLI_HIDE_ABSPATH=1; vhs < doc/demo.tape
