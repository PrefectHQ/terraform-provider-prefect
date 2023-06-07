package api

import "github.com/google/uuid"

// PrefectClient returns clients for different aspects of our API.
type PrefectClient interface {
	Accounts() (AccountsClient, error)
	WorkPools(accountID uuid.UUID, workspaceID uuid.UUID) (WorkPoolsClient, error)
	Variables(accountID uuid.UUID, workspaceID uuid.UUID) (VariablesClient, error)
}
