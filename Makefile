TAG=$(shell git describe --tags --abbrev=0)
COMMIT=$(shell git rev-parse --short HEAD)
NOW=$(shell date "+%F %T%:z")
FLAGS=-s -w \
	-X 'github.com/tjblackheart/andcli/internal/buildinfo.Commit=$(COMMIT)' \
	-X 'github.com/tjblackheart/andcli/internal/buildinfo.BuildDate=$(NOW)'

build:
	go build -ldflags "$(FLAGS) -X 'github.com/tjblackheart/andcli/internal/buildinfo.AppVersion=$(TAG)'" -trimpath -o builds/andcli ./cmd/andcli

ci:
	go build -ldflags "$(FLAGS) -X 'github.com/tjblackheart/andcli/internal/buildinfo.AppVersion=$(CI_TAG)'" -trimpath -o builds/andcli_$(RELEASE) ./cmd/andcli

compress: build
	upx builds/andcli*

clean:
	rm -rf builds/*

docs:
	export ANDCLI_HIDE_ABSPATH=1; vhs < doc/demo.tape

test:
	go test -coverprofile .coverage ./...
	go tool cover -func .coverage
	rm -f .coverage
