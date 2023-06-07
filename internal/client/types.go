package client

import (
	"net/http"

	"github.com/google/uuid"
)

type Client struct {
	hc                 *http.Client
	endpoint           string
	apiKey             string
	defaultAccountID   uuid.UUID
	defaultWorkspaceID uuid.UUID
}

type Option func(c *Client) error
