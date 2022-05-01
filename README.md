# Podbit - **Podboat Improved**

![Podbit Player](https://www.ethanjmarshall.co.uk/wp-content/uploads/2022/04/screenshot-220408-2219-04.png)

[![AUR Release](https://img.shields.io/aur/version/podbit?color=1793d1&label=podbit&logo=arch-linux)](https://aur.archlinux.org/packages/podbit/)
[![Makefile CI](https://github.com/ethanv2/podbit/actions/workflows/makefile.yml/badge.svg)](https://github.com/ethanv2/podbit/actions/workflows/makefile.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ethanv2/podbit)](https://goreportcard.com/report/github.com/ethanv2/podbit)

Podbit is a replacement for ``newsboat``'s standard podboat tool for listening to podcasts. It is minimal, performant and abides by the UNIX-philosophy, with an ncurses terminal user interface.

Podbit runs on Linux and the BSDs.

## Features

* Reads the standard ``newsboat`` queue file to integrate seamlessly
* Automatic podcast downloading, including in parallel
* Podcast playing using ``mpv``
* Podcast caching and automatic deletion once finished

## Requirements

Podbit is written in Go. So, to build, you will need a copy of the Go command line tool. In addition, you will need:

* *ncurses* development libraries and headers
* A normal install of ``mpv``
* A copy of GNU Make
* Newsboat to enqueue podcasts - *(optional)*
* A YouTube downloader tool, such as ``youtube-dl`` or ``yt-dlp``, to download YouTube podcasts - *(optional)*

Because of security issues in the Go tool, the provided Makefile must be used instead of simply ``go build``.

## Unicode support

By default, Unicode support is not built in for compatability reasons. If you require unicode support, you will need to build the program slightly differently. Run ``make clean`` and then the following:

```bash
sed -i '/rthornton128\/goncurses/d' go.sum go.mod
for f in $(grep -rl "github.com/rthornton128/goncurses")
do
	sed -i 's/rthornton128\/goncurses/vit1251\/go-ncursesw/' "$f"
done

go mod tidy
make
```

After this, the program should more-or-less perfectly support unicode, as we deal mainly with uninterpreted text. There are some known issues with text truncation at the edges of the screen (slicing mid-rune, for example).

When reporting bugs, please ensure you **mention if you used these steps**! Please note that using these steps will make the program useless on many platforms (which is why it is not built in). This includes Windows (pdcurses does not support wchars) and OpenBSD (for some reason).
