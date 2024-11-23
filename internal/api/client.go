package api

import "github.com/google/uuid"

// PrefectClient returns clients for different aspects of our API.
//
//nolint:interfacebloat // we'll accept a larger PrefectClient interface
type PrefectClient interface {
	Accounts(accountID uuid.UUID) (AccountsClient, error)
	Automations(accountID uuid.UUID, workspaceID uuid.UUID) (AutomationsClient, error)
	AccountMemberships(accountID uuid.UUID) (AccountMembershipsClient, error)
	AccountRoles(accountID uuid.UUID) (AccountRolesClient, error)
	BlockDocuments(accountID uuid.UUID, workspaceID uuid.UUID) (BlockDocumentClient, error)
	BlockSchemas(accountID uuid.UUID, workspaceID uuid.UUID) (BlockSchemaClient, error)
	BlockTypes(accountID uuid.UUID, workspaceID uuid.UUID) (BlockTypeClient, error)
	Collections(accountID uuid.UUID, workspaceID uuid.UUID) (CollectionsClient, error)
	Deployments(accountID uuid.UUID, workspaceID uuid.UUID) (DeploymentsClient, error)
	DeploymentAccess(accountID uuid.UUID, workspaceID uuid.UUID) (DeploymentAccessClient, error)
	Teams(accountID uuid.UUID) (TeamsClient, error)
	Flows(accountID uuid.UUID, workspaceID uuid.UUID) (FlowsClient, error)
	Workspaces(accountID uuid.UUID) (WorkspacesClient, error)
	WorkspaceAccess(accountID uuid.UUID, workspaceID uuid.UUID) (WorkspaceAccessClient, error)
	WorkspaceRoles(accountID uuid.UUID) (WorkspaceRolesClient, error)
	WorkPools(accountID uuid.UUID, workspaceID uuid.UUID) (WorkPoolsClient, error)
	Variables(accountID uuid.UUID, workspaceID uuid.UUID) (VariablesClient, error)
	ServiceAccounts(accountID uuid.UUID) (ServiceAccountsClient, error)
}
