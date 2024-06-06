package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#custom-diagnostics-types

// CreateClientErrorDiagnostic returns an error diagnostic for when one of the
// HTTP clients failed to be instantiated.
//
//nolint:ireturn // required by Terraform API
func CreateClientErrorDiagnostic(clientName string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Error creating %s client", clientName),
		fmt.Sprintf("Could not create %s client, due to error: %s. If you believe this to be a bug in the provider, please report this to the maintainers.", clientName, err.Error()),
	)
}

// ResourceClientErrorDiagnostic returns an error diagnostic for when a
// client call fails during any resource operations (CRUD).
//
//nolint:ireturn // required by Terraform API
func ResourceClientErrorDiagnostic(resourceName string, operation string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Error during %s %s", operation, resourceName),
		fmt.Sprintf("Could not %s %s, unexpected error: %s", operation, resourceName, err),
	)
}
