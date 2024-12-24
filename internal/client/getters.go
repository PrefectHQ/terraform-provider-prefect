package client

// GetEndpointHost returns the endpoint host.
// eg. https://api.prefect.cloud
func (c *Client) GetEndpointHost() string {
	return c.endpointHost
}
