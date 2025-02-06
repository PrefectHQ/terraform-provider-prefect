package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// ImportState imports the resource into Terraform state.
//
// We'll allow input values in the form of:
// - "id,workspace_id"
// - "id"
//
// This is the most common import pattern. If the import pattern for
// a resource differs, define it in that resource's `ImportState` method.
func ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	maxInputCount := 2
	inputParts := strings.Split(req.ID, ",")

	// eg. "foo,bar,baz"
	if len(inputParts) > maxInputCount {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected a maximum of 2 import identifiers, in the form of `id,workspace_id`. Got %q", req.ID),
		)

		return
	}

	// eg. ",foo" or "foo,"
	if len(inputParts) == maxInputCount && (inputParts[0] == "" || inputParts[1] == "") {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected non-empty import identifiers, in the form of `id,workspace_id`. Got %q", req.ID),
		)

		return
	}

	if len(inputParts) == maxInputCount {
		id := inputParts[0]
		workspaceID := inputParts[1]

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), id)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
	} else {
		resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
	}
}
