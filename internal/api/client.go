package api

// PrefectClient returns clients for different aspects of our API.
type PrefectClient interface {
	WorkPools() WorkPoolsClient
	Variables() VariablesClient
}
