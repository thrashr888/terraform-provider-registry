// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
)

// Ensure PolicyLibraryResource satisfies various resource interfaces.
var _ resource.Resource = &PolicyLibraryResource{}
var _ resource.ResourceWithConfigure = &PolicyLibraryResource{}
var _ resource.ResourceWithImportState = &PolicyLibraryResource{}

// PolicyLibraryResource defines the resource implementation for policy libraries.
type PolicyLibraryResource struct {
	client *client.RegistryClient
}

// PolicyLibraryResourceModel describes the resource data model.
type PolicyLibraryResourceModel struct {
	ID        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
	Title     types.String `tfsdk:"title"`
	Source    types.String `tfsdk:"source"`
	Verified  types.Bool   `tfsdk:"verified"`
	Downloads types.Int64  `tfsdk:"downloads"`
}

// NewPolicyLibraryResource creates a new resource.
func NewPolicyLibraryResource() resource.Resource {
	return &PolicyLibraryResource{}
}

func (r *PolicyLibraryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_library"
}

func (r *PolicyLibraryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Represents a Terraform policy library from the registry. This is a read-only resource that can only be imported, not created or modified.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Full policy library ID (namespace/name).",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"namespace": schema.StringAttribute{
				Description: "Policy library namespace (e.g., hashicorp).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Description: "Policy library name (e.g., gcp-networking-terraform).",
				Required:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"title": schema.StringAttribute{
				Description: "Policy library title.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Source repository URL.",
				Computed:    true,
			},
			"verified": schema.BoolAttribute{
				Description: "Whether the policy library is verified.",
				Computed:    true,
			},
			"downloads": schema.Int64Attribute{
				Description: "Total download count.",
				Computed:    true,
			},
		},
	}
}

func (r *PolicyLibraryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PolicyLibraryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data PolicyLibraryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readPolicyLibrary(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicyLibraryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data PolicyLibraryResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readPolicyLibrary(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicyLibraryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data PolicyLibraryResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.readPolicyLibrary(ctx, &data); err != nil {
		resp.Diagnostics.AddError("Registry API Error", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *PolicyLibraryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// Nothing to delete - this is a read-only resource
}

func (r *PolicyLibraryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *PolicyLibraryResource) readPolicyLibrary(ctx context.Context, data *PolicyLibraryResourceModel) error {
	namespace := data.Namespace.ValueString()
	name := data.Name.ValueString()

	url := fmt.Sprintf("https://%s/v2/policies/%s/%s", r.client.Host, namespace, name)

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
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var apiResp struct {
		Data struct {
			ID         string `json:"id"`
			Attributes struct {
				FullName  string `json:"full-name"`
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				Title     string `json:"title"`
				Source    string `json:"source"`
				Verified  bool   `json:"verified"`
				Downloads int64  `json:"downloads"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return err
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", namespace, name))
	data.Namespace = types.StringValue(namespace)
	data.Name = types.StringValue(name)
	data.Title = types.StringValue(apiResp.Data.Attributes.Title)
	data.Source = types.StringValue(apiResp.Data.Attributes.Source)
	data.Verified = types.BoolValue(apiResp.Data.Attributes.Verified)
	data.Downloads = types.Int64Value(apiResp.Data.Attributes.Downloads)

	return nil
}
