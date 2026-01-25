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

	"github.com/thrashr888/terraform-provider-registry/internal/client"
)

// Ensure ModuleVersionsDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &ModuleVersionsDataSource{}
var _ datasource.DataSourceWithConfigure = &ModuleVersionsDataSource{}

// ModuleVersionsDataSource defines the data source implementation.
type ModuleVersionsDataSource struct {
	client *client.RegistryClient
}

// ModuleVersionsDataSourceModel describes the data source data model.
type ModuleVersionsDataSourceModel struct {
	ID             types.String `tfsdk:"id"`
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	TargetProvider types.String `tfsdk:"target_provider"`
	Versions       types.List   `tfsdk:"versions"`
}

// NewModuleVersionsDataSource creates a new data source.
func NewModuleVersionsDataSource() datasource.DataSource {
	return &ModuleVersionsDataSource{}
}

func (d *ModuleVersionsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module_versions"
}

func (d *ModuleVersionsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches available versions for a Terraform module.",

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
			"versions": schema.ListNestedAttribute{
				Description: "List of available versions.",
				Computed:    true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"version": schema.StringAttribute{
							Description: "Version string.",
							Computed:    true,
						},
						"submodules": schema.ListNestedAttribute{
							Description: "Submodules in this version.",
							Computed:    true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"path": schema.StringAttribute{
										Description: "Submodule path.",
										Computed:    true,
									},
								},
							},
						},
						"root": schema.SingleNestedAttribute{
							Description: "Root module information.",
							Computed:    true,
							Attributes: map[string]schema.Attribute{
								"providers": schema.ListNestedAttribute{
									Description: "Required providers.",
									Computed:    true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: map[string]schema.Attribute{
											"name": schema.StringAttribute{
												Description: "Provider name.",
												Computed:    true,
											},
											"namespace": schema.StringAttribute{
												Description: "Provider namespace.",
												Computed:    true,
											},
											"source": schema.StringAttribute{
												Description: "Provider source.",
												Computed:    true,
											},
											"version": schema.StringAttribute{
												Description: "Version constraint.",
												Computed:    true,
											},
										},
									},
								},
								"dependencies": schema.ListAttribute{
									Description: "Module dependencies.",
									Computed:    true,
									ElementType: types.StringType,
								},
							},
						},
					},
				},
			},
		},
	}
}

func (d *ModuleVersionsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ModuleVersionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ModuleVersionsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Fetch versions from registry
	url := fmt.Sprintf("%s/modules/%s/%s/%s/versions",
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
	defer httpResp.Body.Close()

	if httpResp.StatusCode != http.StatusOK {
		resp.Diagnostics.AddError(
			"Registry API Error",
			fmt.Sprintf("Unexpected status code: %d", httpResp.StatusCode),
		)
		return
	}

	var apiResp ModuleVersionsAPIResponse
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

	// Build version list
	providerType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":      types.StringType,
			"namespace": types.StringType,
			"source":    types.StringType,
			"version":   types.StringType,
		},
	}

	submoduleType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"path": types.StringType,
		},
	}

	rootType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"providers":    types.ListType{ElemType: providerType},
			"dependencies": types.ListType{ElemType: types.StringType},
		},
	}

	versionType := types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"version":    types.StringType,
			"submodules": types.ListType{ElemType: submoduleType},
			"root":       rootType,
		},
	}

	var versionObjects []attr.Value
	for _, mod := range apiResp.Modules {
		for _, v := range mod.Versions {
			// Build submodules list
			var submoduleObjects []attr.Value
			for _, sm := range v.Submodules {
				obj, _ := types.ObjectValue(submoduleType.AttrTypes, map[string]attr.Value{
					"path": types.StringValue(sm.Path),
				})
				submoduleObjects = append(submoduleObjects, obj)
			}
			submodulesList, _ := types.ListValue(submoduleType, submoduleObjects)

			// Build providers list
			var providerObjects []attr.Value
			for _, p := range v.Root.Providers {
				obj, _ := types.ObjectValue(providerType.AttrTypes, map[string]attr.Value{
					"name":      types.StringValue(p.Name),
					"namespace": types.StringValue(p.Namespace),
					"source":    types.StringValue(p.Source),
					"version":   types.StringValue(p.Version),
				})
				providerObjects = append(providerObjects, obj)
			}
			providersList, _ := types.ListValue(providerType, providerObjects)

			// Build dependencies list
			var depValues []attr.Value
			for _, dep := range v.Root.Dependencies {
				depValues = append(depValues, types.StringValue(dep))
			}
			depsList, _ := types.ListValue(types.StringType, depValues)

			// Build root object
			rootObj, _ := types.ObjectValue(rootType.AttrTypes, map[string]attr.Value{
				"providers":    providersList,
				"dependencies": depsList,
			})

			// Build version object
			versionObj, _ := types.ObjectValue(versionType.AttrTypes, map[string]attr.Value{
				"version":    types.StringValue(v.Version),
				"submodules": submodulesList,
				"root":       rootObj,
			})
			versionObjects = append(versionObjects, versionObj)
		}
	}

	versionsList, diags := types.ListValue(versionType, versionObjects)
	resp.Diagnostics.Append(diags...)
	data.Versions = versionsList

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// ModuleVersionsAPIResponse represents the registry API response for module versions.
type ModuleVersionsAPIResponse struct {
	Modules []ModuleVersionsModule `json:"modules"`
}

// ModuleVersionsModule represents a module in the versions response.
type ModuleVersionsModule struct {
	Source   string                       `json:"source"`
	Versions []ModuleVersionDetailResponse `json:"versions"`
}

// ModuleVersionDetailResponse represents a detailed module version.
type ModuleVersionDetailResponse struct {
	Version    string                    `json:"version"`
	Submodules []ModuleSubmoduleResponse `json:"submodules"`
	Root       ModuleRootResponse        `json:"root"`
}

// ModuleSubmoduleResponse represents a submodule.
type ModuleSubmoduleResponse struct {
	Path string `json:"path"`
}

// ModuleRootResponse represents the root module.
type ModuleRootResponse struct {
	Providers    []ModuleProviderResponse `json:"providers"`
	Dependencies []string                 `json:"dependencies"`
}

// ModuleProviderResponse represents a provider dependency.
type ModuleProviderResponse struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	Source    string `json:"source"`
	Version   string `json:"version"`
}
