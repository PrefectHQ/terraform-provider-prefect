package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// BotsClient is a client for working with Service Accounts
type BotsClient interface {
	Get(ctx context.Context, id uuid.UUID) (*Bot, error)
}

// BotAPIKey represents the nested API Key
// included in a Service Account response
type BotAPIKey struct {
	BaseModel
	Name       string     `json:"name"`
	Key        *string    `json:"key"`
	Expiration *time.Time `json:"expiration"`
}

// Bot is the base representation of a Service Account
type Bot struct {
	BaseModel
	Name            string     `json:"name"`
	AccountID       uuid.UUID  `json:"account_id"`
	AccountRoleName string     `json:"account_role_name"`
	APIKey          *BotAPIKey `json:"api_key"`
}
