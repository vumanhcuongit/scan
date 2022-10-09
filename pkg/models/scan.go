package models

import (
	"time"
)

type Scan struct {
	ID             int64     `json:"id"`
	RepositoryID   int64     `json:"repository_id"`
	RepositoryName string    `json:"repository_name"`
	RepositoryURL  string    `json:"repository_url"`
	Findings       string    `json:"findings"`
	Status         string    `json:"status"`
	QueueAt        time.Time `json:"queued_at"`
	ScanningAt     time.Time `json:"scanning_at"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type ScanFilter struct {
	RepositoryID   *int64
	RepositoryName *string
	Status         *string
}
