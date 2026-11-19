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

// Ensure ModuleResource satisfies various resource interfaces.
var _ resource.Resource = &ModuleResource{}
var _ resource.ResourceWithConfigure = &ModuleResource{}
var _ resource.ResourceWithImportState = &ModuleResource{}

// ModuleResource defines the resource implementation for registry modules.
// This is a read-only resource that can only be imported, not created.
type ModuleResource struct {
	client *client.RegistryClient
}

// ModuleResourceModel describes the resource data model.
type ModuleResourceModel struct {
	ID             types.String `tfsdk:"id"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	TargetProvider types.String `tfsdk:"target_provider"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	Verified       types.Bool   `tfsdk:"verified"`
	Downloads      types.Int64  `tfsdk:"downloads"`
}

// NewModuleResource creates a new resource.
func NewModuleResource() resource.Resource {
	return &ModuleResource{}
}

func (r *ModuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module"
}

func (r *ModuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a Terraform module from the registry. This is a read-only resource that can only be imported, not created or modified.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Full module ID (namespace/name/provider).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Module namespace (e.g., terraform-aws-modules).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Module name (e.g., vpc).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"target_provider": schema.StringAttribute{
				Description: "Target provider (e.g., aws).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"description": schema.StringAttribute{
				Description: "Module description.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Source repository URL.",
				Computed:    true,
			},
			"verified": schema.BoolAttribute{
				Description: "Whether the module is verified.",
				Computed:    true,
			},
			"downloads": schema.Int64Attribute{
				Description: "Total download count.",
				Computed:    true,
			},
		},
	}
}

func (r *ModuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ModuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data ModuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// This is a read-only resource - we just read the current state from the registry
	if err := r.readModule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data ModuleResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readModule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data ModuleResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Read-only resource - just refresh from registry
	if err := r.readModule(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *ModuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to delete - this is a read-only resource
	// Just remove from state (default behavior)
}

func (r *ModuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Import ID is "namespace/name/provider", e.g. "terraform-aws-modules/vpc/aws".
	importByID(ctx, req, resp, "namespace", "name", "target_provider")
}

func (r *ModuleResource) readModule(ctx context.Context, data *ModuleResourceModel) error {
	namespace := data.Namespace.ValueString()
	name := data.Name.ValueString()
	targetProvider := data.TargetProvider.ValueString()

	url := fmt.Sprintf("%s/modules/%s/%s/%s", r.client.BaseURL(), namespace, name, targetProvider)

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
		Provider    string `json:"provider"`
		Description string `json:"description"`
		Source      string `json:"source"`
		Verified    bool   `json:"verified"`
		Downloads   int64  `json:"downloads"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return err
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", namespace, name, targetProvider))
	data.Namespace = types.StringValue(namespace)
	data.Name = types.StringValue(name)
	data.TargetProvider = types.StringValue(targetProvider)
	data.Description = types.StringValue(apiResp.Description)
	data.Source = types.StringValue(apiResp.Source)
	data.Verified = types.BoolValue(apiResp.Verified)
	data.Downloads = types.Int64Value(apiResp.Downloads)

	return nil
}
