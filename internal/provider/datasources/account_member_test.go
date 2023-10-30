package datasources_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/testutils"
)

func fixtureAccAccountMember(email string) string {
	return fmt.Sprintf(`
data "prefect_account_member" "member" {
	email = "%s"
}
	`, email)
}

//nolint:paralleltest // we use the resource.ParallelTest helper instead
func TestAccDatasource_account_member(t *testing.T) {
	dataSourceName := "data.prefect_account_member.member"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV6ProviderFactories: testutils.TestAccProtoV6ProviderFactories,
		PreCheck:                 func() { testutils.AccTestPreCheck(t) },
		Steps: []resource.TestStep{
			{
				Config: fixtureAccAccountMember("marvin-test@prefect.io"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "email", "marvin-test@prefect.io"),
					resource.TestCheckResourceAttrSet(dataSourceName, "id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_role_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "account_role_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "actor_id"),
					resource.TestCheckResourceAttrSet(dataSourceName, "first_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "handle"),
					resource.TestCheckResourceAttrSet(dataSourceName, "user_id"),
				),
			},
		},
	})
}
