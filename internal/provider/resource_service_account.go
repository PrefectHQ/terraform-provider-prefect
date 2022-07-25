package provider

import (
	"context"
	"fmt"

	"terraform-provider-prefect/api"
	"terraform-provider-prefect/internal/util"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type serviceAccountResourceType struct{}

// service_account resource schema
// nolint:funlen
func (r serviceAccountResourceType) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{

		Description: `
Service account resource.
`,
		Attributes: map[string]tfsdk.Attribute{
			"id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Server generated UUID.",
			},
			"name": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					StringNotNull(),
				},
				Description: "Name",
			},
			"role": {
				Type:     types.StringType,
				Required: true,
				Validators: []tfsdk.AttributeValidator{
					RoleIsValid(),
				},
				Description: fmt.Sprintf("role must be one of [%s, %s, %s].", api.Membership_roleReadOnlyUser, api.Membership_roleTenantAdmin, api.Membership_roleUser),
			},
			"membership_id": {
				Type:        types.StringType,
				Computed:    true,
				Description: "Server generated UUID.",
			},
			"api_keys": {
				Optional: true,
				Validators: []tfsdk.AttributeValidator{
					APIKeyNamesAreUnique(),
				},
				Description: "API keys. Keys are identified by their name in config so names must be unique.",
				Attributes: tfsdk.ListNestedAttributes(map[string]tfsdk.Attribute{
					"id": {
						Type:        types.StringType,
						Computed:    true,
						Description: "Server generated UUID.",
					},
					"name": {
						Type:     types.StringType,
						Required: true,
						Validators: []tfsdk.AttributeValidator{
							StringNotNull(),
						},
						Description: "Name. Changing the name will delete the old key and create a new one.",
					},
					"expiration": {
						Type:     types.StringType,
						Optional: true,
						Computed: true,
						Validators: []tfsdk.AttributeValidator{
							IsRFC3339Time(),
						},
						Description: "Expiration date time in RFC3339 UTC format, eg: 2015-10-21T16:29:00+00:00",
					},
					"key": {
						Type:        types.StringType,
						Computed:    true,
						Description: "Key",
						Sensitive:   true,
					},
				}),
			},
		},
	}, nil
}

// New resource instance
func (r serviceAccountResourceType) NewResource(_ context.Context, in tfsdk.Provider) (tfsdk.Resource, diag.Diagnostics) {
	provider, diags := convertProviderType(in)

	return serviceAccountResource{
		provider: provider,
	}, diags
}

type serviceAccountResource struct {
	provider provider
}

type serviceAccountResourceData struct {
	ID           types.String `tfsdk:"id"`
	Name         types.String `tfsdk:"name"`
	Role         types.String `tfsdk:"role"`
	MembershipId types.String `tfsdk:"membership_id"`
	APIKeys      []apiKey     `tfsdk:"api_keys"`
}

type apiKey struct {
	ID         types.String `tfsdk:"id"`
	Name       types.String `tfsdk:"name"`
	Expiration types.String `tfsdk:"expiration"`
	Key        types.String `tfsdk:"key"`
}

// Create a new resource
// nolint:funlen
func (r serviceAccountResource) Create(ctx context.Context, req tfsdk.CreateResourceRequest, resp *tfsdk.CreateResourceResponse) {
	if !r.provider.configured {
		resp.Diagnostics.AddError(
			"Provider not configured",
			"The provider hasn't been configured before apply, likely because it depends on an unknown value from another resource.",
		)
		return
	}

	// Retrieve values from config (ie: .tf file)
	var config serviceAccountResourceData
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Lookup role id
	var roleId *api.UUID
	if roleName := util.ToString(config.Role); roleName != nil {
		if id, ok := r.provider.client.Tenant.RoleIds[*roleName]; ok {
			roleId = &id
		} else {
			resp.Diagnostics.AddError(
				"Invalid role name",
				fmt.Sprintf("Invalid role name: %v", *roleName),
			)
			return
		}
	}

	// Create new service account
	serviceAccount, err := api.CreateServiceAccount(r.provider.client.GQLClient, ctx, r.provider.client.Tenant.Id, config.Name.Value, roleId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not create service account",
			fmt.Sprintf("%v", err),
		)
		return
	}

	serviceAccountId := string(*serviceAccount.Id)

	// Get membership id
	users, err := api.GetTenantUser(r.provider.client.GQLClient, ctx, serviceAccountId)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read service account",
			fmt.Sprintf("%v (service account ID = %s)", err, serviceAccountId),
		)
		return
	}
	// First and only membership is for the current tenant
	var membershipId = users[0].Memberships[0].Id

	// Generate api keys
	for i, apikey := range config.APIKeys {
		createdKey, err := api.CreateAPIKey(r.provider.client.GQLClient, ctx,
			*serviceAccount.Id, apikey.Name.Value, (*api.DateTime)(util.ToString(apikey.Expiration)), nil)
		if err != nil {
			resp.Diagnostics.AddError(
				fmt.Sprintf("Could not create api key %s", apikey.Name.Value),
				fmt.Sprintf("%v (service account ID = %s)", err, serviceAccountId),
			)
			return
		}
		// Update values that were unknown from API response
		config.APIKeys[i].ID = types.String{Value: string(*createdKey.Id)}
		config.APIKeys[i].Key = types.String{Value: *createdKey.Key}
	}

	// Update values that were unknown from API response
	config.ID = types.String{Value: serviceAccountId}
	config.MembershipId = types.String{Value: membershipId}

	// Set state
	diags = resp.State.Set(ctx, config)
	resp.Diagnostics.Append(diags...)
}

