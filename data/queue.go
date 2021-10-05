package data

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Queue parsing/management errors
var (
	QueueNotFound    error  = errors.New("Error: Failed to locate newsboat queue file")
	QueueNotOpen     error  = errors.New("Error: Cannot parse a queue that is not open")
	QueueIOFailed    error  = errors.New("Error: IO error while reading from queue file")
	QueueSyntaxError string = "Error: Malformed queue: Syntax error on line %d"
)

var QUEUE_DIRS = []string{
	".local/share/newsboat",
	".newsboat",
}

// Name of the file for the queue
const QUEUE_FILE = "queue"

// Possible states of download queue
const (
	STATE_PENDING  = iota // Pending download
	STATE_READY           // Downloaded and ready to play
	STATE_PLAYED          // Played at least once
	STATE_FINISHED        // Finished to the end
)

type QueueItem struct {
	url   string
	path  string
	state int
}

type Queue struct {
	path string
	file *os.File

	items []QueueItem
}

func (q *Queue) parseField(fields []string, num int) {
	var item QueueItem

	item.url = fields[0]
	item.path = strings.ReplaceAll(fields[1], "\"", "")

	if num == 2 {
		item.state = STATE_PENDING
	} else {
		switch fields[2] {
		case "downloaded":
			item.state = STATE_READY
		case "played":
			item.state = STATE_PLAYED
		case "finished":
			item.state = STATE_FINISHED
		default:
			item.state = STATE_PENDING
		}
	}

	q.items = append(q.items, item)
}

func (q *Queue) Open() error {
	// First try the most likely places
	var err error
	var found bool = false
	home, _ := os.UserHomeDir()
	data := os.Getenv("XDG_DATA_HOME")

	for _, elem := range QUEUE_DIRS {
		q.path = filepath.Join(home, elem, QUEUE_FILE)
		q.file, err = os.Open(q.path)

		if err == nil {
			found = true
			break
		}
	}

	// Next try XDG
	if !found {
		q.path = filepath.Join(home, data, "newsboat", QUEUE_FILE)
		q.file, err = os.Open(q.path)

		if err == nil {
			found = true
		}
	}

	// If we still haven't found it, we never will
	if !found {
		return QueueNotFound
	}

	scanner := bufio.NewScanner(q.file)
	scanner.Split(bufio.ScanLines)

	i := 1
	for scanner.Scan() {
		if scanner.Err() != nil {
			return QueueIOFailed
		}

		elem := scanner.Text()
		fields := strings.Split(elem, " ")
		num := len(fields)

		if num < 2 {
			return fmt.Errorf(QueueSyntaxError, i)
		}

		q.parseField(fields, num)

		i++
	}

	return nil
}

func (q *Queue) Reload() {
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

		for _, elem := range q.items {
			if elem.url == fields[0] {
				continue scanloop
			}
		}

		q.parseField(fields, num)

	}
}
