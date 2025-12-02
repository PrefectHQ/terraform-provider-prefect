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

func fixtureAccTeamResourceNoDescription(name string) string {
	return fmt.Sprintf(`
resource "prefect_team" "test" {
	name = %q
}
	`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team(t *testing.T) {
	// Teams are not supported in OSS.
	testutils.SkipTestsIfOSS(t)

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

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_no_description(t *testing.T) {
	// Teams are not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	randomName := testutils.NewRandomPrefixedString()

	resourceName := "prefect_team.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Test creating a team without a description field
				Config: fixtureAccTeamResourceNoDescription(randomName),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
					// Description should be empty string (computed), not null
					testutils.ExpectKnownValue(resourceName, "description", ""),
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

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_empty_description(t *testing.T) {
	// Teams are not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	randomName := testutils.NewRandomPrefixedString()

	resourceName := "prefect_team.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Test creating a team with explicit empty string description
				Config: fixtureAccTeamResource(randomName, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
					testutils.ExpectKnownValue(resourceName, "description", ""),
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

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_description_updates(t *testing.T) {
	// Teams are not supported in OSS.
	testutils.SkipTestsIfOSS(t)

	randomName := testutils.NewRandomPrefixedString()
	randomDescription := testutils.NewRandomPrefixedString()
	randomDescription2 := testutils.NewRandomPrefixedString()

	resourceName := "prefect_team.test"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				// Start with no description
				Config: fixtureAccTeamResourceNoDescription(randomName),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
					testutils.ExpectKnownValue(resourceName, "description", ""),
				},
			},
			{
				// Update to add a description
				Config: fixtureAccTeamResource(randomName, randomDescription),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
					testutils.ExpectKnownValue(resourceName, "description", randomDescription),
				},
			},
			{
				// Update to a different description
				Config: fixtureAccTeamResource(randomName, randomDescription2),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
					testutils.ExpectKnownValue(resourceName, "description", randomDescription2),
				},
			},
			{
				// Update to explicit empty string to clear description
				Config: fixtureAccTeamResource(randomName, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(resourceName, "name", randomName),
					testutils.ExpectKnownValue(resourceName, "description", ""),
				},
			},
		},
	})
}
