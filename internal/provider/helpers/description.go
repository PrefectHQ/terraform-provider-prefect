package helpers

import (
	"fmt"
	"strings"
)

const (
	TierPrefectCloudFree = "Prefect Cloud (free)"
	TierPrefectCloudPaid = "Prefect Cloud (paid)"
	TierPrefectOSS       = "Prefect OSS"

	descriptionTemplate = `
%s

This feature is available in the following product tiers: %s.
`
)

var (
	AllTiers = []string{
		TierPrefectCloudFree,
		TierPrefectCloudPaid,
		TierPrefectOSS,
	}
)

// DescriptionWithTiers adds a note to the description denoting which tiers
// the resource or datasource supports.
func DescriptionWithTiers(description string, tiers []string) string {
	tiersFormatted := strings.Join(tiers, ", ")

	return fmt.Sprintf(descriptionTemplate, description, tiersFormatted)
}
