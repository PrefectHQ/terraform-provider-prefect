package api

import "github.com/google/uuid"

// PrefectClient returns clients for different aspects of our API.
type PrefectClient interface {
	Accounts() (AccountsClient, error)
	AccountRoles(accountID uuid.UUID) (AccountRolesClient, error)
	Workspaces(accountID uuid.UUID) (WorkspacesClient, error)
	WorkspaceRoles(accountID uuid.UUID) (WorkspaceRolesClient, error)
	WorkPools(accountID uuid.UUID, workspaceID uuid.UUID) (WorkPoolsClient, error)
	Variables(accountID uuid.UUID, workspaceID uuid.UUID) (VariablesClient, error)
	ServiceAccounts(accountID uuid.UUID) (ServiceAccountsClient, error)
}
