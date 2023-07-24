import (
	"io/ioutil"
	"net/http"
	"bytes"
    "fmt"
	"log"
	"encoding/json"
	"context"
)

type ServiceAccountAPIKey struct {
	Id 			string
	Created 	string
	Name 		string
	Expiration 	string
	key 		string
}

type CreateServiceAccountResponse struct {
	Id 			string
	Created		string
	Updated		string
	AccountId	string
	Name 		string
	AccountRoleName string
	APIKey		ServiceAccountAPIKey
}

type ReadServiceAccountResponse struct {
	Id 				string
	ActorId			string
	Created			string
	Updated			string
	AccountId		string
	Name 			string
	AccountRoleName string
	APIKey			ServiceAccountAPIKey
}

type DeleteServiceAccountResponse struct {
	Detail 	string
}

type UpdateServiceAccountResponse struct {
	Detail 	string
}

type RotateServiceAccountAPIKeyResponse struct {
	Id 			string
	Created		string
	Updated		string
	AccountId	string
	Name 		string
	AccountRoleName string
	APIKey		ServiceAccountAPIKey
}