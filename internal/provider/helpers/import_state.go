package helpers

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func importState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, identifier string) {
	maxInputCount := 2
	inputParts := strings.Split(req.ID, ",")

	// eg. "foo,bar,baz"
	if len(inputParts) > maxInputCount {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected a maximum of 2 import identifiers, in the form of `%s,workspace_id`. Got %q", identifier, req.ID),
		)

		return
	}

	// eg. ",foo" or "foo,"
	if len(inputParts) == maxInputCount && (inputParts[0] == "" || inputParts[1] == "") {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected non-empty import identifiers, in the form of `%s,workspace_id`. Got %q", identifier, req.ID),
		)

		return
	}

	if len(inputParts) == maxInputCount {
		id := inputParts[0]
		workspaceID := inputParts[1]

		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(identifier), id)...)
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("workspace_id"), workspaceID)...)
	} else {
		resource.ImportStatePassthroughID(ctx, path.Root(identifier), req, resp)
	}
}

// ImportState imports the resource into Terraform state.
//
// Allows input values in the form of:
// - "id,workspace_id"
// - "id"
//
// To import by name instead of ID, see ImportStateByName.
func ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importState(ctx, req, resp, "id")
}

// ImportState imports the resource into Terraform state.
//
// Allows input values in the form of:
// - "name,workspace_id"
// - "name"
//
// To import by ID instead of name, see ImportState.
func ImportStateByName(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	importState(ctx, req, resp, "name")
}
