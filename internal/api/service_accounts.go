package api

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ServiceAccountsClient interface {
	Create(ctx context.Context, request ServiceAccountCreateRequest) (*ServiceAccount, error)
	List(ctx context.Context, filter ServiceAccountFilterRequest) ([]*ServiceAccount, error)
	Get(ctx context.Context, id string) (*ServiceAccount, error)
	Update(ctx context.Context, id string, data ServiceAccountUpdateRequest) error
	Delete(ctx context.Context, id string) error
	RotateKey(ctx context.Context, id string, data ServiceAccountRotateKeyRequest) (*ServiceAccount, error)
}

/*** REQUEST DATA STRUCTS ***/

type ServiceAccountCreateRequest struct {
	Name             string     `json:"name"`
	APIKeyExpiration string     `json:"api_key_expiration,omitempty"`
	AccountRoleID    *uuid.UUID `json:"account_role_id,omitempty"`
}

type ServiceAccountUpdateRequest struct {
	Name          string     `json:"name"`
	AccountRoleID *uuid.UUID `json:"account_role_id,omitempty"`
}

type ServiceAccountRotateKeyRequest struct {
	APIKeyExpiration *time.Time `json:"api_key_expiration"`
}

type ServiceAccountFilterRequest struct {
	Any []uuid.UUID `json:"any_"`
}

/*** RESPONSE DATA STRUCTS ***/

// ServiceAccount is a representation of a created service account (from a Create response).
type ServiceAccount struct {
	BaseModel
	AccountID       uuid.UUID            `json:"account_id"`
	Name            string               `json:"name"`
	AccountRoleName string               `json:"account_role_name"`
	APIKey          ServiceAccountAPIKey `json:"api_key"`
}

type ServiceAccountAPIKey struct {
	ID         string     `json:"id"`
	Created    *time.Time `json:"created"`
	Expiration *time.Time `json:"expiration"`
	Name       string     `json:"name"`
	Key        string     `json:"key"`
}

// ServiceAccountNoKey is a representation of Service Account details obtained from a List/Filter
// and excludes the actual key value for the api_key.
type ServiceAccountNoKey struct {
	BaseModel
	AccountID       uuid.UUID                 `json:"account_id"`
	Name            string                    `json:"name"`
	AccountRoleName string                    `json:"account_role_name"`
	APIKey          ServiceAccountAPIKeyNoKey `json:"api_key"`
}

// Represents an api_key block received from a List/Filter response, which excludes the actual key.
type ServiceAccountAPIKeyNoKey struct {
	ID         string     `json:"id"`
	Created    *time.Time `json:"created"`
	Name       string     `json:"name"`
	Expiration *time.Time `json:"expiration"`
}
