EXE = podbit

UISRC = ui/ui.go ui/list.go ui/player.go ui/rawplayer.go
INPUTSRC = input/input.go
SRC = main.go ver.go ${INPUTSRC} ${UISRC}

export CGO_CFLAGS_ALLOW=".*"
export CGO_LDFLAGS_ALLOW=".*"

${EXE}: ${SRC}
	go build

clean:
	go clean

install:
	go install

uninstall:
	go clean
	rm -f ${GOPATH}/bin/${EXE}

.PHONY = clean
