package ev

// Event classes.
// ev.Handler passes events as integers between goroutines. These integers are
// not arbitrary and are selected by the caller. This allows the receiving
// goroutine to select which events are relevant to it using simple arithmetic
// and/or a switch case.
const (
	Keystroke = iota
	Resize
	PlayerChanged
	DownloadChanged
)
