EXE = podbit

UISRC    = ui/ui.go ui/input.go colors/colors.go ui/library.go ui/player.go ui/queue.go ui/download.go ui/tray.go
UICOMPS  = ui/components/menu.go ui/components/table.go
SOUNDSRC = sound/sound.go sound/queue.go
DATASRC  = data/data.go data/queue.go data/db.go data/cache.go
SRC = main.go ver.go ${INPUTSRC} ${UISRC} ${DATASRC} ${UICOMPS} ${SOUNDSRC}

${EXE}: ${SRC}
	CGO_LDFLAGS_ALLOW=".*" go build

check:
	CGO_LDFLAGS_ALLOW=".*" go run -race . 2>race.log
clean:
	go clean

install: ${EXE}
	go install
	make

uninstall:
	go clean
	rm -f ${GOPATH}/bin/${EXE}

.PHONY = check clean install uninstall
