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

// Ensure ProviderDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &ProviderDataSource{}
var _ datasource.DataSourceWithConfigure = &ProviderDataSource{}

// ProviderDataSource defines the data source implementation.
type ProviderDataSource struct {
	client *client.RegistryClient
}

// ProviderDataSourceModel describes the data source data model.
type ProviderDataSourceModel struct {
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

// NewProviderDataSource creates a new data source.
func NewProviderDataSource() datasource.DataSource {
	return &ProviderDataSource{}
}

func (d *ProviderDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

func (d *ProviderDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Terraform provider from the registry.",

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

func (d *ProviderDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProviderDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProviderDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch provider from registry
	url := fmt.Sprintf("%s/providers/%s/%s",
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
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Registry API Error",
			fmt.Sprintf("Unexpected status code: %d", httpResp.StatusCode),
		)
		return
	}

	var apiResp ProviderAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Namespace.ValueString(), data.Name.ValueString()))
	data.Description = types.StringValue(apiResp.Description)
	data.Source = types.StringValue(apiResp.Source)
	data.Tier = types.StringValue(apiResp.Tier)
	data.PublishedAt = types.StringValue(apiResp.PublishedAt)
	data.Downloads = types.Int64Value(apiResp.Downloads)
	// Use the version field from the response (it's the latest)
	data.LatestVersion = types.StringValue(apiResp.Version)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ProviderAPIResponse represents the registry API response for a provider.
type ProviderAPIResponse struct {
	ID          string   `json:"id"`
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Source      string   `json:"source"`
	Tier        string   `json:"tier"`
	PublishedAt string   `json:"published_at"`
	Downloads   int64    `json:"downloads"`
	Versions    []string `json:"versions"`
}
