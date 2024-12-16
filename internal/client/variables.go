package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.VariablesClient(&VariablesClient{})

// VariablesClient is a client for working with variables.
type VariablesClient struct {
	hc          *http.Client
	routePrefix string
	apiKey      string
}

// Variables returns a VariablesClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) Variables(accountID uuid.UUID, workspaceID uuid.UUID) (api.VariablesClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	if workspaceID == uuid.Nil {
		workspaceID = c.defaultWorkspaceID
	}

	if err := validateCloudEndpoint(c.endpoint, accountID, workspaceID); err != nil {
		return nil, err
	}

	return &VariablesClient{
		hc:          c.hc,
		apiKey:      c.apiKey,
		routePrefix: getWorkspaceScopedURL(c.endpoint, accountID, workspaceID, "variables"),
	}, nil
}

// Create returns details for a new variable.
func (c *VariablesClient) Create(ctx context.Context, data api.VariableCreate) (*api.Variable, error) {
	cfg := requestConfig{
		method:       http.MethodPost,
		url:          c.routePrefix + "/",
		body:         &data,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusCreated,
	}

	var variable api.Variable
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &variable); err != nil {
		return nil, fmt.Errorf("failed to create variable: %w", err)
	}

	return &variable, nil
}

// List returns a list of variables matching filter criteria.
func (c *VariablesClient) List(ctx context.Context, filter api.VariableFilter) ([]api.Variable, error) {
	_ = ctx
	_ = filter

	return nil, nil
}

// Get returns details for a variable by ID.
func (c *VariablesClient) Get(ctx context.Context, variableID uuid.UUID) (*api.Variable, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          c.routePrefix + "/" + variableID.String(),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var variable api.Variable
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &variable); err != nil {
		return nil, fmt.Errorf("failed to get variable: %w", err)
	}

	return &variable, nil
}

// GetByName returns details for a variable by name.
func (c *VariablesClient) GetByName(ctx context.Context, name string) (*api.Variable, error) {
	cfg := requestConfig{
		method:       http.MethodGet,
		url:          c.routePrefix + "/name/" + name,
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOK,
	}

	var variable api.Variable
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &variable); err != nil {
		return nil, fmt.Errorf("failed to get variable by name: %w", err)
	}

	return &variable, nil
}

// Update modifies an existing variable by ID.
func (c *VariablesClient) Update(ctx context.Context, variableID uuid.UUID, data api.VariableUpdate) error {
	cfg := requestConfig{
		method:       http.MethodPatch,
		url:          c.routePrefix + "/" + variableID.String(),
		body:         &data,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to update variable: %w", err)
	}

	defer resp.Body.Close()

	return nil
}

// Delete removes a variable by ID.
func (c *VariablesClient) Delete(ctx context.Context, variableID uuid.UUID) error {
	cfg := requestConfig{
		method:       http.MethodDelete,
		url:          c.routePrefix + "/" + variableID.String(),
		body:         http.NoBody,
		apiKey:       c.apiKey,
		successCodes: successCodesStatusOKOrNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete variable: %w", err)
	}

	defer resp.Body.Close()

	return nil
}
