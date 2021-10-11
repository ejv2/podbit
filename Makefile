EXE = podbit

UISRC    = ui/ui.go ui/input.go ui/colors.go ui/library.go ui/player.go ui/rawplayer.go ui/queue.go ui/tray.go
UICOMPS  = ui/components/menu.go
SOUNDSRC = sound/sound.go sound/queue.go
DATASRC  = data/data.go data/queue.go data/db.go data/cache.go
SRC = main.go ver.go ${INPUTSRC} ${UISRC} ${DATASRC} ${UICOMPS} ${SOUNDSRC}

${EXE}: ${SRC}
	CGO_LDFLAGS_ALLOW=".*" go build

clean:
	go clean

install:
	go install

uninstall:
	go clean
	rm -f ${GOPATH}/bin/${EXE}

.PHONY = clean
