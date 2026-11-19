// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
)

// Ensure ProviderResource satisfies various resource interfaces.
var _ resource.Resource = &ProviderResource{}
var _ resource.ResourceWithConfigure = &ProviderResource{}
var _ resource.ResourceWithImportState = &ProviderResource{}
var _ resource.ResourceWithIdentity = &ProviderResource{}

// ProviderResource defines the resource implementation for registry providers.
// This is a read-only resource that can only be imported, not created.
type ProviderResource struct {
	client *client.RegistryClient
}

// ProviderResourceModel describes the resource data model.
type ProviderResourceModel struct {
	ID            types.String `tfsdk:"id"`
	Namespace     types.String `tfsdk:"namespace"`
	Name          types.String `tfsdk:"name"`
	Description   types.String `tfsdk:"description"`
	Source        types.String `tfsdk:"source"`
	Tier          types.String `tfsdk:"tier"`
	PublishedAt   types.String `tfsdk:"published_at"`
	Downloads     types.Int64  `tfsdk:"downloads"`
	LatestVersion types.String `tfsdk:"latest_version"`
}

// NewProviderResource creates a new resource.
func NewProviderResource() resource.Resource {
	return &ProviderResource{}
}

func (r *ProviderResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

func (r *ProviderResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a Terraform provider from the registry. This is a read-only resource that can only be imported, not created or modified.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Full provider ID (namespace/name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Provider namespace (e.g., hashicorp).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Provider name (e.g., aws).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Provider description.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Source repository URL.",
				Computed:    true,
			},
			"tier": schema.StringAttribute{
				Description: "Provider tier (official, partner, community).",
				Computed:    true,
			},
			"published_at": schema.StringAttribute{
				Description: "Publication date.",
				Computed:    true,
			},
			"downloads": schema.Int64Attribute{
				Description: "Total download count.",
				Computed:    true,
			},
			"latest_version": schema.StringAttribute{
				Description: "Latest version string.",
				Computed:    true,
			},
		},
	}
}

func (r *ProviderResource) IdentitySchema(ctx context.Context, req resource.IdentitySchemaRequest, resp *resource.IdentitySchemaResponse) {
	resp.IdentitySchema = namespaceNameIdentitySchema("Provider")
}

func (r *ProviderResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.RegistryClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.RegistryClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *ProviderResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ProviderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// This is a read-only resource - we just read the current state from the registry
	if err := r.readProvider(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(setIdentity(ctx, resp.Identity, ProviderIdentityModel{Namespace: data.Namespace, Name: data.Name})...)
}

func (r *ProviderResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ProviderResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readProvider(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(setIdentity(ctx, resp.Identity, ProviderIdentityModel{Namespace: data.Namespace, Name: data.Name})...)
}

func (r *ProviderResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ProviderResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read-only resource - just refresh from registry
	if err := r.readProvider(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
	resp.Diagnostics.Append(setIdentity(ctx, resp.Identity, ProviderIdentityModel{Namespace: data.Namespace, Name: data.Name})...)
}

func (r *ProviderResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to delete - this is a read-only resource
	// Just remove from state (default behavior)
}

func (r *ProviderResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID is "namespace/name", e.g. "hashicorp/aws".
	importByID(ctx, req, resp, "namespace", "name")
}

func (r *ProviderResource) readProvider(ctx context.Context, data *ProviderResourceModel) error {
	namespace := data.Namespace.ValueString()
	name := data.Name.ValueString()

	// If we only have ID, parse it
	if namespace == "" && !data.ID.IsNull() {
		id := data.ID.ValueString()
		var parsed struct{ namespace, name string }
		if _, err := fmt.Sscanf(id, "%s/%s", &parsed.namespace, &parsed.name); err == nil {
			namespace = parsed.namespace
			name = parsed.name
		}
	}

	url := fmt.Sprintf("%s/providers/%s/%s", r.client.BaseURL(), namespace, name)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return err
	}

	if r.client.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+r.client.Token)
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return err
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var apiResp struct {
		ID          string `json:"id"`
		Namespace   string `json:"namespace"`
		Name        string `json:"name"`
		Version     string `json:"version"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Tier        string `json:"tier"`
		PublishedAt string `json:"published_at"`
		Downloads   int64  `json:"downloads"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return err
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", namespace, name))
	data.Namespace = types.StringValue(namespace)
	data.Name = types.StringValue(name)
	data.Description = types.StringValue(apiResp.Description)
	data.Source = types.StringValue(apiResp.Source)
	data.Tier = types.StringValue(apiResp.Tier)
	data.PublishedAt = types.StringValue(apiResp.PublishedAt)
	data.Downloads = types.Int64Value(apiResp.Downloads)
	data.LatestVersion = types.StringValue(apiResp.Version)

	return nil
}
