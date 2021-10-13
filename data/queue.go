package data

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// Range function typedefs
type RangeFunc func(i int, item *QueueItem) bool

// Queue parsing/management errors
var (
	ErrorNotFound    error  = errors.New("Error: Failed to locate newsboat queue file")
	ErrorNotOpen     error  = errors.New("Error: Cannot parse a queue that is not open")
	ErrorIOFailed    error  = errors.New("Error: IO error while reading from queue file")
	ErrorQueueSyntax string = "Error: Malformed queue: Syntax error on line %d"
)

// PossibleDirs are the locations where the queue will search for a newsboat
// queue file
var PossibleDirs = []string{
	".local/share/newsboat",
	".newsboat",
}

// QueueFilename is the name of the file for the queue
const QueueFilename = "queue"

// Possible states of download queue
const (
	StatePending  = iota // Pending download
	StateReady           // Downloaded and ready to play
	StatePlayed          // Played at least once
	StateFinished        // Finished to the end
)

// StateStrings are the names used to serialise or display queue
// statuses to the user
var StateStrings [4]string = [4]string{
	"",
	"downloaded",
	"played",
	"finished",
}

// QueueItem represents an item in the player queue
// as provided by newsboat
type QueueItem struct {
	URL   string
	Path  string
	State int
}

// Queue represents the newsboat queue
type Queue struct {
	path string
	file *os.File

	mutex   sync.RWMutex
	Items []QueueItem
}

func (q *Queue) parseField(fields []string, num int) {
	var item QueueItem

	item.URL = fields[0]
	item.Path = strings.ReplaceAll(fields[1], "\"", "")

	f, err := os.Open(item.Path)
	if num == 2 || (err != nil && os.IsNotExist(err)) {
		item.State = StatePending
	} else {
		switch fields[2] {
		case "downloaded":
			item.State = StateReady
		case "played":
			item.State = StatePlayed
		case "finished":
			item.State = StateFinished
		default:
			item.State = StateReady
		}
	}

	f.Close()
	q.Items = append(q.Items, item)
}

// Open opens and parses the newsboat queue file
// Returned errors are usually fatal to the application
func (q *Queue) Open() error {
	// First try the most likely places
	var err error
	var found bool = false
	home, _ := os.UserHomeDir()
	data := os.Getenv("XDG_DATA_HOME")

	for _, elem := range PossibleDirs {
		q.path = filepath.Join(home, elem, QueueFilename)
		q.file, err = os.Open(q.path)

		if err == nil {
			found = true
			break
		}
	}

	// Next try XDG
	if !found {
		q.path = filepath.Join(home, data, "newsboat", QueueFilename)
		q.file, err = os.Open(q.path)

		if err == nil {
			found = true
		}
	}

	// If we still haven't found it, we never will
	if !found {
		return ErrorNotFound
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	scanner := bufio.NewScanner(q.file)
	scanner.Split(bufio.ScanLines)

	i := 1
	for scanner.Scan() {
		if scanner.Err() != nil {
			return ErrorIOFailed
		}

		elem := scanner.Text()
		fields := strings.Split(elem, " ")
		num := len(fields)

		if num < 2 {
			return fmt.Errorf(ErrorQueueSyntax, i)
		}

		q.parseField(fields, num)

		i++
	}

	return nil
}

// Reload performs a hot-reload
//
// Merges are performed on the simple basis that new lines are the
// data we want. All other changes are wiped and completely ignored.
func (q *Queue) Reload() {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	q.file.Close()

	var err error
	q.file, err = os.Open(q.path)
	if err != nil {
		fmt.Println("WARNING: Failed to open queue when reloading")
		return
	}

	scanner := bufio.NewScanner(q.file)

scanloop:
	for scanner.Scan() {
		if scanner.Err() != nil {
			fmt.Println("WARNING: Failed to reload queue")
			return
		}

		elem := scanner.Text()

		fields := strings.Split(elem, " ")
		num := len(fields)

		if num < 2 {
			fmt.Println("WARNING: Invalid queue found while reloading")
			return
		}

		for _, elem := range q.Items {
			if elem.URL == fields[0] {
				continue scanloop
			}
		}

		q.parseField(fields, num)

	}
}

// Save dumps the current state into the queue file, disregarding changes
// and without syncing contained state.
//
// The file is first truncated and blanked.
func (q *Queue) Save() {
	q.mutex.RLock()
	defer q.mutex.RUnlock()

	file, err := os.OpenFile(q.path, os.O_TRUNC|os.O_WRONLY, os.ModePerm)
	if err != nil {
		fmt.Printf("WARNING: Failed to save queue file: %s\n", err.Error())
	}

	for _, elem := range q.Items {
		fmt.Fprintf(file, "%s \"%s\" %s\n", elem.URL, elem.Path, StateStrings[elem.State])
	}
}

// Range loops through the queue array in a thread-safe fashion
// using a callback which recieves each item in the queue in the
// same format as a for range loop.
//
// It *IS* save to modify the queue in the callback
func (q *Queue) Range(callback RangeFunc) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for i, item := range q.Items {
		if !callback(i, &item) {
			return
		}
	}
}
