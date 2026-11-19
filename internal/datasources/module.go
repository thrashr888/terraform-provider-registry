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

// Ensure ModuleDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &ModuleDataSource{}
var _ datasource.DataSourceWithConfigure = &ModuleDataSource{}

// ModuleDataSource defines the data source implementation.
type ModuleDataSource struct {
	client *client.RegistryClient
}

// ModuleDataSourceModel describes the data source data model.
type ModuleDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	TargetProvider types.String `tfsdk:"target_provider"`
	Description    types.String `tfsdk:"description"`
	Source         types.String `tfsdk:"source"`
	Verified       types.Bool   `tfsdk:"verified"`
	PublishedAt    types.String `tfsdk:"published_at"`
	Downloads      types.Int64  `tfsdk:"downloads"`
	LatestVersion  types.String `tfsdk:"latest_version"`
}

// NewModuleDataSource creates a new data source.
func NewModuleDataSource() datasource.DataSource {
	return &ModuleDataSource{}
}

func (d *ModuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module"
}

func (d *ModuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches information about a Terraform module from the registry.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Full module ID (namespace/name/provider).",
				Computed:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "Module namespace (e.g., hashicorp).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Module name (e.g., consul).",
				Required:    true,
			},
			"target_provider": schema.StringAttribute{
				Description: "Target provider (e.g., aws).",
				Required:    true,
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

func (d *ModuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModuleDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch module from registry
	url := fmt.Sprintf("%s/modules/%s/%s/%s",
		d.client.BaseURL(),
		data.Namespace.ValueString(),
		data.Name.ValueString(),
		data.TargetProvider.ValueString(),
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

	var apiResp ModuleAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s/%s",
		data.Namespace.ValueString(),
		data.Name.ValueString(),
		data.TargetProvider.ValueString(),
	))
	data.Description = types.StringValue(apiResp.Description)
	data.Source = types.StringValue(apiResp.Source)
	data.Verified = types.BoolValue(apiResp.Verified)
	data.PublishedAt = types.StringValue(apiResp.PublishedAt)
	data.Downloads = types.Int64Value(apiResp.Downloads)
	if len(apiResp.Versions) > 0 {
		data.LatestVersion = types.StringValue(apiResp.Versions[0])
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ModuleAPIResponse represents the registry API response for a module.
type ModuleAPIResponse struct {
	ID          string   `json:"id"`
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	Provider    string   `json:"provider"`
	Description string   `json:"description"`
	Source      string   `json:"source"`
	Verified    bool     `json:"verified"`
	PublishedAt string   `json:"published_at"`
	Downloads   int64    `json:"downloads"`
	Versions    []string `json:"versions"`
}
