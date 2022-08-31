package prefect

import (
	"context"
	"fmt"
	"net/http"
	"terraform-provider-prefect/api"

	gqlclient "git.sr.ht/~emersion/gqlclient"
)

type transport struct {
	http.RoundTripper

	header http.Header
}

func (tr *transport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, values := range tr.header {
		for _, v := range values {
			req.Header.Add(k, v)
		}
	}
	return tr.RoundTripper.RoundTrip(req)
}

const DefaultAPIServer string = "https://api.prefect.io"

type Client struct {
	GQLClient *gqlclient.Client
	Tenant    Tenant
}

type Tenant struct {
	Id api.UUID
	// role name => role id map
	RoleIds map[string]api.UUID
}

func NewClient(ctx context.Context, apiKey, apiServer *string) (*Client, error) {
	var endpoint string

	if apiServer == nil {
		endpoint = DefaultAPIServer
	} else {
		endpoint = *apiServer
	}

	tr := transport{
		RoundTripper: http.DefaultTransport,
		header:       make(http.Header),
	}

	tr.header.Add("Authorization", fmt.Sprintf("Bearer %s", *apiKey))

	httpClient := http.Client{Transport: &tr}
	gqlClient := gqlclient.New(endpoint, &httpClient)

	// Fetch tenant metadata, needed for service account api calls

	// Fetch tenant id
	currentUserResponse, err := api.CurrentUser(gqlClient, ctx)
	if err != nil {
		return nil, err
	}

	var tenantId = (api.UUID)(currentUserResponse[0].Default_membership.Tenant.Id)

	// Fetch auth roles
	authRolesResponse, err := api.AuthRoles(gqlClient, ctx)
	if err != nil {
		return nil, err
	}

	authRoles := make(map[string]api.UUID)

	for _, r := range authRolesResponse {
		authRoles[r.Name] = (api.UUID)(r.Id)
	}

	c := &Client{
		GQLClient: gqlClient,
		Tenant: Tenant{
			Id:      tenantId,
			RoleIds: authRoles,
		},
	}

	return c, nil
}
