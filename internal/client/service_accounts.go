package client

import (
	"context"
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
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          sa.routePrefix + "/",
		body:         &request,
		apiKey:       sa.apiKey,
		successCodes: successCodesStatusCreated,
	}

	var serviceAccount api.ServiceAccount
	if err := requestWithDecodeResponse(ctx, sa.hc, cfg, &serviceAccount); err != nil {
		return nil, fmt.Errorf("failed to create service account: %w", err)
	}

	return &serviceAccount, nil
}

func (sa *ServiceAccountsClient) List(ctx context.Context, names []string) ([]*api.ServiceAccount, error) {
	filter := api.ServiceAccountFilter{}
	filter.ServiceAccounts.Name.Any = names

	cfg := requestConfig{
		method:       http.MethodPost,
		url:          sa.routePrefix + "/filter",
		body:         &filter,
		apiKey:       sa.apiKey,
		successCodes: successCodesStatusOK,
	}

	var serviceAccounts []*api.ServiceAccount
	if err := requestWithDecodeResponse(ctx, sa.hc, cfg, &serviceAccounts); err != nil {
		return nil, fmt.Errorf("failed to list service accounts: %w", err)
	}

	return serviceAccounts, nil
}

func (sa *ServiceAccountsClient) Get(ctx context.Context, botID string) (*api.ServiceAccount, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          sa.routePrefix + "/" + botID,
		body:         http.NoBody,
		apiKey:       sa.apiKey,
		successCodes: successCodesStatusOK,
	}

	var serviceAccount api.ServiceAccount
	if err := requestWithDecodeResponse(ctx, sa.hc, cfg, &serviceAccount); err != nil {
		return nil, fmt.Errorf("failed to get service account: %w", err)
	}

	return &serviceAccount, nil
}

func (sa *ServiceAccountsClient) Update(ctx context.Context, botID string, requestPayload api.ServiceAccountUpdateRequest) error {
	cfg := requestConfig{
		method:       http.MethodPatch,
		url:          sa.routePrefix + "/" + botID,
		body:         &requestPayload,
		apiKey:       sa.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, sa.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update service account: %w", err)
	}

	defer resp.Body.Close()

	return nil
}

func (sa *ServiceAccountsClient) Delete(ctx context.Context, botID string) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          sa.routePrefix + "/" + botID,
		body:         http.NoBody,
		apiKey:       sa.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, sa.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete service account: %w", err)
	}

	defer resp.Body.Close()

	return nil
}

func (sa *ServiceAccountsClient) RotateKey(ctx context.Context, serviceAccountID string, data api.ServiceAccountRotateKeyRequest) (*api.ServiceAccount, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          fmt.Sprintf("%s/%s/rotate_api_key", sa.routePrefix, serviceAccountID),
		body:         &data,
		apiKey:       sa.apiKey,
		successCodes: successCodesStatusCreated,
	}

	var serviceAccount api.ServiceAccount
	if err := requestWithDecodeResponse(ctx, sa.hc, cfg, &serviceAccount); err != nil {
		return nil, fmt.Errorf("failed to rotate service account key: %w", err)
	}

	return &serviceAccount, nil
}
