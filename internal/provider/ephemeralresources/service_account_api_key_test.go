package ephemeralresources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccEphemeral_service_account_api_key(t *testing.T) {
	workspace := testutils.NewEphemeralWorkspace()
	echoResourceName := "echo.test"

	providerFactories := testutils.TestAccProtoV6ProviderFactories
	providerFactories["echo"] = echoprovider.NewProviderServer()

	resource.ParallelTest(t, resource.TestCase{
		// Ephemeral resources are only available in 1.10 and later
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		ProtoV6ProviderFactories: providerFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccServiceAccountDataSource(workspace.Resource),
				ConfigStateChecks: []statecheck.StateCheck{
					testutils.ExpectKnownValue(echoResourceName, "data", "foo"),
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

	service_account_id = prefect_service_account.test.id
}

provider "echo" {
	data = ephemeral.prefect_service_account_api_key.test.value
}

resource "echo" "test" {}
	`, workspace)
}
