#!/usr/bin/env just

# to override vars at runtime: just --set tag v1.0.0
tag     := `git describe --tags --abbrev=0`
commit  := `git rev-parse --short HEAD`
now     := datetime('%F %T%z')
ldflags := ("
	-s -w
    -X 'github.com/tjblackheart/andcli/v2/internal/buildinfo.AppVersion="+tag+"'
    -X 'github.com/tjblackheart/andcli/v2/internal/buildinfo.Commit="+commit+"'
    -X 'github.com/tjblackheart/andcli/v2/internal/buildinfo.BuildDate="+now+"'
")

all: clean build

build:
    go build -ldflags="{{ldflags}}" -trimpath -o builds/andcli ./cmd/andcli/...

ci:
    go build -ldflags="{{ldflags}}" -trimpath -o builds/andcli_`echo $RELEASE` ./cmd/andcli/...

compress: build
	upx builds/andcli*

clean:
	rm -f builds/*

docs:
	export ANDCLI_HIDE_ABSPATH=1
	vhs < doc/demo.tape

test:
	go test -coverprofile .coverage ./...
	go tool cover -func .coverage
	rm -f .coverage
