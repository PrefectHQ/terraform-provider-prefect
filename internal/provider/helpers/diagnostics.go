package helpers

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
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

// ParseUUIDErrorDiagnostic returns an error diagnostic for when an attempt
// to parse a string into a UUID fails.
//
//nolint:ireturn // required by Terraform API
func ParseUUIDErrorDiagnostic(resourceName string, err error) diag.Diagnostic {
	return diag.NewAttributeErrorDiagnostic(
		path.Root("id"),
		fmt.Sprintf("Error parsing %s ID", resourceName),
		fmt.Sprintf("Could not parse %s ID to UUID, unexpected error: %s", resourceName, err.Error()),
	)
}

// AddProfileWarning Helper function to handle warning messages.
func AddProfileWarning(resp *provider.ConfigureResponse, profileName, profileFilePath string, err error) {
	var title, message string

	// Determine the file path to display
	filePath := profileFilePath
	if filePath == "" {
		filePath = "~/.prefect/profiles.toml"
	}

	// Determine title and message based on whether profile name is specified
	if profileName != "" {
		title = "Failed to load specified Prefect profile"
		message = fmt.Sprintf("Could not load Prefect profile '%s' from %s: %s", profileName, filePath, err)
	} else {
		title = "Failed to load Prefect profile"
		message = fmt.Sprintf("Could not load Prefect profile from %s: %s", filePath, err)
	}

	resp.Diagnostics.AddWarning(title, message)
}
