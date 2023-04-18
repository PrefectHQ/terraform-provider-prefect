package prefect_api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

func (c *Client) GetAllBlockDocuments(ctx context.Context, workspaceID string) ([]BlockDocument, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_documents/filter", c.PrefectApiUrl, c.PrefectAccountId, workspaceID), nil)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	blockDocuments := []BlockDocument{}

	err = json.Unmarshal(body, &blockDocuments)
	if err != nil {
		return nil, err
	}

	return blockDocuments, nil
}

func (c *Client) GetBlockDocument(ctx context.Context, blockDocumentId string, workspaceID string) (*BlockDocument, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_documents/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, blockDocumentId), nil)

	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	blockDocument := BlockDocument{}
	err = json.Unmarshal(body, &blockDocument)
	if err != nil {
		return nil, err
	}

	return &blockDocument, nil
}

func (c *Client) CreateBlockDocument(ctx context.Context, blockDocument BlockDocument, workspaceID string) (*BlockDocument, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(blockDocument)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_documents/", c.PrefectApiUrl, c.PrefectAccountId, workspaceID), &buf)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	}

	newBlockDocument := BlockDocument{}
	err = json.Unmarshal(body, &newBlockDocument)
	if err != nil {
		return nil, err
	}

	return &newBlockDocument, nil
}

func (c *Client) UpdateBlockDocument(ctx context.Context, blockDocument BlockDocument, blockDocumentId string, workspaceID string) (*BlockDocument, error) {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(blockDocument)
	if err != nil {
		panic(err)
	}

	req, err := http.NewRequestWithContext(ctx, "PATCH", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_documents/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, blockDocumentId), &buf)
	if err != nil {
		return nil, err
	}

	body, err := c.doRequest(req, c.PrefectApiKey)
	if err != nil {
		return nil, err
	} else if len(body) == 0 {
		return nil, nil
	} else {
		updatedBlockDocument := BlockDocument{}
		err = json.Unmarshal(body, &updatedBlockDocument)
		if err != nil {
			return nil, err
		}

		return &updatedBlockDocument, nil
	}
}

func (c *Client) DeleteBlockDocument(ctx context.Context, blockDocumentId string, workspaceID string) error {
	req, err := http.NewRequestWithContext(ctx, "DELETE", fmt.Sprintf("%s/accounts/%s/workspaces/%s/block_documents/%s", c.PrefectApiUrl, c.PrefectAccountId, workspaceID, blockDocumentId), nil)
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
