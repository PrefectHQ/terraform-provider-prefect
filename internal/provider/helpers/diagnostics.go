package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
)

// https://developer.hashicorp.com/terraform/plugin/framework/diagnostics#custom-diagnostics-types
func CreateClientErrorDiagnostic(clientName string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Error creating %s client", clientName),
		fmt.Sprintf("Could not create %s client, unexpected error: %s. This is a bug in the provider, please report this to the maintainers.", clientName, err.Error()),
	)
}
