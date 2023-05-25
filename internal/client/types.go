package client

import "net/http"

type Client struct {
	hc       *http.Client
	endpoint string
	apiKey   string
}

type Option func(c *Client) error
