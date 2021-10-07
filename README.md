# Podbit - **Podboat Improved**

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

Because of security issues in the Go tool, the provided Makefile must be used instead of simply ``go build``.