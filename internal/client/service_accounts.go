package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
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

	// Since service accounts are account scoped. Generate from util.getAccountScopedURL
	// e.g. this will generate routePrefix ending in /accounts/bots
    routePrefix := getAccountScopedURL(c.endpoint, accountID, "bots")

    return &ServiceAccountsClient{
        hc:          c.hc,
        apiKey:      c.apiKey,
        routePrefix: routePrefix,
    }, nil
}


func (sa *ServiceAccountsClient) Create(ctx context.Context, request api.ServiceAccountCreateRequest) (*api.ServiceAccount, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&request); err != nil {
		return nil, fmt.Errorf("failed to encode request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, sa.routePrefix+"/", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var response api.ServiceAccount
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &response, nil
}


func (c *ServiceAccountsClient) List(ctx context.Context, filter api.ServiceAccountFilterRequest) ([]*api.ServiceAccount, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&filter); err != nil {
		return nil, fmt.Errorf("failed to encode filter: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.routePrefix+"/filter", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, c.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var serviceAccounts []*api.ServiceAccountNoKey
	if err := json.NewDecoder(resp.Body).Decode(&serviceAccounts); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return pools, nil
}


func (sa *ServiceAccountsClient) Get(ctx context.Context, botId string) (*api.ServiceAccount, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.routePrefix+"/"+botId, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return nil, fmt.Errorf("status code %s", resp.Status)
	}

	var response api.ServiceAccountNoKey
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &response, nil
}


func (sa *ServiceAccountsClient) Update(ctx context.Context, botId string, request api.ServiceAccountUpdateRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&request); err != nil {
		return fmt.Errorf("failed to encode request data: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.routePrefix+"/"+botId, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, sa.apiKey)

	resp, err := c.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		return fmt.Errorf("status code %s", resp.Status)
	}
	return nil
}


func (sa *ServiceAccountsClient) Delete(ctx context.Context, botId string) error {
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
	return nil
}

