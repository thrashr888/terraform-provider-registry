// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
	"github.com/thrashr888/terraform-provider-registry/internal/datasources"
	"github.com/thrashr888/terraform-provider-registry/internal/functions"
	"github.com/thrashr888/terraform-provider-registry/internal/listresources"
	"github.com/thrashr888/terraform-provider-registry/internal/resources"
)

// Ensure RegistryProvider satisfies various provider interfaces.
var _ provider.Provider = &RegistryProvider{}
var _ provider.ProviderWithListResources = &RegistryProvider{}
var _ provider.ProviderWithFunctions = &RegistryProvider{}

// RegistryProvider defines the provider implementation.
type RegistryProvider struct {
	version string
}

// RegistryProviderModel describes the provider data model.
type RegistryProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

// New returns a new provider factory function.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &RegistryProvider{
			version: version,
		}
	}
}

func (p *RegistryProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "registry"
	resp.Version = p.version
}

func (p *RegistryProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Provider for querying the Terraform Registry API.",
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Description: "The Terraform Registry host. Defaults to registry.terraform.io.",
				Optional:    true,
			},
			"token": schema.StringAttribute{
				Description: "API token for private registry authentication.",
				Optional:    true,
				Sensitive:   true,
			},
		},
	}
}

func (p *RegistryProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config RegistryProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	if resp.Diagnostics.HasError() {
		return
	}

	// Default host
	host := "registry.terraform.io"
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	// Create HTTP client
	registryClient := client.New(host, config.Token.ValueString(), p.version)

	// Make client available to data sources, resources, and list resources
	resp.DataSourceData = registryClient
	resp.ResourceData = registryClient
	resp.ListResourceData = registryClient
}

func (p *RegistryProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		resources.NewProviderResource,
		resources.NewModuleResource,
		resources.NewPolicyLibraryResource,
	}
}

func (p *RegistryProvider) ListResources(ctx context.Context) []func() list.ListResource {
	return []func() list.ListResource{
		listresources.NewProviderListResource,
		listresources.NewModuleListResource,
		listresources.NewPolicyLibraryListResource,
	}
}

func (p *RegistryProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		datasources.NewProviderDataSource,
		datasources.NewProviderVersionsDataSource,
		datasources.NewModuleDataSource,
		datasources.NewModuleVersionsDataSource,
		datasources.NewSearchDataSource,
		datasources.NewProviderDocsDataSource,
		datasources.NewPolicyLibraryDataSource,
		datasources.NewPolicyLibraryVersionsDataSource,
	}
}

func (p *RegistryProvider) Functions(ctx context.Context) []func() function.Function {
	return []func() function.Function{
		functions.NewVersionSatisfiesFunction,
	}
}
