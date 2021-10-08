EXE = podbit

UISRC    = ui/ui.go ui/list.go ui/player.go ui/rawplayer.go ui/tray.go
UICOMPS  = ui/components/menu.go
INPUTSRC = input/input.go
DATASRC  = data/data.go data/queue.go data/db.go
SRC = main.go ver.go ${INPUTSRC} ${UISRC} ${DATASRC} ${UICOMPS}

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
