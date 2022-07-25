//go:build tools

package tools

import (
	// Used for documentation generation - added here to prevent go mod tidy from removing it
	_ "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs"
)
