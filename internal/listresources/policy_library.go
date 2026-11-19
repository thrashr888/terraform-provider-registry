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

// Ensure PolicyLibraryListResource satisfies various list resource interfaces.
var _ list.ListResource = &PolicyLibraryListResource{}
var _ list.ListResourceWithConfigure = &PolicyLibraryListResource{}

// PolicyLibraryListResource defines the list resource implementation.
type PolicyLibraryListResource struct {
	client *client.RegistryClient
}

// PolicyLibraryListConfigModel describes the config model for the list block.
type PolicyLibraryListConfigModel struct {
	Namespace types.String `tfsdk:"namespace"`
}

// NewPolicyLibraryListResource creates a new list resource.
func NewPolicyLibraryListResource() list.ListResource {
	return &PolicyLibraryListResource{}
}

func (r *PolicyLibraryListResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_policy_library"
}

func (r *PolicyLibraryListResource) ListResourceConfigSchema(ctx context.Context, req list.ListResourceSchemaRequest, resp *list.ListResourceSchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Lists Terraform policy libraries from the registry by namespace.",
		Attributes: map[string]schema.Attribute{
			"namespace": schema.StringAttribute{
				Description: "Policy library namespace to filter by (e.g., hashicorp).",
				Required:    true,
			},
		},
	}
}

func (r *PolicyLibraryListResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *PolicyLibraryListResource) List(ctx context.Context, req list.ListRequest, resp *list.ListResultsStream) {
	var config PolicyLibraryListConfigModel

	diags := req.Config.Get(ctx, &config)
	if diags.HasError() {
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// Fetch policy libraries from registry v2 API
	libraries, err := r.fetchPolicyLibraries(ctx, config.Namespace.ValueString(), int(req.Limit))
	if err != nil {
		diags.AddError("Registry API Error", err.Error())
		resp.Results = list.ListResultsStreamDiagnostics(diags)
		return
	}

	// Stream results
	resp.Results = func(push func(list.ListResult) bool) {
		for _, lib := range libraries {
			result := req.NewListResult(ctx)
			result.DisplayName = fmt.Sprintf("%s/%s", lib.Namespace, lib.Name)

			// Set identity
			identity := resources.PolicyLibraryIdentityModel{
				Namespace: types.StringValue(lib.Namespace),
				Name:      types.StringValue(lib.Name),
			}
			result.Diagnostics.Append(result.Identity.Set(ctx, identity)...)

			// Set resource if requested
			if req.IncludeResource {
				resourceData := resources.PolicyLibraryResourceModel{
					ID:        types.StringValue(fmt.Sprintf("%s/%s", lib.Namespace, lib.Name)),
					Namespace: types.StringValue(lib.Namespace),
					Name:      types.StringValue(lib.Name),
					Title:     types.StringValue(lib.Title),
					Source:    types.StringValue(lib.Source),
					Verified:  types.BoolValue(lib.Verified),
					Downloads: types.Int64Value(lib.Downloads),
				}
				result.Diagnostics.Append(result.Resource.Set(ctx, resourceData)...)
			}

			if !push(result) {
				return
			}
		}
	}
}

// PolicyLibraryAPIItem represents a policy library from the v2 API.
type PolicyLibraryAPIItem struct {
	Namespace string
	Name      string
	Title     string
	Source    string
	Verified  bool
	Downloads int64
}

func (r *PolicyLibraryListResource) fetchPolicyLibraries(ctx context.Context, namespace string, limit int) ([]PolicyLibraryAPIItem, error) {
	if limit == 0 {
		limit = 100
	}

	// Use v2 API for listing policy libraries
	query := url.Values{}
	query.Set("filter[namespace]", namespace)
	query.Set("page[size]", strconv.Itoa(limit))
	reqURL := fmt.Sprintf("https://%s/v2/policies?%s", r.client.Host, query.Encode())

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
				Name      string `json:"name"`
				Namespace string `json:"namespace"`
				Title     string `json:"title"`
				Source    string `json:"source"`
				Verified  bool   `json:"verified"`
				Downloads int64  `json:"downloads"`
			} `json:"attributes"`
		} `json:"data"`
	}

	if err := json.NewDecoder(httpResp.Body).Decode(&apiResp); err != nil {
		return nil, err
	}

	libraries := make([]PolicyLibraryAPIItem, 0, len(apiResp.Data))
	for _, lib := range apiResp.Data {
		libraries = append(libraries, PolicyLibraryAPIItem{
			Namespace: lib.Attributes.Namespace,
			Name:      lib.Attributes.Name,
			Title:     lib.Attributes.Title,
			Source:    lib.Attributes.Source,
			Verified:  lib.Attributes.Verified,
			Downloads: lib.Attributes.Downloads,
		})
	}

	return libraries, nil
}