// Read resource information
func (r serviceAccountResource) Read(ctx context.Context, req tfsdk.ReadResourceRequest, resp *tfsdk.ReadResourceResponse) {
	// Get current state
	var state serviceAccountResourceData

	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get user from API
	users, err := api.GetTenantUser(r.provider.client.GQLClient, ctx, state.ID.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not read service account",
			fmt.Sprintf("%v (service account ID = %s)", err, state.ID.Value),
		)
		return
	}

	if len(users) == 0 {
		// doesn't exist ie: has been deleted outside of terraform
		resp.State.RemoveResource(ctx)
		return
	}

	var user = users[0]

	if len(user.Memberships) != 1 {
		resp.Diagnostics.AddError(
			"Unexpected memberships",
			fmt.Sprintf("Service account %s has %d memberships, expected only 1.", state.Name.Value, len(user.Memberships)))
		return
	}

	// Update state with values from API response
	state.Name = types.String{Value: *user.First_name}
	state.MembershipId = types.String{Value: users[0].Memberships[0].Id}
	state.Role = types.String{Value: user.Memberships[0].Role_detail.Name}

	apiKeys, err := r.fetchAPIKeysAsState(ctx, user.Id, state)
	if err != nil {
		diags.AddError(
			"Could not read api keys",
			fmt.Sprintf("%v (user ID = %s)", err, *user.Id),
		)
	}
	state.APIKeys = apiKeys

	// Set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r serviceAccountResource) fetchAPIKeysAsState(ctx context.Context, userId *string, state serviceAccountResourceData) ([]apiKey, error) {

	// fetch keys from API
	fetchedKeys, err := api.APIKeysByUser(r.provider.client.GQLClient, ctx, *userId)
	if err != nil {
		return nil, err
	}

	if len(fetchedKeys) == 0 {
		return nil, nil
	}

	// turn state keys into map of indexed keys
	type IndexedAPIKey struct {
		// position of the key in state
		index  int
		apiKey *apiKey
	}

	stateKeysById := make(map[string]IndexedAPIKey)
	for i, apiKey := range state.APIKeys {
		stateKeysById[apiKey.ID.Value] = IndexedAPIKey{i, &state.APIKeys[i]}
	}

	// build state from fetched keys
	// order the fetched keys by their position in state, followed by keys not in state
	apiKeys := make([]apiKey, len(fetchedKeys))

	// insertion index for new keys (ie: keys not in state) begins at the end of array
	newKeyIndex := len(fetchedKeys) - 1
	for _, fetchedKey := range fetchedKeys {
		var index int
		var key types.String
		stateKey, ok := stateKeysById[fetchedKey.Id]
		if ok {
			index = stateKey.index
			// we can never get the key back after creation, so just pass it through
			// TODO: add a redact_on_fresh mode which redacts the key so its removed from state on refresh
			key = stateKey.apiKey.Key
		} else {
			// new key created outside of terraform, add from end of array
			index = newKeyIndex
			newKeyIndex--

			// we can never get the key back after creation, so declare it unknown
			key = types.String{Unknown: true}
		}

		apiKeys[index] = apiKey{
			ID:         types.String{Value: fetchedKey.Id},
			Name:       types.String{Value: fetchedKey.Name},
			Expiration: util.FromString(fetchedKey.Expires_at),
			Key:        key,
		}
	}

	return apiKeys, nil
}

