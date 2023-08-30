package api

import (
	"context"
	"github.com/google/uuid"
)

type ServiceAccountsClient interface {
	Create(ctx context.Context, request ServiceAccountCreateRequest) (*ServiceAccount, error)
	List(ctx context.Context, filter ServiceAccountFilterRequest) ([]*ServiceAccount, error)
	Get(ctx context.Context, name string) (*ServiceAccount, error)
	Update(ctx context.Context, name string, data ServiceAccountUpdateRequest) error
	Delete(ctx context.Context, name string) error
}


/*** REQUEST DATA STRUCTS ***/

type ServiceAccountCreateRequest struct {
	Name            string `json:"name"`
	APIKeyExpiration string `json:"api_key_expiration"`
	AccountRoleId   string `json:"account_role_id"`
}

type ServiceAccountUpdateRequest struct {
	Name string `json:"name"`
}

type ServiceAccountFilterRequest struct {
	Any []uuid.UUID `json:"any_"`
}

/*** RESPONSE DATA STRUCTS ***/

// ServiceAccount is a representation of a created service account (from a Create response)
type ServiceAccount struct {
	BaseModel
	AccountId		string 					`json:"account_id"`
	Name             string                 `json:"name"`	
	AccountRoleName string 					`json:"account_role_name"`
	APIKey			ServiceAccountAPIKey 	`json:"api_key"`
}

type ServiceAccountAPIKey struct {
	Id 			string `json:"id"`
	Created 	string `json:"created"`
	Name 		string `json:"name"`
	Expiration 	string `json:"expiration"`
	Key 		string `json:"key"`
}


// ServiceAccountNoKey is a representation of Service Account details obtained from a List/Filter
// and excludes the actual key value for the api_key
type ServiceAccountNoKey struct {
	BaseModel
	AccountId		string 					`json:"account_id"`
	Name             string                 `json:"name"`	
	AccountRoleName string 					`json:"account_role_name"`
	APIKey			ServiceAccountAPIKeyNoKey 	`json:"api_key"`
}

// Represents an api_key block received from a List/Filter response, which excludes the actual key
type ServiceAccountAPIKeyNoKey struct {
	Id 			string `json:"id"`
	Created 	string `json:"created"`
	Name 		string `json:"name"`
	Expiration 	string `json:"expiration"`
}




