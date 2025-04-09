package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"slices"
	"strings"
)

type result struct {
	Paths map[string]string `json:"paths"`
}

// Relates the Prefect API resource names to a list of matching Terraform
// resource names.
//
// TODO: get this from registering Resources and Datasources?
//
//nolint:godox // this script is being saved in the parking lot
var relations = map[string][]string{
	"artifacts":                 {},
	"automations":               {"automation_resource"},
	"block_capabilities":        {"block"},
	"block_documents":           {"block"},
	"block_schemas":             {"block"},
	"block_types":               {"block"},
	"bot_access":                {"workspace_access"},
	"collections":               {},
	"concurrency_limits":        {"task_run_concurrency_limit"},
	"global_concurrency_limits": {"global_concurrency_limit"},
	"deployments":               {"deployment"},
	"events":                    {},
	"flows":                     {"flow"},
	"slas":                      {},
	"team_access":               {"workspace_access"},
	"user_access":               {"workspace_access"},
	"variables":                 {"variable"},
	"webhooks":                  {"webhooks"},
	"work_pools":                {"work_pool", "work_pools"},
	"work_queues":               {"work_queue", "work_queues"},
}

// Some resources are intentionally ignored because they do not make sense
// to implement in the provider.
var ignoredFields = []string{
	"access",
	"api",
	"auth",
	"flow_runs",
	"flow_run_states",
	"health",
	"hello",
	"invitations",
	"lenses",
	"logs",
	"managed_execution",
	"notifications",
	"resources",
	"saved_searches",
	"tags",
	"task_run_states",
	"task_runs",
	"task_workers",
	"templates",
	"transfer",
	"ui",
	"validate_transfer",
	"v2", // global concurrency limits
}

func main() {
	for _, api := range getNonImplementedAPIs(
		getAPIResources(),
		getTerraformResources(),
	) {
		log.Printf("API '%s' has no Terraform resource equivalent", api)
	}
}

// getNonImplementedAPIs will return a list of API resources that have no
// Terraform resource equivalent.
func getNonImplementedAPIs(resultsAPI, resultsTF []string) []string {
	result := make([]string, 0)

	// Check each API resource for a matching TF resource.
	for _, resultAPI := range resultsAPI {
		// Get the names of the related Terraform resource names for the given API.
		relatedNames, ok := relations[resultAPI]
		if !ok {
			// If we didn't find any related names, make sure it's not because
			// it's one of the fields we intentionally ignore.
			if !slices.Contains(ignoredFields, resultAPI) {
				log.Printf("No related Terraform resources found for API resource: %s", resultAPI)
			}

			continue
		}

		// Check if any of the related names are found in the Terraform resources.
		found := false
		for _, relatedName := range relatedNames {
			if slices.Contains(resultsTF, relatedName) {
				found = true
			}
		}

		if !found {
			result = append(result, resultAPI)
		}
	}

	return result
}

// getTerraformResources will return a list of Terraform resources that are
// implemented based on available files under `internal/provider/`.
func getTerraformResources() []string {
	//nolint:godox // this script is being saved in the parking lot
	// TODO: also get datasources
	entries, err := os.ReadDir("internal/provider/resources")
	if err != nil {
		log.Printf("Error reading directory: %v", err)

		return nil
	}

	result := make([]string, 0)
	for _, entry := range entries {
		if strings.Contains(entry.Name(), "_test") {
			continue
		}

		name := strings.ReplaceAll(entry.Name(), ".go", "")

		result = append(result, name)
	}

	return result
}

func getAPIResources() []string {
	r := request()

	result := make([]string, 0)
	for key := range r.Paths {
		key = cleanAPIResource(key)

		if key == "" {
			continue
		}

		result = append(result, key)
	}

	result = dedupe(result)

	return result
}

func cleanAPIResource(resource string) string {
	resource = strings.ReplaceAll(resource, "/api/accounts/{account_id}/workspaces/{workspace_id}", "")

	parts := strings.Split(resource, "/")
	if len(parts) > 1 {
		return parts[1]
	}

	return resource
}

func dedupe(list []string) []string {
	slices.Sort(list)

	return slices.Compact(list)
}

func request() result {
	var r result

	ctx := context.Background()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://api.prefect.cloud/api/openapi.json", http.NoBody)
	if err != nil {
		log.Printf("Error creating request: %v", err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("Error getting openapi.json: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading body: %v", err)

		return r
	}

	// Ignore the error because it's known that there is an error unmarshalling
	// because the payload isn't completely mapped out in the 'result' struct.
	_ = json.Unmarshal(body, &r)

	return r
}
