package prefect_api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (c *Client) GetAllWorkQueues(ctx context.Context, workspaceID string) ([]WorkQueue, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accounts/%s/workspaces/%s/work_queues/filter", c.PrefectApiUrl, c.PrefectAccountId, workspaceID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	workQueues := []WorkQueue{}

	err = json.Unmarshal(body, &workQueues)
	if err != nil {
		return nil, err
	}

	return workQueues, nil
}

func (c *Client) GetWorkQueue(ctx context.Context, workQueueId string, workspaceID string) (*WorkQueue, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/accounts/%s/workspaces/%s/work_queues/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, workQueueId), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	workQueue := WorkQueue{}
	err = json.Unmarshal(body, &workQueue)
	if err != nil {
		return nil, err
	}

	return &workQueue, nil
}

func (c *Client) CreateWorkQueue(ctx context.Context, workQueue WorkQueue, workspaceID string) (*WorkQueue, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(workQueue)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accounts/%s/workspaces/%s/work_queues/", c.PrefectApiUrl, c.PrefectAccountId, workspaceID), &buf)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	newWorkQueue := WorkQueue{}
	err = json.Unmarshal(body, &newWorkQueue)
	if err != nil {
		return nil, err
	}

	return &newWorkQueue, nil
}

func (c *Client) UpdateWorkQueue(ctx context.Context, workQueue WorkQueue, workQueueId string, workspaceID string) (*WorkQueue, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(workQueue)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/accounts/%s/workspaces/%s/work_queues/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, workQueueId), &buf)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	} else if len(body) == 0 {
		return nil, nil
	} else {
		updatedWorkQueue := WorkQueue{}
		err = json.Unmarshal(body, &updatedWorkQueue)
		if err != nil {
			return nil, err
		}

		return &updatedWorkQueue, nil
	}
}

func (c *Client) DeleteWorkQueue(ctx context.Context, workQueueId string, workspaceID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/accounts/%s/workspaces/%s/work_queues/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, workQueueId), nil)
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
