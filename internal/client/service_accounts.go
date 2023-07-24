package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	// @TODO: Update import path to parent repo
	"github.com/armalite/terraform-provider-prefect/internal/api"
)

type ServiceAccountsClient struct {
	hc          *http.Client
	apiKey      string
	routePrefix string
}

func getServiceAccountScopedURL(endpoint string, accountID uuid.UUID, resource string) string {
    return fmt.Sprintf("%s/accounts/%s/%s", endpoint, accountID, resource)
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

    routePrefix := getServiceAccountScopedURL(c.endpoint, accountID, "service_accounts")

    return &ServiceAccountsClient{
        hc:          c.hc,
        routePrefix: routePrefix,
    }, nil
}



func (sa *serviceAccounts) CreateServiceAccount(ctx context.Context, accountId string, request CreateServiceAccountRequest) (*CreateServiceAccountResponse, error) {
	path := sa.client.BaseURL + "/accounts/" + accountId + "/bots/"
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sa.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response CreateServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *serviceAccounts) ReadServiceAccount(ctx context.Context, accountId string, botId string) (*ReadServiceAccountResponse, error) {
	path := sa.client.BaseURL + "/accounts/" + accountId + "/bots/" + botId

	req, err := http.NewRequest("GET", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sa.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response ReadServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *serviceAccounts) UpdateServiceAccount(ctx context.Context, accountId string, botId string, request UpdateServiceAccountRequest) (*UpdateServiceAccountResponse, error) {
	path := sa.client.BaseURL + "/accounts/" + accountId + "/bots/" + botId
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("PATCH", path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sa.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response UpdateServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *serviceAccounts) DeleteServiceAccount(ctx context.Context, accountId string, botId string) (*DeleteServiceAccountResponse, error) {
	path := sa.client.BaseURL + "/accounts/" + accountId + "/bots/" + botId

	req, err := http.NewRequest("DELETE", path, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sa.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response DeleteServiceAccountResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}

func (sa *serviceAccounts) RotateServiceAccountAPIKey(ctx context.Context, accountId string, botId string, request RotateServiceAccountAPIKeyRequest) (*RotateServiceAccountAPIKeyResponse, error) {
	path := sa.client.BaseURL + "/accounts/" + accountId + "/bots/" + botId + "/rotate_api_key"
	body, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", path, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := sa.client.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var response RotateServiceAccountAPIKeyResponse
	json.NewDecoder(resp.Body).Decode(&response)
	return &response, nil
}
