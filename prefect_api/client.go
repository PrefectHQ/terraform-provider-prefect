package prefect_api

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// HostURL - Default Hashicups URL
const HostURL string = "http://localhost:19090"

// Client -
type Client struct {
	HTTPClient       *http.Client
	PrefectApiUrl    string
	PrefectApiKey    string
	PrefectAccountId string
}

// NewClient -
func NewClient(prefect_api_url string, prefect_api_key string, prefect_account_id string) (*Client, error) {
	c := Client{
		HTTPClient:       &http.Client{Timeout: 10 * time.Second},
		PrefectApiUrl:    prefect_api_url,
		PrefectApiKey:    prefect_api_key,
		PrefectAccountId: prefect_account_id,
	}

	return &c, nil
}

func (c *Client) doRequest(req *http.Request, PrefectApiKey string) ([]byte, error) {
	token := c.PrefectApiKey

	req.Header.Set("Authorization", "Bearer "+token)

	res, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK && res.StatusCode != http.StatusCreated && res.StatusCode != http.StatusNoContent {
		return nil, fmt.Errorf("status: %d, body: %s", res.StatusCode, body)
	}

	return body, err
}
