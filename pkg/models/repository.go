package models

import (
	"errors"
	"strings"
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

func NewRepository(repositoryURL string) (*Repository, error) {
	splittedURL := strings.Split(repositoryURL, "/")
	if len(splittedURL) != 5 {
		return nil, errors.New("invalid repository")
	}

	owner := splittedURL[3]
	name := splittedURL[4]
	return &Repository{
		RepositoryURL: repositoryURL,
		Owner:         owner,
		Name:          name,
	}, nil
}
