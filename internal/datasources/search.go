// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure SearchDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &SearchDataSource{}
var _ datasource.DataSourceWithConfigure = &SearchDataSource{}

// SearchDataSource defines the data source implementation.
type SearchDataSource struct {
	client *RegistryClient
}

// SearchDataSourceModel describes the data source data model.
type SearchDataSourceModel struct {
	ID         types.String `tfsdk:"id"`
	Query      types.String `tfsdk:"query"`
	Type       types.String `tfsdk:"type"`
	Namespace  types.String `tfsdk:"namespace"`
	Tier       types.String `tfsdk:"tier"`
	Limit      types.Int64  `tfsdk:"limit"`
	Offset     types.Int64  `tfsdk:"offset"`
	Results    types.List   `tfsdk:"results"`
	TotalCount types.Int64  `tfsdk:"total_count"`
}

// NewSearchDataSource creates a new data source.
func NewSearchDataSource() datasource.DataSource {
	return &SearchDataSource{}
}

func (d *SearchDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_search"
}

func (d *SearchDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Search the Terraform Registry for providers or modules.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Search query ID.",
				Computed:    true,
			},
			"query": schema.StringAttribute{
				Description: "Search query string.",
				Required:    true,
			},
			"type": schema.StringAttribute{
				Description: "Search type: 'providers' or 'modules'.",
				Required:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "Filter by namespace.",
				Optional:    true,
			},
			"tier": schema.StringAttribute{
				Description: "Filter by tier (providers only).",
				Optional:    true,
			},
			"limit": schema.Int64Attribute{
				Description: "Maximum results to return (default: 20).",
				Optional:    true,
			},
			"offset": schema.Int64Attribute{
				Description: "Pagination offset.",
				Optional:    true,
			},
			"results": schema.ListNestedAttribute{
				Description: "Search results.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Description: "Result ID.",
							Computed:    true,
						},
						"namespace": schema.StringAttribute{
							Description: "Namespace.",
							Computed:    true,
						},
						"name": schema.StringAttribute{
							Description: "Name.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Description.",
							Computed:    true,
						},
						"tier": schema.StringAttribute{
							Description: "Tier (providers only).",
							Computed:    true,
						},
						"downloads": schema.Int64Attribute{
							Description: "Download count.",
							Computed:    true,
						},
					},
				},
			},
			"total_count": schema.Int64Attribute{
				Description: "Total number of matching results.",
				Computed:    true,
			},
		},
	}
}

func (d *SearchDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *SearchDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SearchDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Build search URL
	searchURL := fmt.Sprintf("%s/%s", d.client.BaseURL(), data.Type.ValueString())
	params := url.Values{}
	params.Set("q", data.Query.ValueString())

	if !data.Namespace.IsNull() {
		params.Set("namespace", data.Namespace.ValueString())
	}
	if !data.Tier.IsNull() {
		params.Set("tier", data.Tier.ValueString())
	}
	if !data.Limit.IsNull() {
		params.Set("limit", fmt.Sprintf("%d", data.Limit.ValueInt64()))
	}
	if !data.Offset.IsNull() {
		params.Set("offset", fmt.Sprintf("%d", data.Offset.ValueInt64()))
	}

	fullURL := searchURL + "?" + params.Encode()

	httpReq, err := http.NewRequestWithContext(ctx, "GET", fullURL, nil)
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

	var apiResp SearchAPIResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", err.Error())
		return
	}

	// Map response to model
	data.ID = types.StringValue(fmt.Sprintf("%s/%s", data.Type.ValueString(), data.Query.ValueString()))
	data.TotalCount = types.Int64Value(int64(len(apiResp.Providers) + len(apiResp.Modules)))

	// Convert results to list
	resultType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"id":          types.StringType,
			"namespace":   types.StringType,
			"name":        types.StringType,
			"description": types.StringType,
			"tier":        types.StringType,
			"downloads":   types.Int64Type,
		},
	}

	var resultObjects []attr.Value

	// Add providers
	for _, p := range apiResp.Providers {
		obj, _ := types.ObjectValue(resultType.AttrTypes, map[string]attr.Value{
			"id":          types.StringValue(fmt.Sprintf("%s/%s", p.Namespace, p.Name)),
			"namespace":   types.StringValue(p.Namespace),
			"name":        types.StringValue(p.Name),
			"description": types.StringValue(p.Description),
			"tier":        types.StringValue(p.Tier),
			"downloads":   types.Int64Value(p.Downloads),
		})
		resultObjects = append(resultObjects, obj)
	}

	// Add modules
	for _, m := range apiResp.Modules {
		obj, _ := types.ObjectValue(resultType.AttrTypes, map[string]attr.Value{
			"id":          types.StringValue(fmt.Sprintf("%s/%s", m.Namespace, m.Name)),
			"namespace":   types.StringValue(m.Namespace),
			"name":        types.StringValue(m.Name),
			"description": types.StringValue(m.Description),
			"tier":        types.StringValue(""),
			"downloads":   types.Int64Value(m.Downloads),
		})
		resultObjects = append(resultObjects, obj)
	}

	resultsList, diags := types.ListValue(resultType, resultObjects)
	resp.Diagnostics.Append(diags...)
	data.Results = resultsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// SearchAPIResponse represents the registry search API response.
type SearchAPIResponse struct {
	Providers []SearchResultProvider `json:"providers"`
	Modules   []SearchResultModule   `json:"modules"`
}

// SearchResultProvider represents a provider search result.
type SearchResultProvider struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Tier        string `json:"tier"`
	Downloads   int64  `json:"downloads"`
}

// SearchResultModule represents a module search result.
type SearchResultModule struct {
	ID          string `json:"id"`
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Downloads   int64  `json:"downloads"`
}
