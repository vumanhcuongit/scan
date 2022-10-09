package models

import (
	"time"
)

const (
	RegistrationStatusWaitingForReview = "waiting_for_review"
	RegistrationStatusApproved         = "approved"
	RegistrationStatusRejected         = "rejected"
)

type Registration struct {
	ID                     int64     `json:"id"`
	AppIdentifier          string    `json:"app_identifier"`
	WebhookEndpoint        string    `json:"webhook_endpoint"`
	SandboxWebhookEndpoint string    `json:"sandbox_webhook_endpoint"`
	Status                 string    `json:"status"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}

type RegistrationFilter struct {
	AppIdentifier *string `json:"status"`
}
