package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

const (
	reportMessage = "Please report this issue to the provider developers."
)

// CreateClientErrorDiagnostic returns an error diagnostic for when one of the
// HTTP clients failed to be instantiated.
//
//nolint:ireturn // required by Terraform API
func CreateClientErrorDiagnostic(clientName string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Error creating %s client", clientName),
		fmt.Sprintf("Could not create %s client, due to error: %s. %s", clientName, err.Error(), reportMessage),
	)
}

// ResourceClientErrorDiagnostic returns an error diagnostic for when a
// client call fails during any resource operations (CRUD).
//
//nolint:ireturn // required by Terraform API
func ResourceClientErrorDiagnostic(resourceName string, operation string, err error) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Error during %s %s", operation, resourceName),
		fmt.Sprintf("Could not %s %s, unexpected error: %s", operation, resourceName, err.Error()),
	)
}

// ConfigureTypeErrorDiagnostic returns an error diagnostic for when a
// given type does not implement PrefectClient.
//
//nolint:ireturn // required by Terraform API
func ConfigureTypeErrorDiagnostic(componentKind string, data any) diag.Diagnostic {
	return diag.NewErrorDiagnostic(
		fmt.Sprintf("Unexpected %s Configure type", componentKind),
		fmt.Sprintf("Expected api.PrefectClient type, got %T. %s", data, reportMessage),
	)
}

// SerializeDataErrorDiagnostic returns an error diagnostic for when an
// attempt to serialize data into a JSON string fails.
//
//nolint:ireturn // required by Terraform API
func SerializeDataErrorDiagnostic(pathRoot, resourceName string, err error) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path.Root(pathRoot),
		fmt.Sprintf("Failed to serialize %s data", resourceName),
		fmt.Sprintf("Could not serialize %s as JSON string", err.Error()),
	)
}
