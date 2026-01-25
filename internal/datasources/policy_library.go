// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
)

// Ensure PolicyLibraryDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &PolicyLibraryDataSource{}
var _ datasource.DataSourceWithConfigure = &PolicyLibraryDataSource{}

// PolicyLibraryDataSource defines the data source implementation.
type PolicyLibraryDataSource struct {
	client *client.RegistryClient
}

// PolicyLibraryDataSourceModel describes the data source data model.
type PolicyLibraryDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
	FullName  types.String `tfsdk:"full_name"`
	Title     types.String `tfsdk:"title"`
	Source    types.String `tfsdk:"source"`
	Ingress   types.String `tfsdk:"ingress"`
	OwnerName types.String `tfsdk:"owner_name"`
	Downloads types.Int64  `tfsdk:"downloads"`
	Verified  types.Bool   `tfsdk:"verified"`
}

// NewPolicyLibraryDataSource creates a new data source.
func NewPolicyLibraryDataSource() datasource.DataSource {
	return &PolicyLibraryDataSource{}
}

func (d *PolicyLibraryDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_library"
}

func (d *PolicyLibraryDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Sentinel or OPA policy library from the registry.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Policy library ID.",
				Computed:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "Policy library namespace (e.g., hashicorp).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Policy library name.",
				Required:    true,
			},
			"full_name": schema.StringAttribute{
				Description: "Full policy library name (namespace/name).",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "Policy library title/description.",
				Computed:    true,
			},
			"source": schema.StringAttribute{
				Description: "Source repository URL.",
				Computed:    true,
			},
			"ingress": schema.StringAttribute{
				Description: "Ingress type (e.g., github).",
				Computed:    true,
			},
			"owner_name": schema.StringAttribute{
				Description: "Owner name.",
				Computed:    true,
			},
			"downloads": schema.Int64Attribute{
				Description: "Total download count.",
				Computed:    true,
			},
			"verified": schema.BoolAttribute{
				Description: "Whether the policy library is verified.",
				Computed:    true,
			},
		},
	}
}

func (d *PolicyLibraryDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.RegistryClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.RegistryClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = c
}

func (d *PolicyLibraryDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PolicyLibraryDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch policy library from registry v2 API
	url := fmt.Sprintf("https://%s/v2/policies/%s/%s",
		d.client.Host,
		data.Namespace.ValueString(),
		data.Name.ValueString(),
	)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		resp.Diagnostics.AddError("Request Error", err.Error())
		return
	}

	if d.client.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	}

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		resp.Diagnostics.AddError("HTTP Error", err.Error())
		return
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Registry API Error",
			fmt.Sprintf("Unexpected status code: %d", httpResp.StatusCode),
		)
		return
	}

	var apiResp PolicyLibraryAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(apiResp.Data.ID)
	data.FullName = types.StringValue(apiResp.Data.Attributes.FullName)
	data.Title = types.StringValue(apiResp.Data.Attributes.Title)
	data.Source = types.StringValue(apiResp.Data.Attributes.Source)
	data.Ingress = types.StringValue(apiResp.Data.Attributes.Ingress)
	data.OwnerName = types.StringValue(apiResp.Data.Attributes.OwnerName)
	data.Downloads = types.Int64Value(apiResp.Data.Attributes.Downloads)
	data.Verified = types.BoolValue(apiResp.Data.Attributes.Verified)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// PolicyLibraryAPIResponse represents the registry API response for a policy library.
type PolicyLibraryAPIResponse struct {
	Data PolicyLibraryData `json:"data"`
}

// PolicyLibraryData represents the data object in the response.
type PolicyLibraryData struct {
	Type       string                    `json:"type"`
	ID         string                    `json:"id"`
	Attributes PolicyLibraryAttributes   `json:"attributes"`
}

// PolicyLibraryAttributes represents the policy library attributes.
type PolicyLibraryAttributes struct {
	Downloads int64  `json:"downloads"`
	FullName  string `json:"full-name"`
	Ingress   string `json:"ingress"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	OwnerName string `json:"owner-name"`
	Source    string `json:"source"`
	Title     string `json:"title"`
	Verified  bool   `json:"verified"`
}
