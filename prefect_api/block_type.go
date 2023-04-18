package prefect_api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func (c *Client) GetAllBlockTypes(ctx context.Context, workspaceID string) ([]BlockType, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_types/filter", c.PrefectApiUrl, c.PrefectAccountId, workspaceID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	blockTypes := []BlockType{}

	err = json.Unmarshal(body, &blockTypes)
	if err != nil {
		return nil, err
	}

	return blockTypes, nil
}

func (c *Client) GetBlockTypeById(ctx context.Context, blockTypeId string, workspaceID string) (*BlockType, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_types/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, blockTypeId), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	blockType := BlockType{}
	err = json.Unmarshal(body, &blockType)
	if err != nil {
		return nil, err
	}

	return &blockType, nil
}

func (c *Client) GetBlockTypeBySlug(ctx context.Context, slug string, workspaceID string) (*BlockType, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_types/slug/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, slug), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	blockType := BlockType{}
	err = json.Unmarshal(body, &blockType)
	if err != nil {
		return nil, err
	}

	return &blockType, nil
}
