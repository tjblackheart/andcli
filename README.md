# andcli

[![Go Report Card](https://goreportcard.com/badge/github.com/tjblackheart/andcli)](https://goreportcard.com/report/github.com/tjblackheart/andcli) ![Build](https://github.com/tjblackheart/andcli/actions/workflows/build.yaml/badge.svg)

View, decrypt and copy 2FA tokens from encrypted backup files directly in your shell. At the moment andcli can handle encrypted backups from the following providers:

* [andotp](https://github.com/andOTP/andOTP)
* [aegis](https://getaegis.app)
* [twofas](https://2fas.com)

![Demo](doc/demo.gif "Demo")

## Installation

Download a [prebuild release](https://github.com/tjblackheart/andcli/releases) and place it somewhere in your $PATH. If you have Go installed you can build it yourself: `go install -v github.com/tjblackheart/andcli/cmd/andcli@latest`.

## Usage

1. Export an **encrypted, password protected** backup from your 2FA app and save it into your preferred cloud provider (i.e. Dropbox, Nextcloud...).
2. Fire up `andcli` and point it to this file with `-f <path-to-file>`. Specify the vault type via `-t <type>`: choose between `andotp` or `aegis` or `twofas`. The path and type will get cached, so you have to do this only once.
3. Enter the encryption password.
4. To search an entry, type a word. Press `ESC` to clear the current query.
5. Navigate via keyboard, press `Enter` to view a token and press `c` to copy it into the clipboard.
6. If you are running Linux: Press the middle mouse button to paste the token. On Mac, hit CMD+v. On Windows, hit Ctrl+v.

## TODO

* ~~At the moment it is not possible to copy a token on a Windows machine.~~
* The test coverage sucks (less).
* ~~Implement a search.~~

## Options

```bash
Usage of andcli:
  -f string
        Path to the encrypted vault
  -t string
        Vault type (andotp, aegis, twofas)
  -c string
        Clipboard command (by default is the first of `xclip`, `wl-copy` or `pbcopy` found)
  -v    Show current version
```

## Credits

* [Bubbletea](https://github.com/charmbracelet/bubbletea)
* [GoTP](https://github.com/xlzd/gotp)
* [go-andotp](https://github.com/grijul/go-andotp)
* [color](https://github.com/fatih/color)
* [vhs](https://github.com/charmbracelet/vhs)

## License

[MIT](LICENSE.md)
