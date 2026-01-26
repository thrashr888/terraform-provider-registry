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

// Ensure PolicyLibraryVersionsDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &PolicyLibraryVersionsDataSource{}
var _ datasource.DataSourceWithConfigure = &PolicyLibraryVersionsDataSource{}

// PolicyLibraryVersionsDataSource defines the data source implementation.
type PolicyLibraryVersionsDataSource struct {
	client *client.RegistryClient
}

// PolicyLibraryVersionsDataSourceModel describes the data source data model.
type PolicyLibraryVersionsDataSourceModel struct {
	ID        types.String                        `tfsdk:"id"`
	Namespace types.String                        `tfsdk:"namespace"`
	Name      types.String                        `tfsdk:"name"`
	Versions  []PolicyLibraryVersionModel         `tfsdk:"versions"`
}

// PolicyLibraryVersionModel describes a single version.
type PolicyLibraryVersionModel struct {
	Version     types.String `tfsdk:"version"`
	Tag         types.String `tfsdk:"tag"`
	Description types.String `tfsdk:"description"`
	PublishedAt types.String `tfsdk:"published_at"`
	Downloads   types.Int64  `tfsdk:"downloads"`
}

// NewPolicyLibraryVersionsDataSource creates a new data source.
func NewPolicyLibraryVersionsDataSource() datasource.DataSource {
	return &PolicyLibraryVersionsDataSource{}
}

func (d *PolicyLibraryVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_library_versions"
}

func (d *PolicyLibraryVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists available versions for a Terraform policy library.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Full policy library ID (namespace/name).",
				Computed:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "Policy library namespace (e.g., hashicorp).",
				Required:    true,
			},
			"name": schema.StringAttribute{
				Description: "Policy library name (e.g., gcp-networking-terraform).",
				Required:    true,
			},
			"versions": schema.ListNestedAttribute{
				Description: "List of available versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							Description: "Version number.",
							Computed:    true,
						},
						"tag": schema.StringAttribute{
							Description: "Git tag for this version.",
							Computed:    true,
						},
						"description": schema.StringAttribute{
							Description: "Version description.",
							Computed:    true,
						},
						"published_at": schema.StringAttribute{
							Description: "Publication date.",
							Computed:    true,
						},
						"downloads": schema.Int64Attribute{
							Description: "Download count for this version.",
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *PolicyLibraryVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PolicyLibraryVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data PolicyLibraryVersionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	namespace := data.Namespace.ValueString()
	name := data.Name.ValueString()

	// Use v2 API with include=versions
	url := fmt.Sprintf("https://%s/v2/policies/%s/%s?include=versions",
		d.client.Host, namespace, name)

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

	var apiResp struct {
		Data struct {
			ID         string `json:"id"`
			Attributes struct {
				FullName  string `json:"full-name"`
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
			} `json:"attributes"`
		} `json:"data"`
		Included []struct {
			Type       string `json:"type"`
			ID         string `json:"id"`
			Attributes struct {
				Version     string `json:"version"`
				Tag         string `json:"tag"`
				Description string `json:"description"`
				PublishedAt string `json:"published-at"`
				Downloads   int64  `json:"downloads"`
			} `json:"attributes"`
		} `json:"included"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		resp.Diagnostics.AddError("JSON Decode Error", err.Error())
		return
	}

	data.ID = types.StringValue(fmt.Sprintf("%s/%s", namespace, name))
	data.Versions = make([]PolicyLibraryVersionModel, 0, len(apiResp.Included))

	for _, v := range apiResp.Included {
		if v.Type == "policy-library-versions" {
			data.Versions = append(data.Versions, PolicyLibraryVersionModel{
				Version:     types.StringValue(v.Attributes.Version),
				Tag:         types.StringValue(v.Attributes.Tag),
				Description: types.StringValue(v.Attributes.Description),
				PublishedAt: types.StringValue(v.Attributes.PublishedAt),
				Downloads:   types.Int64Value(v.Attributes.Downloads),
			})
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
