package models

import (
	"time"
)

const (
	ScanStatusPending    = "Pending"
	ScanStatusQueued     = "Queued"
	ScanStatusInProgress = "In Progress"
	ScanStatusSuccess    = "Success"
	ScanStatusFailure    = "Failure"
)

type Scan struct {
	ID             int64      `json:"id"`
	RepositoryID   int64      `json:"repository_id"`
	RepositoryName string     `json:"repository_name"`
	RepositoryURL  string     `json:"repository_url"`
	Findings       string     `json:"findings"`
	Status         string     `json:"status"`
	QueuedAt       *time.Time `json:"queued_at"`
	ScanningAt     *time.Time `json:"scanning_at"`
	FinishedAt     *time.Time `json:"finished_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ScanFilter struct {
	RepositoryID   *int64
	RepositoryName *string
	Status         *string
}

func NewScan(repository *Repository) (*Scan, error) {
	return &Scan{
		RepositoryID:   repository.ID,
		RepositoryName: repository.Name,
		RepositoryURL:  repository.RepositoryURL,
		Status:         ScanStatusPending,
	}, nil
}