// Update resource
func (r serviceAccountResource) Update(ctx context.Context, req tfsdk.UpdateResourceRequest, resp *tfsdk.UpdateResourceResponse) {
	// Get plan values
	var plan serviceAccountResourceData
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get current state
	var state serviceAccountResourceData
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// update name
	if !state.Name.Equal(plan.Name) {
		resp.Diagnostics.AddError(
			"Error updating name",
			"Prefect does not allow admins to change user names. Please revert your config change. See https://github.com/PrefectHQ/prefect/issues/5542",
		)
	}

	// update role
	if !state.Role.Equal(plan.Role) {
		_, err := api.SetMembershipRole(r.provider.client.GQLClient, ctx, (api.UUID)(state.MembershipId.Value), plan.Role.Value)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating role",
				fmt.Sprintf("%v (MembershipShipId = %s)", err, state.MembershipId.Value),
			)
		} else {
			state.Role = plan.Role
		}
	}

	// update api keys
	r.updateAPIKeys(ctx, &state, &plan, &resp.Diagnostics)

	// bail if any failures without setting state
	if diags.HasError() {
		return
	}

	// Set state
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// nolint: funlen
func (r serviceAccountResource) updateAPIKeys(ctx context.Context,
	state *serviceAccountResourceData, plan *serviceAccountResourceData, diags *diag.Diagnostics) {

	if !Equal(state.APIKeys, plan.APIKeys) {

		// index by name because that's all we have in the plan (ID is unknown at this stage)
		stateKeysByName := make(map[string]*apiKey)
		for i, apiKey := range state.APIKeys {
			stateKeysByName[apiKey.Name.Value] = &state.APIKeys[i]
		}

		// rebuild state from plan, updating via the Prefect API as we go
		state.APIKeys = make([]apiKey, len(plan.APIKeys))

		for i, planKey := range plan.APIKeys {
			stateKey, ok := stateKeysByName[planKey.Name.Value]

			if ok {
				// check expiration hasn't changed
				if !stateKey.Expiration.Equal(plan.APIKeys[i].Expiration) {
					diags.AddError(
						"Error updating api key expiration.",
						"api key expiration cannot be changed after creation. Please revert your config change.",
					)
					continue
				}
				state.APIKeys[i] = *stateKey
				delete(stateKeysByName, planKey.Name.Value)
			} else {
				// create new key
				createdKey, err := api.CreateAPIKey(r.provider.client.GQLClient, ctx,
					api.UUID(state.ID.Value), planKey.Name.Value, (*api.DateTime)(util.ToString(planKey.Expiration)), nil)
				if err != nil {
					diags.AddError(
						fmt.Sprintf("Could not create api key %s", planKey.Name.Value),
						fmt.Sprintf("%v (service account ID = %s)", err, state.ID.Value),
					)
					continue
				}
				state.APIKeys[i] = apiKey{
					ID:         types.String{Value: string(*createdKey.Id)},
					Name:       planKey.Name,
					Expiration: planKey.Expiration,
					Key:        types.String{Value: *createdKey.Key},
				}
			}
		}

		// delete any api keys that are left in state_keys_map and therefore missing from the plan
		for _, orphanKey := range stateKeysByName {
			successPayload, err := api.DeleteAPIKey(r.provider.client.GQLClient, ctx, api.UUID(orphanKey.ID.Value))
			if err != nil {
				diags.AddError(
					"Could not delete api key",
					fmt.Sprintf("%v (api key name = %s)", err, orphanKey.Name.Value),
				)
				continue
			}
			if !*successPayload.Success {
				diags.AddError(
					"Could not delete api key",
					fmt.Sprintf("%v (api key name = %s)", *successPayload.Error, orphanKey.Name.Value),
				)
				continue
			}
		}
	}
}

func Equal(left, right []apiKey) bool {
	if len(left) != len(right) {
		return false
	}

	for i := range left {
		if !left[i].Name.Equal(right[i].Name) || !left[i].Expiration.Equal(right[i].Expiration) {
			return false
		}
	}
	return true
}

// Delete resource
func (r serviceAccountResource) Delete(ctx context.Context, req tfsdk.DeleteResourceRequest, resp *tfsdk.DeleteResourceResponse) {
	var state serviceAccountResourceData
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	successPayload, err := api.DeleteServiceAccount(r.provider.client.GQLClient, ctx, (api.UUID)(state.ID.Value))
	if err != nil {
		resp.Diagnostics.AddError(
			"Could not delete service account",
			fmt.Sprintf("%v (service account ID = %s)", err, state.ID.Value),
		)
		return
	}
	if !*successPayload.Success {
		resp.Diagnostics.AddError(
			"Could not delete service account",
			fmt.Sprintf("%v (service account ID = %s)", *successPayload.Error, state.ID.Value),
		)
		return
	}

	// Remove resource from state
	resp.State.RemoveResource(ctx)
}

// Import resource
func (r serviceAccountResource) ImportState(ctx context.Context, req tfsdk.ImportResourceStateRequest, resp *tfsdk.ImportResourceStateResponse) {
	tfsdk.ResourceImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
