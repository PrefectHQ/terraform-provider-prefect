package client

// GetEndpointHost returns the endpoint host,
// which is the API domain without the trailing subpath.
// eg. https://api.prefect.cloud
func (c *Client) GetEndpointHost() string {
	return c.endpointHost
}
