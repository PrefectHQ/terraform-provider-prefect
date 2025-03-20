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

This feature is available in the following [product plan(s)](https://www.prefect.io/pricing): %s.
`
)

var (
	AllPlans = []string{
		PlanPrefectOSS,
		PlanPrefectCloudFree,
		PlanPrefectCloudPro,
		PlanPrefectCloudEnterprise,
	}

	AllCloudPlans = []string{
		PlanPrefectCloudFree,
		PlanPrefectCloudPro,
		PlanPrefectCloudEnterprise,
	}
)

// DescriptionWithPlans adds a note to the description denoting which plans
// the resource or datasource supports.
//
// This function will add a note to the description denoting which plans
// the resource or datasource supports.
//
// Plan support information can be found in a few ways:
// - in the product docs: https://docs.prefect.io
// - in the product plans page: https://www.prefect.io/pricing
// - in the "Account settings" page on the left panel
// - in an instance of the OSS UI, manually finding which features are available
//
// This function is most often used in the `Schema` method of a resource or datasource.
func DescriptionWithPlans(description string, plans ...string) string {
	plansFormatted := strings.Join(plans, ", ")

	return fmt.Sprintf(descriptionTemplate, description, plansFormatted)
}
