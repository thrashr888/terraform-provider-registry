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
)

// Ensure ProviderListResource satisfies various list resource interfaces.
var _ list.ListResource = &ProviderListResource{}
var _ list.ListResourceWithConfigure = &ProviderListResource{}

// ProviderListResource defines the list resource implementation.
type ProviderListResource struct {
	client *client.RegistryClient
}

// ProviderListConfigModel describes the config model for the list block.
type ProviderListConfigModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Tier      types.String `tfsdk:"tier"`
}

// ProviderIdentityModel describes the identity of a listed provider.
type ProviderIdentityModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
}

// NewProviderListResource creates a new list resource.
func NewProviderListResource() list.ListResource {
	return &ProviderListResource{}
}

func (r *ProviderListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider"
}

func (r *ProviderListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Terraform providers from the registry by namespace.",
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				Description: "Provider namespace to filter by (e.g., hashicorp).",
				Required:    true,
			},
			"tier": schema.StringAttribute{
				Description: "Provider tier to filter by (official, partner, community).",
				Optional:    true,
			},
		},
	}
}

func (r *ProviderListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *ProviderListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var config ProviderListConfigModel

	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// Fetch providers from registry v2 API
	providers, err := r.fetchProviders(ctx, config.Namespace.ValueString(), config.Tier.ValueString(), int(req.Limit))
	if err != nil {
		diags.AddError("Registry API Error", err.Error())
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// Stream results
	resp.Results = func(push func(list.ListResult) bool) {
		for _, p := range providers {
			result := req.NewListResult(ctx)
			result.DisplayName = fmt.Sprintf("%s/%s", p.Namespace, p.Name)

			// Set identity
			identity := ProviderIdentityModel{
				Namespace: types.StringValue(p.Namespace),
				Name:      types.StringValue(p.Name),
			}
			result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

			// Set resource if requested
			if req.IncludeResource {
				resourceData := map[string]interface{}{
					"id":             fmt.Sprintf("%s/%s", p.Namespace, p.Name),
					"namespace":      p.Namespace,
					"name":           p.Name,
					"description":    p.Description,
					"source":         p.Source,
					"tier":           p.Tier,
					"published_at":   "",
					"downloads":      p.Downloads,
					"latest_version": "",
				}
				result.Diagnostics.Append(result.Resource.Set(ctx, resourceData)...)
			}

			if !push(result) {
				return
			}
		}
	}
}

// ProviderAPIItem represents a provider from the v2 API.
type ProviderAPIItem struct {
	Namespace   string
	Name        string
	Description string
	Source      string
	Tier        string
	Downloads   int64
}

func (r *ProviderListResource) fetchProviders(ctx context.Context, namespace, tier string, limit int) ([]ProviderAPIItem, error) {
	if limit == 0 {
		limit = 100
	}

	// Use v2 API for listing providers
	query := url.Values{}
	query.Set("filter[namespace]", namespace)
	query.Set("page[size]", strconv.Itoa(limit))
	if tier != "" {
		query.Set("filter[tier]", tier)
	}
	reqURL := fmt.Sprintf("https://%s/v2/providers?%s", r.client.Host, query.Encode())

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
				Name        string `json:"name"`
				Namespace   string `json:"namespace"`
				Description string `json:"description"`
				Source      string `json:"source"`
				Tier        string `json:"tier"`
				Downloads   int64  `json:"downloads"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	providers := make([]ProviderAPIItem, 0, len(apiResp.Data))
	for _, p := range apiResp.Data {
		providers = append(providers, ProviderAPIItem{
			Namespace:   p.Attributes.Namespace,
			Name:        p.Attributes.Name,
			Description: p.Attributes.Description,
			Source:      p.Attributes.Source,
			Tier:        p.Attributes.Tier,
			Downloads:   p.Attributes.Downloads,
		})
	}

	return providers, nil
}
