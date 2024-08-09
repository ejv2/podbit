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
	// DatabaseDirname is the directory name in which the database will be stored.
	DatabaseDirname = "podbit"

	// DatabaseFilename is the file name of the database on disk.
	DatabaseFilename = "db"
)

// DB-related error values.
var (
	ErrorDatabaseIOFailed = errors.New("Error: IO error while reading from database file")
	ErrorDatabaseSyntax   = "Error: Malformed database: Syntax Error on Line %d"
)

// Podcast is the human-provided podcast info.
type Podcast struct {
	RegexPattern string
	FriendlyName string

	pat *regexp.Regexp
}

// Owns returns true if the given url is a member of this podcast.
func (p *Podcast) Owns(url string) bool {
	if p.pat == nil {
		panic("invalid podcast: regex pattern not compiled")
	}

	return p.pat.MatchString(url)
}

// Database aggregates all podcast data from the database.
type Database struct {
	path           string
	podcasts       []Podcast
	defaultPodcast Podcast
}

func initDatabase(db *Database) error {
	var err error
	file, err := os.Open(db.path)

	if err != nil {
		file, err = os.Create(db.path)
		if err != nil {
			return err
		}

		file.Close()
	} else {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		scanner.Split(bufio.ScanLines)

		i := 1
		for scanner.Scan() {
			var p Podcast

			if scanner.Err() != nil {
				return ErrorDatabaseIOFailed
			}

			elem := scanner.Text()
			if strings.HasPrefix(elem, "#") {
				continue
			}
			if len(elem) == 0 {
				continue
			}

			fields := strings.Split(elem, " ")
			num := len(fields)

			if num < 2 {
				return fmt.Errorf(ErrorDatabaseSyntax, i)
			}

			p.RegexPattern = fields[0]
			p.FriendlyName = strings.Join(fields[1:], " ")

			pat, err := regexp.Compile(p.RegexPattern)
			if err != nil {
				return fmt.Errorf("Error: Malformed database: Invalid regex on Line %d: %w", i, err)
			}
			p.pat = pat

			db.podcasts = append(db.podcasts, p)

			i++
		}
	}

	return nil
}

// Open opens and parses the database.
// Returned errors are usually fatal to the application.
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

	// Add a default podcast, when none else are present
	db.defaultPodcast = Podcast{
		FriendlyName: UnknownPodcastName,
		RegexPattern: ".*",
		pat:          regexp.MustCompile(".*"),
	}
	db.podcasts = append(db.podcasts, db.defaultPodcast)

	return nil
}

// Save saves the database to disk.
// Errors are ignored, as save operations are usually done during application
// use and are temporary (or nothing can be done).
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

// GetPodcasts returns all podcasts configured in the db. This guarantees that
// the same name will never be given twice (i.e podcasts with multi regex will
// be merged.)
func (db *Database) GetPodcasts() []Podcast {
	list := make([]Podcast, 0, len(db.podcasts))
	seen := make(map[string]bool, len(db.podcasts))
	for _, pod := range db.podcasts {
		if seen, ok := seen[pod.FriendlyName]; !(ok || seen) {
			list = append(list, pod)
		}
	}

	return list
}

// GetPodcastNames returns the friendly names of all configured podcasts,
// gnaranteeing the absence of duplicates.
func (db *Database) GetPodcastNames() []string {
	pod := db.GetPodcasts()
	nam := make([]string, 0, len(pod))
	for _, p := range pod {
		nam = append(nam, p.FriendlyName)
	}

	return nam
}

// GetFriendlyName returns the user-configured friendly name for a
// specified URL. If one cannot be found, the url is returned.
func (db *Database) GetFriendlyName(url string) string {
	for _, elem := range db.podcasts {
		if elem.Owns(url) {
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

// GetOwner returns the owning podcast, if known. If the owning podcast is
// known, the default podcast is returned.
func (db *Database) GetOwner(url string) Podcast {
	for _, elem := range db.podcasts {
		if elem.Owns(url) {
			return elem
		}
	}

	return db.defaultPodcast
}
