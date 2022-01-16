package data

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	// DatabaseDirname is the directory name in which the database will be stored
	DatabaseDirname = "podbit"

	// DatabaseFilename is the file name of the database on disk
	DatabaseFilename = "db"
)

// Error values
var (
	ErrorDatabaseIOFailed error  = errors.New("Error: IO error while reading from database file")
	ErrorDatabaseSyntax   string = "Error: Malformed database: Syntax Error on Line %d"
)

// Podcast is the human-provided podcast info
type Podcast struct {
	RegexPattern string
	FriendlyName string
}

// Database aggregates all podcast data from the database
type Database struct {
	path     string
	podcasts []Podcast
}

func initDatabase(db *Database) error {
	var err error
	file, err := os.Open(db.path)
	defer file.Close()

	if err != nil {
		file, err = os.Create(db.path)
		if err != nil {
			return err
		}
	} else {
		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		i := 1
		for scanner.Scan() {
			var p Podcast

			if scanner.Err() != nil {
				return ErrorDatabaseIOFailed
			}

			elem := scanner.Text()
			fields := strings.Split(elem, " ")
			num := len(fields)

			if num < 2 {
				return fmt.Errorf(ErrorDatabaseSyntax, i)
			}

			p.RegexPattern = fields[0]
			p.FriendlyName = strings.Join(fields[1:], " ")

			db.podcasts = append(db.podcasts, p)

			i++
		}
	}

	return nil
}

// Open opens and parses the database
// Returned errors are usually fatal to the application
func (db *Database) Open() error {
	data := os.Getenv("XDG_DATA_HOME")
	if data == "" {
		home, _ := os.UserHomeDir()
		data = filepath.Join(home, ".local/share")
	}

	db.path = filepath.Join(data, DatabaseDirname, DatabaseFilename)

	// Ensure the database exists and is initialised
	err := initDatabase(db)
	if err != nil {
		return err
	}

	return nil
}

// Save saves the database to disk
// Errors are ignored, as save operations are usually done during application
// use and are temporary (or nothing can be done)
func (db *Database) Save() {
	file, err := os.OpenFile(db.path, os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		fmt.Printf("WARNING: failed to write database: %s\n", ErrorDatabaseIOFailed)
	}

	for _, elem := range db.podcasts {
		fmt.Fprintf(file, "%s %s\n", elem.RegexPattern, elem.FriendlyName)
	}

	file.Close()
}

// GetFriendlyName returns the user-configured friendly name for a
// specified URL. If one cannot be found, the url is returned.
func (db *Database) GetFriendlyName(url string) string {
	for _, elem := range db.podcasts {
		matched, _ := regexp.MatchString(elem.RegexPattern, url)
		if matched {
			return elem.FriendlyName
		}
	}

	return url
}

// GetRegex returns the registered regex for a specified friendly name - as
// returned by GetFriendlyName.
func (db *Database) GetRegex(friendly string) string {
	for _, elem := range db.podcasts {
		if elem.FriendlyName == friendly {
			return elem.RegexPattern
		}
	}

	return ""
}
