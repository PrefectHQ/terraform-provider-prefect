package api

import (
	"io/ioutil"
	"net/http"
	"bytes"
	"fmt"
	"log"
	"encoding/json"
	"context"
)

type ServiceAccountsClient interface {
	Create(ctx context.Context, request CreateServiceAccountRequest) (*ServiceAccount, error)
	List(ctx context.Context, filter WorkPoolFilter) ([]*ServiceAccount, error)
	Get(ctx context.Context, name string) (*ServiceAccount, error)
	Update(ctx context.Context, name string, data WorkPoolUpdate) error
	Delete(ctx context.Context, name string) error
}

// ServiceAccount is a representation of a service account.
type ServiceAccount struct {
	BaseModel
	AccountId		string 					`json:"account_id"`
	Name             string                 `json:"name"`	
	AccountRoleName string 					`json:"account_role_name"`
	APIKey			ServiceAccountAPIKey 	`json:"api_key"`

}

type CreateServiceAccountRequest struct {
	Name            string `json:"name"`
	APIKeyExpiration string `json:"api_key_expiration"`
	AccountRoleId   string `json:"account_role_id"`
}

type UpdateServiceAccountRequest struct {
	Name string `json:"name"`
}

type ServiceAccountAPIKey struct {
	Id 			string `json:"id"`
	Created 	string `json:"created"`
	Name 		string `json:"name"`
	Expiration 	string `json:"expiration"`
	Key 		string `json:"key"`
}


