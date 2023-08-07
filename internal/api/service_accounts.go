package api

import (
	"io/ioutil"
	"net/http"
	"bytes"
	"fmt"
	"log"
	"encoding/json"
	"context"
	"github.com/google/uuid"
)

type ServiceAccountsClient interface {
	Create(ctx context.Context, request CreateServiceAccountRequest) (*ServiceAccount, error)
	List(ctx context.Context, filter WorkPoolFilter) ([]*ServiceAccount, error)
	Get(ctx context.Context, name string) (*ServiceAccount, error)
	Update(ctx context.Context, name string, data WorkPoolUpdate) error
	Delete(ctx context.Context, name string) error
}

// ServiceAccount is a representation of a created service account
type ServiceAccount struct {
	BaseModel
	AccountId		string 					`json:"account_id"`
	Name             string                 `json:"name"`	
	AccountRoleName string 					`json:"account_role_name"`
	APIKey			ServiceAccountAPIKey 	`json:"api_key"`
}

// ServiceAccountFromList is a reprsentation of Service Account details obtained from a List/Filter
type ServiceAccountFromList struct {
	BaseModel
	AccountId		string 					`json:"account_id"`
	Name             string                 `json:"name"`	
	AccountRoleName string 					`json:"account_role_name"`
	APIKey			ServiceAccountAPIKeyFromList 	`json:"api_key"`
}

type ServiceAccountCreate struct {
	Name            string `json:"name"`
	APIKeyExpiration string `json:"api_key_expiration"`
	AccountRoleId   string `json:"account_role_id"`
}

type ServiceAccountUpdate struct {
	Name string `json:"name"`
}

type ServiceAccountAPIKey struct {
	Id 			string `json:"id"`
	Created 	string `json:"created"`
	Name 		string `json:"name"`
	Expiration 	string `json:"expiration"`
	Key 		string `json:"key"`
}

// Represents an api_key block received from a List/Filter response, which excludes the actual key
type ServiceAccountAPIKeyFromList struct {
	Id 			string `json:"id"`
	Created 	string `json:"created"`
	Name 		string `json:"name"`
	Expiration 	string `json:"expiration"`
}

type ServiceAccountFilter struct {
	Any []uuid.UUID `json:"any_"`
}


