# Podbit - **Podboat Improved**

![Podbit Player](https://raw.githubusercontent.com/ejv2/podbit/master/ss.png)

[![AUR Release](https://img.shields.io/aur/version/podbit?color=1793d1&label=podbit&logo=arch-linux)](https://aur.archlinux.org/packages/podbit/)
[![Makefile CI](https://github.com/ejv2/podbit/actions/workflows/makefile.yml/badge.svg)](https://github.com/ejv2/podbit/actions/workflows/makefile.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/ejv2/podbit)](https://goreportcard.com/report/github.com/ejv2/podbit)

Podbit is a replacement for ``newsboat``'s standard podboat tool for listening to podcasts. It is minimal, performant and tries to focus just on being a podcast client, rather than an RSS reader. Podbit has an ncurses terminal user interface and vi-like keybindings.

Podbit runs on Linux and the BSDs.

## Features

* Reads the standard ``newsboat`` queue file to integrate seamlessly
* Automatic podcast downloading, including in parallel
* Podcast playing using ``mpv``
* Podcast caching and automatic deletion once finished
* Vi-like "hjkl" to navigate the interface

## Requirements

Podbit is written in Go. So, to build, you will need a copy of the Go command line tool. In addition, you will need:

* *ncurses* development libraries and headers, including wide character support (*libncusesw*)
* A normal install of ``mpv``
* A copy of GNU Make
* Newsboat to enqueue podcasts - *(optional)*
* A YouTube downloader tool, such as ``youtube-dl`` or ``yt-dlp``, to download YouTube podcasts - *(optional)*

Because of security issues in the Go tool, the provided Makefile must be used instead of simply ``go build``.
