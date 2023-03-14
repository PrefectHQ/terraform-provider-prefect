package prefect_api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (c *Client) GetAllWorkspaces(ctx context.Context) ([]Workspace, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accounts/%s/workspaces/filter", c.PrefectApiUrl, c.PrefectAccountId), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	workspaces := []Workspace{}

	err = json.Unmarshal(body, &workspaces)
	if err != nil {
		return nil, err
	}

	return workspaces, nil
}

func (c *Client) GetWorkspace(ctx context.Context, workspaceID string) (*Workspace, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/accounts/%s/workspaces/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	workspace := Workspace{}
	err = json.Unmarshal(body, &workspace)
	if err != nil {
		return nil, err
	}

	return &workspace, nil
}

func (c *Client) CreateWorkspace(ctx context.Context, workspace Workspace) (*Workspace, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(workspace)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accounts/%s/workspaces/", c.PrefectApiUrl, c.PrefectAccountId), &buf)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	newWorkspace := Workspace{}
	err = json.Unmarshal(body, &newWorkspace)
	if err != nil {
		return nil, err
	}

	return &newWorkspace, nil
}

func (c *Client) DeleteWorkspace(ctx context.Context, workspaceID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/accounts/%s/workspaces/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID), nil)

	if err != nil {
		return err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return err
	}

	if string(body) != "" {
		return errors.New(string(body))
	}

	return nil
}
