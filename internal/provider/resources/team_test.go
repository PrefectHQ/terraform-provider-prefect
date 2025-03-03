package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccTeamResource(name, description string) string {
	return fmt.Sprintf(`
resource "prefect_team" "test" {
	name = %q
	description = %q
}
	`, name, description)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team(t *testing.T) {
	randomName := testutils.NewRandomPrefixedString()
	randomName2 := testutils.NewRandomPrefixedString()

	randomDescription := testutils.NewRandomPrefixedString()
	randomDescription2 := testutils.NewRandomPrefixedString()

	resourceName := "prefect_team.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccTeamResource(randomName, randomDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
					testutils.ExpectKnownValue(resourceName, "description", randomDescription),
				},
			},
			{
				Config: fixtureAccTeamResource(randomName2, randomDescription2),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName2),
					testutils.ExpectKnownValue(resourceName, "description", randomDescription2),
				},
			},
			{
				ImportState:       true,
				ResourceName:      resourceName,
				ImportStateVerify: true,
			},
		},
	})
}
