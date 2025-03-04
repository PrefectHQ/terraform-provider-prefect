package resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

// Test service accounts.

func fixtureAccTeamAccessResourceForServiceAccount(name string) string {
	return fmt.Sprintf(`
resource "prefect_service_account" "test" {
	name = "%s"
}

resource "prefect_team" "test" {
	name = "%s"
	description = "test-team-description"
}

resource "prefect_team_access" "test" {
	member_type = "service_account"
	member_id = prefect_service_account.test.id
	member_actor_id = prefect_service_account.test.actor_id
	team_id = prefect_team.test.id
}
`, name, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_access_service_account(t *testing.T) {
	accessResourceName := "prefect_team_access.test"
	randomName := testutils.NewRandomPrefixedString()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccTeamAccessResourceForServiceAccount(randomName),
				Check:  resource.ComposeAggregateTestCheckFunc(),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(accessResourceName, "member_type", "service_account"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "member_id"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "member_actor_id"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "team_id"),
				},
			},
		},
	})
}

// Test user accounts.

func fixtureAccTeamAccessResourceForUser(name string) string {
	return fmt.Sprintf(`
data "prefect_account_member" "test" {
	email = "marvin@prefect.io"
}

resource "prefect_team" "test" {
	name = "%s"
	description = "test-team-description"
}

resource "prefect_team_access" "test" {
	team_id = prefect_team.test.id
	member_type = "user"
	member_id = data.prefect_account_member.test.user_id
	member_actor_id = data.prefect_account_member.test.actor_id
}
`, name)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccResource_team_access_user(t *testing.T) {
	accessResourceName := "prefect_team_access.test"
	randomName := testutils.NewRandomPrefixedString()

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccTeamAccessResourceForUser(randomName),
				Check:  resource.ComposeAggregateTestCheckFunc(),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(accessResourceName, "member_type", "user"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "member_id"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "member_actor_id"),
					testutils.ExpectKnownValueNotNull(accessResourceName, "team_id"),
				},
			},
		},
	})
}
