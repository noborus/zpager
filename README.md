# ov - Oviewer

[![PkgGoDev](https://pkg.go.dev/badge/github.com/noborus/ov)](https://pkg.go.dev/github.com/noborus/ov)
[![Actions Status](https://github.com/noborus/ov/workflows/Go/badge.svg)](https://github.com/noborus/ov/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/noborus/ov)](https://goreportcard.com/report/github.com/noborus/ov)

ov is a feature rich terminal pager.
It has an effective function for tabular text.

(The old repository name was oviewer.)

![ov.gif](https://raw.githubusercontent.com/noborus/ov/master/docs/ov.gif)

## feature

* Better support for Unicode and East Asian Width.
* Support for compressed files (gzip, bzip2, zstd, lz4, xz).
* Supports column mode.
* Header rows can be fixed.
* Dynamic wrap / nowrap switchable.
* Background color to alternate rows.
* Columns can be selected with separators.
* Shortcut keys are customizable.
* The style of the effect is customizable.

## install

### deb package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.deb
sudo dpkg -i ov_x.x.x-1_amd64.deb
```

### rpm package

You can download the package from [releases](https://github.com/noborus/ov/releases).

```console
sudo rpm -ivh https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x-1_amd64.rpm
```

### Homebrew(macOS or Linux)

```console
brew install noborus/tap/ov
```

### binary

You can download the binary from [releases](https://github.com/noborus/ov/releases).

```console
curl -L -O https://github.com/noborus/ov/releases/download/vx.x.x/ov_x.x.x_linux_amd64.zip
unzip ov_x.x.x_linux_amd64.zip
sudo install ov /usr/local/bin
```

### go get(simplified version)

It will be installed in $GOPATH/bin by the following command.

```console
go get -u github.com/noborus/ov
```

### go get(details or developer version)

First of all, download only with the following command without installing it.

```console
go get -d github.com/noborus/ov
cd $GOPATH/src/github.com/noborus/ov
```

Next, to install to $GOPATH/bin, run the make install command.

```console
make install
```

Or, install it in a PATH location for other users to use
(For example, in /usr/local/bin).

```console
make
sudo install ov /usr/local/bin
```

## Usage

ov supports open file name or standard input.

```console
ov filename
```

```console
cat filename|ov
```

```console
$ ov --help
ov is a feature rich pager(such as more/less).
It supports various compressed files(gzip, bzip2, zstd, lz4, and xz).

Usage:
  ov [flags]

Flags:
  -C, --alternate-rows            color to alternate rows
  -i, --case-sensitive            case-sensitive in search
  -d, --column-delimiter string   column delimiter (default ",")
  -c, --column-mode               column mode
      --config string             config file (default is $HOME/.ov.yaml)
      --debug                     debug mode
      --disable-mouse             disable mouse support
  -X, --exit-write                output the current screen when exiting
  -H, --header int                number of header rows to fix
  -h, --help                      help for ov
      --help-key                  display key bind information
  -n, --line-number               line number
  -F, --quit-if-one-screen        quit if the output fits on one screen
  -x, --tab-width int             tab stop width (default 8)
  -v, --version                   display version information
  -w, --wrap                      wrap mode (default true)
```

It can also be changed after startup.
Refer to the [motion image](docs/image.md).

## config

You can set style and key bindings in the setting file.

Please refer to the sample [ov.yaml](https://github.com/noborus/ov/blob/master/ov.yaml) configuration file.

### psql

Set environment variable `PSQL_PAGER`(PostgreSQL 11 or later).

```sh
export PSQL_PAGER='ov -w=f -H2 -F -C -d "|"'
```

You can also write in `~/.psqlrc` in previous versions.

```filename:~/.psqlrc
\setenv PAGER 'ov -w=f -H2 -F -C -d "|"'
```

### mysql

Use the --pager option with the mysql client.

```console
mysql --pager='ov -w=f -H3 -F -C -d "|"'
```

You can also write in `~/.my.cnf`.

```filename:~/.my.cnf
[client]
pager=ov -w=f -H3 -F -C -d "|"
```

## Mouse support

The ov makes the mouse support its control.
This can be disabled with the option `--disable-mouse`.

If mouse support is enabled, tabs and line breaks will be interpreted correctly when copying.

Copying to the clipboard uses [atotto/clipboard](https://github.com/atotto/clipboard).
For this reason, the 'xclip' or 'xsel' command is required in Linux/Unix environments.

Selecting the range with the mouse and then left-clicking will copy it to the clipboard.

Pasting in ov is done with the middle button.
In other applications, it is pasted from the clipboard (often by pressing the right-click).

## Key bindings

```
  [Escape], [q]              * quit
  [ctrl+c]                   * cancel
  [Q]                        * output screen and quit
  [h]                        * display help screen
  [ctrl+alt+e]               * display log screen
  [ctrl+l]                   * screen sync
  [ctrl+alt+r]               * enable/disable mouse
  [ctrl+k]                   * close current document

	Moving

  [Enter], [Down], [ctrl+N]  * forward by one line
  [Up], [ctrl+p]             * backward by one line
  [Home]                     * go to begin of line
  [End]                      * go to end of line
  [PageDown], [ctrl+v]       * forward by page
  [PageUp], [ctrl+b]         * backward by page
  [ctrl+d]                   * forward a half page
  [ctrl+u]                   * backward a half page
  [left]                     * scroll to left
  [right]                    * scroll to right
  [ctrl+left]                * scroll left half screen
  [ctrl+right]               * scroll right half screen
  [g]                        * number of go to line
  []]                        * next document
  [[]                        * previous document

	Mark position

  [m]                        * mark current position
  [>]                        * move to next marked position
  [<]                        * move to previous marked position

	Search

  [/]                        * forward search mode
  [?]                        * backward search mode
  [n]                        * repeat forward search
  [N]                        * repeat backward search

	Change display

  [w], [W]                   * wrap/nowrap toggle
  [c]                        * column mode toggle
  [C]                        * color to alternate rows toggle
  [G]                        * line number toggle

	Change Display with Input

  [d]                        * delimiter string
  [H]                        * number of header lines
  [t]                        * TAB width

```

## Customize

### Style customization

You can customize the following items.

* StyleAlternate
* StyleHeader
* StyleOverStrike
* StyleOverLine

Specifies the color name for the foreground and background colors.
Specify bool values for Bold, Blink, Shaded, Italic, and Underline.

[Example]

```yaml
StyleAlternate:
  Background: "gray"
  Bold: true
  Underline: true
```

### Key binding customization

You can customize key bindings.

[Example]

```yaml
    down:
        - "Enter"
        - "Down"
        - "ctrl+N"
    up:
        - "Up"
        - "ctrl+p"
```

See [ov.yaml](https://github.com/noborus/ov/blob/master/ov.yaml) for more information..
