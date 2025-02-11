package ephemeral_resources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccEphemeral_service_account_api_key(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	ephemeralResourceName := "ephemeral.prefect_service_account_api_key.test"

	resource.ParallelTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccServiceAccountDataSource(workspace.Resource),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(ephemeralResourceName, "value", "foo"),
				},
			},
		},
	})
}

func fixtureAccServiceAccountDataSource(workspace string) string {
	return fmt.Sprintf(`
%s

resource "prefect_service_account" "test" {
	name = "test"
}

ephemeral "prefect_service_account_api_key" "test" {
	depends_on = [prefect_service_account.test]

	account_id = "9a67b081-4f14-4035-b000-1f715f46231b"
	service_account_id = prefect_service_account.test.id
}
	`, workspace)
}
