package data

// Human-provided podcast info
type Podcast struct {
	friendlyName string
}

type DB struct {
	podcasts map[string] Podcast
}

func (db *DB) Open() error {
	return nil
}
