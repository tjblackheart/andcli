#!/usr/bin/env just

tag     := env('TAG', `git describe --tags --abbrev=0`)
goos    := env('GOOS', `go env GOOS`)
goarch  := env('GOARCH', `go env GOARCH`)
ext     := env('EXT', '')
commit  := `git rev-parse --short HEAD`
now     := datetime('%F %T%z')
ldflags := ("
	-s -w
	-X 'github.com/tjblackheart/andcli/v2/internal/buildinfo.AppVersion="+tag+"'
	-X 'github.com/tjblackheart/andcli/v2/internal/buildinfo.Commit="+commit+"'
	-X 'github.com/tjblackheart/andcli/v2/internal/buildinfo.BuildDate="+now+"'
")

default: clean build

build:
	go build \
		-ldflags="{{ldflags}}" \
		-trimpath \
		-o builds/andcli_{{tag}}_{{goos}}_{{goarch}}{{ext}} \
		./cmd/andcli/...

compress: build
	upx builds/andcli*

clean:
	rm -f builds/*

docs:
	export ANDCLI_HIDE_ABSPATH=1; vhs < doc/demo.tape

test:
	go test -coverprofile .coverage ./...
	go tool cover -func .coverage
	rm -f .coverage

test_race:
	go test -race ./...

fuzz all="":
  #!/usr/bin/env bash
  vaults=(andotp stratum protonpass twofas protonauth keepass) # aegis/ente need a kdf cap first
  all="{{all}}"
  for v in "${vaults[@]}"; do
    if [[ "$all" != "-a" ]]; then
      read -r -p "Fuzz $v (30s)? [Y/n] " reply
      [[ $reply == n ]] && continue
    fi
    go test -run='^$' -fuzz=FuzzOpen -fuzztime=30s "./internal/vaults/$v"
  done
