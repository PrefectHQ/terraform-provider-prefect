package provider_test

import (
	"context"
	"fmt"
	"os"
	"terraform-provider-prefect/internal/prefect"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestMain(m *testing.M) {
	resource.TestMain(m)
}

func sharedClientForRegion(region string) (*prefect.Client, error) {
	apiServer := os.Getenv("PREFECT__CLOUD__API")
	if apiServer == "" {
		apiServer = prefect.DefaultAPIServer
	}

	apiKey := os.Getenv("PREFECT__CLOUD__API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("env var PREFECT__CLOUD__API_KEY must be specified")
	}

	client, err := prefect.NewClient(context.Background(), &apiKey, &apiServer)
	if err != nil {
		return nil, fmt.Errorf("unable to create prefect client: %s", err.Error())
	}

	return client, nil
}
