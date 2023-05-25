package api

type PrefectClient interface {
	WorkPools() WorkPoolsClient
}
