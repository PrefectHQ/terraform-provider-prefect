package client

import (
	"github.com/google/uuid"

	retryablehttp "github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	hc                 *retryablehttp.Client
	endpoint           string
	apiKey             string
	defaultAccountID   uuid.UUID
	defaultWorkspaceID uuid.UUID
}

type Option func(c *Client) error
