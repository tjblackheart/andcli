# andcli

[![Go Report Card](https://goreportcard.com/badge/github.com/tjblackheart/andcli)](https://goreportcard.com/report/github.com/tjblackheart/andcli) ![Build](https://github.com/tjblackheart/andcli/actions/workflows/build.yaml/badge.svg)

andcli lets you work with 2FA tokens directly in your shell, using encrypted backups exported out of your favourite 2FA apps. All the data is held in memory only and will never leave your machine.

andcli can handle input from the following providers:

* [andotp](https://github.com/andOTP/andOTP)
* [Aegis](https://getaegis.app)
* [2fas](https://2fas.com)
* [Stratum / Authenticator Pro](https://stratumauth.com)
* [Keepass](https://www.keepassdx.com/)

At the moment only TOTP entries are supported.

![Demo](doc/demo.gif "Demo")

## Installation

Download a [prebuild release](https://github.com/tjblackheart/andcli/releases) and place it somewhere in your $PATH. If you have Go installed you can install directly: `go install -ldflags='-s -w' github.com/tjblackheart/andcli/v2/cmd/andcli@latest`.

## Usage

1. Export an **encrypted, password protected** backup from your app and save it into your preferred cloud provider (i.e. Dropbox, Nextcloud...).
2. Start `andcli` and point it to this file with `andcli <path/to/file>`. Specify the vault type via `-t <type>`. The path and type will be persisted, so you have to do this only once.
3. Enter the password.
4. To search for an entry, type `/`.
5. Navigate via keyboard, press `Enter` to view a token and press `c` to copy it into the clipboard. Press `u` to hide usernames for this entry, which are visible by default.

Since v2.1.3 it is possible to pipe the password from stdin and skip the input question: `echo $PASSWORD | andcli --passwd-stdin`

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

By default andcli will choose the first system clipboard tool found. For Linux, this could be either `xclip`, `xsel` or `wl-copy`. For Mac, it will always be `pbcopy`, and for a Windows machine it defaults to `clip.exe`. If you need more control over this command, you can either edit the config file and set your preferred command including all flags in the `clipboard_cmd` entry, or you could pass the full command via the `-c` flag.

## Config file

The configuration will get persisted in the default user home config directory. For Linux, this is `$HOME/.config/andcli`. For MacOS, it's `$HOME/Library/Application Support/andcli` and for Windows it should be in `C:\Users\$USER\AppData\Roaming\andcli`.

## Options

```text
Usage: andcli [options] <path/to/file>

Options:
  -c, --clipboard-cmd string   A custom clipboard command, including args (xclip, wl-copy, pbcopy etc.)
  -f, --file string            Path to the encrypted vault (deprecated: Pass the filename directly)
  -h, --help                   Show this help
      --passwd-stdin           Read the vault password from stdin. If set, skips the password input.
  -t, --type string            Vault type (andotp, aegis, twofas, stratum, keepass)
  -v, --version                Prints version info and exits
```

## Implementing new vaults

A usable vault implementation for andcli has to implement an interface providing only one function called `Entries()`. Have a look at the [current implementations](internal/vaults) to see how this works.

You can use the demo registration server implementation at [tools/srv](tools/srv) to quickly create some demo tokens for your vault.

## Thanks

* [Bubbletea](https://github.com/charmbracelet/bubbletea)
* [lipgloss](https://github.com/charmbracelet/lipgloss)
* [GoTP](https://github.com/xlzd/gotp)
* [go-andotp](https://github.com/grijul/go-andotp)
* [vhs](https://github.com/charmbracelet/vhs)
* [gokeepasslib](https://github.com/tobischo/gokeepasslib)

## License

[MIT](LICENSE.md)
