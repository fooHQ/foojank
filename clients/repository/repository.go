package repository

import "time"

type Repository struct {
	Name        string
	Description string
	Size        uint64
}

type File struct {
	Name     string
	Size     uint64
	Modified time.Time
}
