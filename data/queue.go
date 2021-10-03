package data

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"bufio"
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

const (
	QUEUE_FILE = "queue"
)

type QueueItem struct {
	url   string
	path  string
	state string
}

type Queue struct {
	path string
	file *os.File

	items []QueueItem
}

func (q *Queue) Open() error {
	// First try the most likely places
	var err error
	home, _ := os.UserHomeDir()
	data := os.Getenv("XDG_DATA_HOME")

	for _, elem := range QUEUE_DIRS {
		q.path = filepath.Join(home, elem, QUEUE_FILE)
		q.file, err = os.Open(q.path)

		if err == nil {
			return nil
		}
	}

	// Next try XDG
	q.path = filepath.Join(home, data, "newsboat", QUEUE_FILE)
	q.file, err = os.Open(q.path)

	if err == nil {
		return nil
	}

	return QueueNotFound
}

func (q *Queue) Parse() error {
	if q.file == nil {
		return QueueNotOpen
	}

	scanner := bufio.NewScanner(q.file)
	scanner.Split(bufio.ScanLines)

	i := 1
	for scanner.Scan() {
		if scanner.Err() != nil {
			return QueueIOFailed
		}

		elem := scanner.Text()

		var item QueueItem
		fields := strings.Split(elem, " ")
		num := len(fields)

		if num < 2 {
			return fmt.Errorf(QueueSyntaxError, i)
		}


		item.url = fields[0]
		item.path = strings.ReplaceAll(fields[1], "\"", "")

		if num == 2 {
			item.state = "pending"
		} else {
			item.state = fields[2]
		}

		q.items = append(q.items, item)
		i++
	}

	return nil
}
