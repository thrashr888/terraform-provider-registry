// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure ProviderVersionsDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &ProviderVersionsDataSource{}
var _ datasource.DataSourceWithConfigure = &ProviderVersionsDataSource{}

// ProviderVersionsDataSource defines the data source implementation.
type ProviderVersionsDataSource struct {
	client *RegistryClient
}

// ProviderVersionsDataSourceModel describes the data source data model.
type ProviderVersionsDataSourceModel struct {
	ID        types.String `tfsdk:"id"`
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
	Versions  types.List   `tfsdk:"versions"`
}

// NewProviderVersionsDataSource creates a new data source.
func NewProviderVersionsDataSource() datasource.DataSource {
	return &ProviderVersionsDataSource{}
}

func (d *ProviderVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_versions"
}

func (d *ProviderVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches available versions for a Terraform provider.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Full provider ID (namespace/name).",
				Computed:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "Provider namespace (e.g., hashicorp).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Provider name (e.g., aws).",
				Required:    true,
			},
			"versions": schema.ListNestedAttribute{
				Description: "List of available versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							Description: "Version string.",
							Computed:    true,
						},
						"published_at": schema.StringAttribute{
							Description: "Publication date.",
							Computed:    true,
						},
						"protocols": schema.ListAttribute{
							Description: "Supported protocol versions.",
							Computed:    true,
							ElementType: types.StringType,
						},
					},
				},
			},
		},
	}
}

func (d *ProviderVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*RegistryClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *RegistryClient, got: %T", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *ProviderVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProviderVersionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch versions from registry
	url := fmt.Sprintf("%s/providers/%s/%s/versions",
		d.client.BaseURL(),
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

	var apiResp VersionsAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Namespace.ValueString(), data.Name.ValueString()))

	// Convert versions to list
	versionType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"version":      types.StringType,
			"published_at": types.StringType,
			"protocols":    types.ListType{ElemType: types.StringType},
		},
	}

	var versionObjects []attr.Value
	for _, v := range apiResp.Versions {
		protocols := make([]attr.Value, len(v.Protocols))
		for i, p := range v.Protocols {
			protocols[i] = types.StringValue(p)
		}
		protocolsList, _ := types.ListValue(types.StringType, protocols)

		obj, _ := types.ObjectValue(versionType.AttrTypes, map[string]attr.Value{
			"version":      types.StringValue(v.Version),
			"published_at": types.StringValue(v.PublishedAt),
			"protocols":    protocolsList,
		})
		versionObjects = append(versionObjects, obj)
	}

	versionsList, diags := types.ListValue(versionType, versionObjects)
	resp.Diagnostics.Append(diags...)
	data.Versions = versionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// VersionsAPIResponse represents the registry API response for versions.
type VersionsAPIResponse struct {
	Versions []VersionDetailResponse `json:"versions"`
}

// VersionDetailResponse represents a detailed version.
type VersionDetailResponse struct {
	Version     string   `json:"version"`
	PublishedAt string   `json:"published_at"`
	Protocols   []string `json:"protocols"`
}
