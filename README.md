# andcli

[![Go Report Card](https://goreportcard.com/badge/github.com/tjblackheart/andcli)](https://goreportcard.com/report/github.com/tjblackheart/andcli) ![Build](https://github.com/tjblackheart/andcli/actions/workflows/build.yaml/badge.svg)

andcli lets you work with 2FA tokens directly in your shell, using encrypted backups exported out of your favourite 2FA apps. All the data is held in memory only and will never leave your machine.

At the moment andcli can handle input from the following providers:

* [andotp](https://github.com/andOTP/andOTP)
* [aegis](https://getaegis.app)
* [twofas](https://2fas.com)

![Demo](doc/demo.gif "Demo")

## Installation

Download a [prebuild release](https://github.com/tjblackheart/andcli/releases) and place it somewhere in your $PATH. If you have Go installed you can build it yourself: `go install -v github.com/tjblackheart/andcli/cmd@latest`.

## Usage

1. Export an **encrypted, password protected** backup from your app and save it into your preferred cloud provider (i.e. Dropbox, Nextcloud...).
2. Start `andcli` and point it to this file with `andcli <path/to/file>`. Specify the vault type via `-t <type>`. The path and type will be persisted, so you have to do this only once.
3. Enter the password.
4. To search for an entry, type `/`.
5. Navigate via keyboard, press `Enter` to view a token and press `c` to copy it into the clipboard. Press `u` to reveal usernames for this entry, which are hidden by default.

## Keys

```text
↑/k up
↓/j down
/ filter
enter show/hide token
u show/hide usernames
c copy
q quit
```

## Clipboard config

By default, andcli will choose the first system clipboard tool found. For Linux, this could be either `xclip`, `xsel` or `wl-copy` for example. For Mac, it will always be `pbcopy`. If you need more control over this command, you can either edit the config file and set your preferred command including all flags in the `clipboard_cmd` entry, or you could pass the full command via the `-c` flag.

## Config file

The configuration will get persisted in the default user home config directory. For Linux, this is `$HOME/.config/andcli`. For MacOS, it's `$HOME/Library/Application Support/andcli` and for Windows it should be in `C:\Users\$USER\AppData\Roaming\andcli`.

## Options

```text
Usage: andcli [options] <path/to/file>

Options:
  -c string
    	Clipboard command (xclip, wl-copy, pbcopy etc.)
  -f string
    	Path to the encrypted vault (deprecated: Pass the filename directly)
  -t string
    	Vault type (andotp, aegis, twofas)
  -v	Prints version info and exits
```

## Implementing new vaults

A usabe vault implementation for andcli is basically just an interface providing one function called `Entries()`. You just have to figure out the encryption of your app and you're done. Have a look at the [current implementations](internal/vaults) to see how this works.

You can use the demo registration server implementation at [tools/srv](tools/srv) to quickly create some demo tokens for your vault.

## Thanks

* [Bubbletea](https://github.com/charmbracelet/bubbletea)
* [lipgloss](https://github.com/charmbracelet/lipgloss)
* [GoTP](https://github.com/xlzd/gotp)
* [go-andotp](https://github.com/grijul/go-andotp)
* [vhs](https://github.com/charmbracelet/vhs)

## License

[MIT](LICENSE.md)
