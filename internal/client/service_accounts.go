package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	retryablehttp "github.com/hashicorp/go-retryablehttp"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

type ServiceAccountsClient struct {
	hc          *retryablehttp.Client
	apiKey      string
	routePrefix string
}

//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) ServiceAccounts(accountID uuid.UUID) (api.ServiceAccountsClient, error) {
	if c.apiKey == "" {
		return nil, fmt.Errorf("apiKey is not set")
	}

	if c.endpoint == "" {
		return nil, fmt.Errorf("endpoint is not set")
	}

	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
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

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, sa.routePrefix+"/", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var response api.ServiceAccount
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (sa *ServiceAccountsClient) List(ctx context.Context, names []string) ([]*api.ServiceAccount, error) {
	filter := api.ServiceAccountFilter{}
	filter.ServiceAccounts.Name.Any = names

	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&filter); err != nil {
		return nil, fmt.Errorf("failed to encode filter: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, sa.routePrefix+"/filter", &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var serviceAccounts []*api.ServiceAccount
	if err := json.NewDecoder(resp.Body).Decode(&serviceAccounts); err != nil { // THIS IS THE RESPONSE
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return serviceAccounts, nil
}

func (sa *ServiceAccountsClient) Get(ctx context.Context, botID string) (*api.ServiceAccount, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodGet, sa.routePrefix+"/"+botID, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusNotFound:
		return nil, fmt.Errorf("could not find Service Account")
	default:
		bodyBytes, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var response api.ServiceAccount
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

func (sa *ServiceAccountsClient) Update(ctx context.Context, botID string, request api.ServiceAccountUpdateRequest) error {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&request); err != nil {
		return fmt.Errorf("failed to encode request data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPatch, sa.routePrefix+"/"+botID, &buf)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}

	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

func (sa *ServiceAccountsClient) Delete(ctx context.Context, botID string) error {
	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodDelete, sa.routePrefix+"/"+botID, http.NoBody)
	if err != nil {
		return fmt.Errorf("error creating request: %w", err)
	}
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		errorBody, _ := io.ReadAll(resp.Body)

		return fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	return nil
}

func (sa *ServiceAccountsClient) RotateKey(ctx context.Context, serviceAccountID string, data api.ServiceAccountRotateKeyRequest) (*api.ServiceAccount, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(&data); err != nil {
		return nil, fmt.Errorf("failed to encode request data: %w", err)
	}

	req, err := retryablehttp.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/%s/rotate_api_key", sa.routePrefix, serviceAccountID), &buf)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}
	setDefaultHeaders(req, sa.apiKey)

	resp, err := sa.hc.Do(req)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		errorBody, _ := io.ReadAll(resp.Body)

		return nil, fmt.Errorf("status code %s, error=%s", resp.Status, errorBody)
	}

	var serviceAccount api.ServiceAccount
	if err := json.NewDecoder(resp.Body).Decode(&serviceAccount); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &serviceAccount, nil
}
