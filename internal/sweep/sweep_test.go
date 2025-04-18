package sweep_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/prefecthq/terraform-provider-prefect/internal/sweep"
)

// TestMain adds sweeper functionality to the "go test" command.
func TestMain(m *testing.M) {
	sweep.AddWorkspaceSweeper()
	sweep.AddServiceAccountSweeper()
	sweep.AddWorkspaceRoleSweeper()

	resource.TestMain(m)
}
