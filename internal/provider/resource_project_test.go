package provider_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccProjectResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccProjectResourceConfig("testacc-project", "acceptance test project"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prefect_project.test", "name", "testacc-project"),
					resource.TestCheckResourceAttr("prefect_project.test", "description", "acceptance test project"),
				),
			},
			// Import and refresh
			{
				ResourceName:      "prefect_project.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Update name and description and refresh
			{
				Config: testAccProjectResourceConfig("testacc-project-better-name", "better description"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("prefect_project.test", "name", "testacc-project-better-name"),
					resource.TestCheckResourceAttr("prefect_project.test", "description", "better description"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

func testAccProjectResourceConfig(name string, description string) string {
	return fmt.Sprintf(`
resource "prefect_project" "test" {
	name = %[1]q
	description = %[2]q
}
`, name, description)
}
