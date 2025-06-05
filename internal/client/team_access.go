package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/prefecthq/terraform-provider-prefect/internal/api"
)

var _ = api.TeamAccessClient(&TeamAccessClient{})

// TeamAccessClient is a client for the TeamAccess resource.
type TeamAccessClient struct {
	hc              *http.Client
	apiKey          string
	basicAuthKey    string
	routePrefix     string
	csrfClientToken string
	csrfToken       string
}

// TeamAccess is a factory that initializes and returns a TeamAccessClient.
//
//nolint:ireturn // required to support PrefectClient mocking
func (c *Client) TeamAccess(accountID uuid.UUID, teamID uuid.UUID) (api.TeamAccessClient, error) {
	if accountID == uuid.Nil {
		accountID = c.defaultAccountID
	}

	return &TeamAccessClient{
		hc:              c.hc,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		routePrefix:     fmt.Sprintf("%s/accounts/%s/teams/%s", c.endpoint, accountID.String(), teamID.String()),
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
	}, nil
}

// Upsert creates or updates access to a team for a member.
func (c *TeamAccessClient) Upsert(ctx context.Context, memberType string, memberID uuid.UUID) error {
	payload := api.TeamAccessUpsert{
		Members: []api.TeamAccessMember{
			{
				MemberID:   memberID,
				MemberType: memberType,
			},
		},
	}

	cfg := requestConfig{
		method:          http.MethodPut,
		url:             fmt.Sprintf("%s/members", c.routePrefix),
		body:            &payload,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to upsert team access: %w", err)
	}
	defer resp.Body.Close()

	return nil
}

// Read fetches a team access by member actor ID.
func (c *TeamAccessClient) Read(ctx context.Context, teamID, memberID, memberActorID uuid.UUID) (*api.TeamAccess, error) {
	cfg := requestConfig{
		method:          http.MethodGet,
		url:             c.routePrefix,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusOK,
	}

	var teamAccessRead api.TeamAccessRead
	if err := requestWithDecodeResponse(ctx, c.hc, cfg, &teamAccessRead); err != nil {
		return nil, fmt.Errorf("failed to get team access: %w", err)
	}

	// Find the memberships entry that matches the member ID.
	var teamAccess *api.TeamAccess
	for _, membership := range teamAccessRead.Memberships {
		if membership.ActorID == memberActorID {
			teamAccess = &api.TeamAccess{
				TeamID:        teamID,
				MemberID:      memberID,
				MemberActorID: memberActorID,
				MemberType:    membership.Type,
			}

			break
		}
	}

	if teamAccess == nil {
		return nil, fmt.Errorf("client.Read: team access not found for member ID: %s", memberActorID.String())
	}

	return teamAccess, nil
}

// Delete deletes a team access by member ID.
func (c *TeamAccessClient) Delete(ctx context.Context, memberID uuid.UUID) error {
	cfg := requestConfig{
		method:          http.MethodDelete,
		url:             fmt.Sprintf("%s/members/%s", c.routePrefix, memberID.String()),
		body:            http.NoBody,
		apiKey:          c.apiKey,
		basicAuthKey:    c.basicAuthKey,
		csrfClientToken: c.csrfClientToken,
		csrfToken:       c.csrfToken,
		successCodes:    successCodesStatusNoContent,
	}

	resp, err := request(ctx, c.hc, cfg)
	if err != nil {
		return fmt.Errorf("failed to delete team access: %w", err)
	}
	defer resp.Body.Close()

	return nil
}
