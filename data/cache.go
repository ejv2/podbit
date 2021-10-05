package data

type Cache struct {
	dir string

	episodes []Episode
}

type Episode struct {
	path string
}
