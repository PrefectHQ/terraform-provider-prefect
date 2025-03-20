package helpers

import (
	"fmt"
	"strings"
)

const (
	PlanPrefectCloudFree       = "Prefect Cloud (Free)"
	PlanPrefectCloudPro        = "Prefect Cloud (Pro)"
	PlanPrefectCloudEnterprise = "Prefect Cloud (Enterprise)"
	PlanPrefectOSS             = "Prefect OSS"

	descriptionTemplate = `
%s

This feature is available in the following [product plans](https://www.prefect.io/pricing): %s.
`
)

var (
	AllPlans = []string{
		PlanPrefectCloudFree,
		PlanPrefectCloudPro,
		PlanPrefectCloudEnterprise,
		PlanPrefectOSS,
	}
)

// DescriptionWithPlans adds a note to the description denoting which plans
// the resource or datasource supports.
func DescriptionWithPlans(description string, plans ...string) string {
	plansFormatted := strings.Join(plans, ", ")

	return fmt.Sprintf(descriptionTemplate, description, plansFormatted)
}
