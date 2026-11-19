// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package listresources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework/list"
	"github.com/hashicorp/terraform-plugin-framework/list/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
	"github.com/thrashr888/terraform-provider-registry/internal/resources"
)

// Ensure ModuleListResource satisfies various list resource interfaces.
var _ list.ListResource = &ModuleListResource{}
var _ list.ListResourceWithConfigure = &ModuleListResource{}

// ModuleListResource defines the list resource implementation.
type ModuleListResource struct {
	client *client.RegistryClient
}

// ModuleListConfigModel describes the config model for the list block.
type ModuleListConfigModel struct {
	Query          types.String `tfsdk:"query"`
	Namespace      types.String `tfsdk:"namespace"`
	TargetProvider types.String `tfsdk:"target_provider"`
}

// NewModuleListResource creates a new list resource.
func NewModuleListResource() list.ListResource {
	return &ModuleListResource{}
}

func (r *ModuleListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_module"
}

func (r *ModuleListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists or searches Terraform modules in the registry. Set `query` to search by text, `namespace` to list a namespace, or both.",
		Attributes: map[string]schema.Attribute{
			"query": schema.StringAttribute{
				Description: "Text to search for (e.g., vpc). At least one of `query` or `namespace` is required.",
				Optional:    true,
			},
			"namespace": schema.StringAttribute{
				Description: "Module namespace to filter by (e.g., terraform-aws-modules). At least one of `query` or `namespace` is required.",
				Optional:    true,
			},
			"target_provider": schema.StringAttribute{
				Description: "Target provider to filter by (e.g., aws).",
				Optional:    true,
			},
		},
	}
}

func (r *ModuleListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	c, ok := req.ProviderData.(*client.RegistryClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected List Resource Configure Type",
			fmt.Sprintf("Expected *client.RegistryClient, got: %T", req.ProviderData),
		)
		return
	}

	r.client = c
}

func (r *ModuleListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var config ModuleListConfigModel

	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	if config.Query.ValueString() == "" && config.Namespace.ValueString() == "" {
		diags.AddError("Missing Search Criteria", "Set at least one of query or namespace.")
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	var modules []ModuleAPIItem
	var err error
	if config.Query.ValueString() != "" {
		modules, err = r.searchModules(ctx, config.Query.ValueString(), config.Namespace.ValueString(), config.TargetProvider.ValueString(), int(req.Limit))
	} else {
		modules, err = r.fetchModules(ctx, config.Namespace.ValueString(), config.TargetProvider.ValueString(), int(req.Limit))
	}
	if err != nil {
		diags.AddError("Registry API Error", err.Error())
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// Stream results
	resp.Results = func(push func(list.ListResult) bool) {
		for _, m := range modules {
			result := req.NewListResult(ctx)
			result.DisplayName = fmt.Sprintf("%s/%s/%s", m.Namespace, m.Name, m.TargetProvider)

			// Set identity
			identity := resources.ModuleIdentityModel{
				Namespace:      types.StringValue(m.Namespace),
				Name:           types.StringValue(m.Name),
				TargetProvider: types.StringValue(m.TargetProvider),
			}
			result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

			// Set resource if requested
			if req.IncludeResource {
				resourceData := resources.ModuleResourceModel{
					ID:             types.StringValue(fmt.Sprintf("%s/%s/%s", m.Namespace, m.Name, m.TargetProvider)),
					Namespace:      types.StringValue(m.Namespace),
					Name:           types.StringValue(m.Name),
					TargetProvider: types.StringValue(m.TargetProvider),
					Description:    types.StringValue(m.Description),
					Source:         types.StringValue(m.Source),
					Verified:       types.BoolValue(m.Verified),
					Downloads:      types.Int64Value(m.Downloads),
				}
				result.Diagnostics.Append(result.Resource.Set(ctx, resourceData)...)
			}

			if !push(result) {
				return
			}
		}
	}
}

// ModuleAPIItem represents a module from the v2 API.
type ModuleAPIItem struct {
	Namespace      string
	Name           string
	TargetProvider string
	Description    string
	Source         string
	Verified       bool
	Downloads      int64
}

func (r *ModuleListResource) fetchModules(ctx context.Context, namespace, targetProvider string, limit int) ([]ModuleAPIItem, error) {
	if limit == 0 {
		limit = 100
	}

	// Use v2 API for listing modules
	query := url.Values{}
	query.Set("filter[namespace]", namespace)
	query.Set("page[size]", strconv.Itoa(limit))
	if targetProvider != "" {
		query.Set("filter[provider]", targetProvider)
	}
	reqURL := fmt.Sprintf("https://%s/v2/modules?%s", r.client.Host, query.Encode())

	httpReq, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, err
	}

	if r.client.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+r.client.Token)
	}

	httpResp, err := r.client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var apiResp struct {
		Data []struct {
			ID         string `json:"id"`
			Attributes struct {
				Name         string `json:"name"`
				Namespace    string `json:"namespace"`
				ProviderName string `json:"provider-name"`
				Description  string `json:"description"`
				Source       string `json:"source"`
				Verified     bool   `json:"verified"`
				Downloads    int64  `json:"downloads"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	modules := make([]ModuleAPIItem, 0, len(apiResp.Data))
	for _, m := range apiResp.Data {
		modules = append(modules, ModuleAPIItem{
			Namespace:      m.Attributes.Namespace,
			Name:           m.Attributes.Name,
			TargetProvider: m.Attributes.ProviderName,
			Description:    m.Attributes.Description,
			Source:         m.Attributes.Source,
			Verified:       m.Attributes.Verified,
			Downloads:      m.Attributes.Downloads,
		})
	}

	return modules, nil
}

// searchModules runs a text search with the v1 modules search endpoint.
func (r *ModuleListResource) searchModules(ctx context.Context, q, namespace, targetProvider string, limit int) ([]ModuleAPIItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	query := url.Values{}
	query.Set("q", q)
	query.Set("limit", strconv.Itoa(limit))
	if namespace != "" {
		query.Set("namespace", namespace)
	}
	if targetProvider != "" {
		query.Set("provider", targetProvider)
	}

	var apiResp struct {
		Modules []struct {
			Namespace   string `json:"namespace"`
			Name        string `json:"name"`
			Provider    string `json:"provider"`
			Description string `json:"description"`
			Source      string `json:"source"`
			Verified    bool   `json:"verified"`
			Downloads   int64  `json:"downloads"`
		} `json:"modules"`
	}
	if err := getJSON(ctx, r.client, fmt.Sprintf("%s/modules/search?%s", r.client.BaseURL(), query.Encode()), &apiResp); err != nil {
		return nil, err
	}

	modules := make([]ModuleAPIItem, 0, len(apiResp.Modules))
	for _, m := range apiResp.Modules {
		modules = append(modules, ModuleAPIItem{
			Namespace:      m.Namespace,
			Name:           m.Name,
			TargetProvider: m.Provider,
			Description:    m.Description,
			Source:         m.Source,
			Verified:       m.Verified,
			Downloads:      m.Downloads,
		})
	}
	return modules, nil
}
