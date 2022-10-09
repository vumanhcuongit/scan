package models

import (
	"time"
)

type Repository struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	Owner         string    `json:"owner"`
	RepositoryURL string    `json:"repository_url"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RepositoryFilter struct {
	RepositoryID *int64
}
