package client

import (
	"fmt"
	"net/http"
)

func doRequest(client *http.Client, apiKey string, request *http.Request) (*http.Response, error) {
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("Accept", "application/json")

	if apiKey != "" {
		request.Header.Set("Authorization", "Bearer "+apiKey)
	}

	resp, err := client.Do(request)
	if err != nil {
		return nil, fmt.Errorf("http error: %w", err)
	}

	return resp, nil
}
