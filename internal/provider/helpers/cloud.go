package helpers

import "strings"

// cloudEndpointSubstrings are the host substrings that identify a Prefect Cloud
// (or Cloud-equivalent) endpoint. Any endpoint containing one of these is
// treated as Cloud, which enables Cloud-specific behavior such as requiring an
// API key and an account ID.
//
// The "private.prefect.cloud" entry covers customer-managed Cloud instances
// served from hosts such as "previous-api.private.prefect.cloud" and
// "next-api.private.prefect.cloud".
var cloudEndpointSubstrings = []string{
	"api.prefect.cloud",
	"api.prefect.dev",
	"api.stg.prefect.dev",
	"private.prefect.cloud",
}

func IsCloudEndpoint(endpoint string) bool {
	for _, substr := range cloudEndpointSubstrings {
		if strings.Contains(endpoint, substr) {
			return true
		}
	}

	return false
}
