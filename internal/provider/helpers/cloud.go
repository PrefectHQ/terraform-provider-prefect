package helpers

import "strings"

func IsCloudEndpoint(endpoint string) bool {
	return strings.Contains(endpoint, "api.prefect.cloud") || strings.Contains(endpoint, "api.prefect.dev") || strings.Contains(endpoint, "api.stg.prefect.dev")
}
