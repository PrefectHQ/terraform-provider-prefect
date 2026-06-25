package helpers

import "strings"

// cloudEndpointSubstrings are the host substrings that identify a Prefect Cloud
// (or Cloud-equivalent) endpoint. Any endpoint containing one of these is
// treated as Cloud, which enables Cloud-style behavior such as requiring an
// API key and an account ID, and using the account/workspace-scoped API routes.
//
// The "private.prefect.cloud" and "private.prefect.dev" entries cover
// customer-managed Cloud instances served from hosts such as
// "previous-api.private.prefect.cloud" and "latest-api.private.prefect.dev".
// These instances are account/workspace-scoped and use the same API routes as
// Prefect Cloud, so they are treated as Cloud here. Customer-managed feature
// differences (for example, SSO/domain names or metric-trigger automations) are
// handled separately in the acceptance tests via TEST_CONTEXT=CM.
var cloudEndpointSubstrings = []string{
	"api.prefect.cloud",
	"api.prefect.dev",
	"api.stg.prefect.dev",
	"private.prefect.cloud",
	"private.prefect.dev",
}

func IsCloudEndpoint(endpoint string) bool {
	for _, substr := range cloudEndpointSubstrings {
		if strings.Contains(endpoint, substr) {
			return true
		}
	}

	return false
}
