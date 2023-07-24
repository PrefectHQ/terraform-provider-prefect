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


type CreateServiceAccountRequest struct {
	Name            string `json:"name"`
	APIKeyExpiration string `json:"api_key_expiration"`
	AccountRoleId   string `json:"account_role_id"`
}

type UpdateServiceAccountRequest struct {
	Name string `json:"name"`
}

type RotateServiceAccountAPIKeyRequest struct {
	APIKeyExpiration string `json:"api_key_expiration"`
}

type ServiceAccountAPIKey struct {
	Id 			string `json:"id"`
	Created 	string `json:"created"`
	Name 		string `json:"name"`
	Expiration 	string `json:"expiration"`
	Key 		string `json:"key"`
}

type CreateServiceAccountResponse struct {
	Id 			string `json:"id"`
	Created		string `json:"created"`
	Updated		string `json:"updated"`
	AccountId	string `json:"account_id"`
	Name 		string `json:"name"`
	AccountRoleName string `json:"account_role_name"`
	APIKey		ServiceAccountAPIKey `json:"api_key"`
}

type ReadServiceAccountResponse struct {
	Id 				string `json:"id"`
	ActorId			string `json:"actor_id"`
	Created			string `json:"created"`
	Updated			string `json:"updated"`
	AccountId		string `json:"account_id"`
	Name 			string `json:"name"`
	AccountRoleName string `json:"account_role_name"`
	APIKey			ServiceAccountAPIKey `json:"api_key"`
}

type DeleteServiceAccountResponse struct {
	Detail 	string `json:"detail"`
}

type UpdateServiceAccountResponse struct {
	Detail 	string `json:"detail"`
}

type RotateServiceAccountAPIKeyResponse struct {
	Id 			string `json:"id"`
	Created		string `json:"created"`
	Updated		string `json:"updated"`
	AccountId	string `json:"account_id"`
	Name 		string `json:"name"`
	AccountRoleName string `json:"account_role_name"`
	APIKey		ServiceAccountAPIKey `json:"api_key"`
}
