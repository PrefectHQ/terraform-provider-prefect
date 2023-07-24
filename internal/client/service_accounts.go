package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/armalite/terraform-provider-prefect/internal/api"
)

type ServiceAccountsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}


func (c *Client) ServiceAccounts(accountID uuid.UUID) (api.ServiceAccountsClient, error) {
    if c.apiKey == "" {
        return nil, fmt.Errorf("apiKey is not set")
    }

    if c.endpoint == "" {
        return nil, fmt.Errorf("endpoint is not set")
    }

    if accountID == uuid.Nil {
        return nil, fmt.Errorf("accountID is not set and no default accountID is available")
    }

    routePrefix := getAccountScopedURL(c.endpoint, accountID, "service_accounts")

    return &ServiceAccountsClient{
        hc:          c.hc,
        apiKey:      c.apiKey,
        routePrefix: routePrefix,
    }, nil
}

func (sa *ServiceAccountsClient) CreateServiceAccount(ctx context.Context, request api.CreateServiceAccountRequest) (*api.CreateServiceAccountResponse, error) {
	path := sa.routePrefix + "/"
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response api.CreateServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *ServiceAccountsClient) ReadServiceAccount(ctx context.Context, botId string) (*api.ReadServiceAccountResponse, error) {
	path := sa.routePrefix + "/" + botId

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response api.ReadServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *ServiceAccountsClient) UpdateServiceAccount(ctx context.Context, botId string, request api.UpdateServiceAccountRequest) (*api.UpdateServiceAccountResponse, error) {
	path := sa.routePrefix + "/" + botId
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response api.UpdateServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *ServiceAccountsClient) DeleteServiceAccount(ctx context.Context, botId string) (*api.DeleteServiceAccountResponse, error) {
	path := sa.routePrefix + "/" + botId

	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response api.DeleteServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *ServiceAccountsClient) RotateServiceAccountAPIKey(ctx context.Context, botId string, request api.RotateServiceAccountAPIKeyRequest) (*api.RotateServiceAccountAPIKeyResponse, error) {
	path := sa.routePrefix + "/" + botId + "/rotate_api_key"
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response api.RotateServiceAccountAPIKeyResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}
