package testutils

import (
	"fmt"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func getResourceFromState(state *terraform.State, resourceName string) (*terraform.ResourceState, error) {
	fetchedResource, exists := state.RootModule().Resources[resourceName]
	if !exists {
		return nil, fmt.Errorf("resource not found in state: %s", resourceName)
	}

	return fetchedResource, nil
}

func getResourceIDFromState(state *terraform.State, resourceName, attribute string) (uuid.UUID, error) {
	fetchedResource, err := getResourceFromState(state, resourceName)
	if err != nil {
		return uuid.Nil, fmt.Errorf("resource not found in state: %s", resourceName)
	}

	fetchedResourceID, err := uuid.Parse(fetchedResource.Primary.Attributes[attribute])
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing resource ID: %w", err)
	}

	return fetchedResourceID, nil
}

func getResourceAttributeFromState(state *terraform.State, resourceName, attribute string) (string, error) {
	fetchedResource, err := getResourceFromState(state, resourceName)
	if err != nil {
		return "", fmt.Errorf("resource not found in state: %s", resourceName)
	}

	fetchedResourceAttribute, exists := fetchedResource.Primary.Attributes[attribute]
	if !exists {
		return "", fmt.Errorf("error parsing resource ID: %w", err)
	}

	return fetchedResourceAttribute, nil
}

// GetResourceIDFromState returns the ID of a resource from the state as a UUID.
func GetResourceIDFromState(state *terraform.State, resourceName string) (uuid.UUID, error) {
	return getResourceIDFromState(state, resourceName, "id")
}

// GetResourceAttributeFromState returns the attribute of a resource from the state as a UUID.
func GetResourceIDFromStateByAttribute(state *terraform.State, resourceName, attribute string) (uuid.UUID, error) {
	return getResourceIDFromState(state, resourceName, attribute)
}

// GetResourceAttributeFromStateByAttribute returns the attribute of a resource from the state as a string.
func GetResourceAttributeFromStateByAttribute(state *terraform.State, resourceName, attribute string) (string, error) {
	return getResourceAttributeFromState(state, resourceName, attribute)
}

// GetResourceWorkspaceIDFromState returns the workspace ID from the state as a UUID.
func GetResourceWorkspaceIDFromState(state *terraform.State) (uuid.UUID, error) {
	workspace, exists := state.RootModule().Resources[WorkspaceResourceName]
	if !exists {
		return uuid.Nil, fmt.Errorf("workspace resource not found in state: %s", WorkspaceResourceName)
	}

	workspaceID, err := uuid.Parse(workspace.Primary.ID)
	if err != nil {
		return uuid.Nil, fmt.Errorf("error parsing workspace ID: %w", err)
	}

	return workspaceID, nil
}

// GetResourceWorkspaceImportStateID returns the import state ID for a resource
// for use in import tests.
func GetResourceWorkspaceImportStateID(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceID, err := GetResourceWorkspaceIDFromState(state)
		if err != nil {
			return "", fmt.Errorf("unable to get workspaceID from state: %w", err)
		}

		fetchedResourceID, err := GetResourceIDFromState(state, resourceName)
		if err != nil {
			return "", fmt.Errorf("unable to get resource from state: %w", err)
		}

		return fmt.Sprintf("%s,%s", fetchedResourceID, workspaceID), nil
	}
}

// GetResourceWorkspaceImportStateIDByName returns the import state ID for a resource from the
// "name" attribute for use in import tests.
func GetResourceWorkspaceImportStateIDByName(resourceName string) resource.ImportStateIdFunc {
	return func(state *terraform.State) (string, error) {
		workspaceID, err := GetResourceWorkspaceIDFromState(state)
		if err != nil {
			return "", fmt.Errorf("unable to get workspaceID from state: %w", err)
		}

		fetchedResourceID, err := GetResourceAttributeFromStateByAttribute(state, resourceName, "name")
		if err != nil {
			return "", fmt.Errorf("unable to get resource from state: %w", err)
		}

		return fmt.Sprintf("%s,%s", fetchedResourceID, workspaceID), nil
	}
}
