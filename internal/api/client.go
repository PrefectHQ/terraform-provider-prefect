package api

import "github.com/google/uuid"

// PrefectClient returns clients for different aspects of our API.
type PrefectClient interface {
	Accounts(accountID uuid.UUID) (AccountsClient, error)
	AccountMemberships(accountID uuid.UUID) (AccountMembershipsClient, error)
	AccountRoles(accountID uuid.UUID) (AccountRolesClient, error)
	Collections() (CollectionsClient, error)
	Deployments(accountID uuid.UUID, workspaceID uuid.UUID) (DeploymentsClient, error)
	Workspaces(accountID uuid.UUID) (WorkspacesClient, error)
	WorkspaceAccess(accountID uuid.UUID, workspaceID uuid.UUID) (WorkspaceAccessClient, error)
	WorkspaceRoles(accountID uuid.UUID) (WorkspaceRolesClient, error)
	WorkPools(accountID uuid.UUID, workspaceID uuid.UUID) (WorkPoolsClient, error)
	Variables(accountID uuid.UUID, workspaceID uuid.UUID) (VariablesClient, error)
	ServiceAccounts(accountID uuid.UUID) (ServiceAccountsClient, error)
}
