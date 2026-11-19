// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package datasources

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
)

// Ensure ProviderDocsDataSource satisfies various datasource interfaces.
var _ datasource.DataSource = &ProviderDocsDataSource{}
var _ datasource.DataSourceWithConfigure = &ProviderDocsDataSource{}

// ProviderDocsDataSource defines the data source implementation.
type ProviderDocsDataSource struct {
	client *client.RegistryClient
}

// ProviderDocsDataSourceModel describes the data source data model.
type ProviderDocsDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	Namespace   types.String `tfsdk:"namespace"`
	Name        types.String `tfsdk:"name"`
	Version     types.String `tfsdk:"version"`
	DocType     types.String `tfsdk:"doc_type"`
	DocName     types.String `tfsdk:"doc_name"`
	Content     types.String `tfsdk:"content"`
	Title       types.String `tfsdk:"title"`
	Category    types.String `tfsdk:"category"`
	Subcategory types.String `tfsdk:"subcategory"`
	Language    types.String `tfsdk:"language"`
	Path        types.String `tfsdk:"path"`
}

// NewProviderDocsDataSource creates a new data source.
func NewProviderDocsDataSource() datasource.DataSource {
	return &ProviderDocsDataSource{}
}

func (d *ProviderDocsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_provider_docs"
}

func (d *ProviderDocsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Fetches documentation for a Terraform provider resource or data source.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Unique identifier for this documentation.",
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
			"version": schema.StringAttribute{
				Description: "Provider version (e.g., 5.0.0).",
				Required:    true,
			},
			"doc_type": schema.StringAttribute{
				Description: "Documentation type: resources, data-sources, or guides.",
				Required:    true,
			},
			"doc_name": schema.StringAttribute{
				Description: "Documentation name/slug (e.g., instance for aws_instance).",
				Required:    true,
			},
			"content": schema.StringAttribute{
				Description: "Documentation content in markdown format.",
				Computed:    true,
			},
			"title": schema.StringAttribute{
				Description: "Document title.",
				Computed:    true,
			},
			"category": schema.StringAttribute{
				Description: "Document category.",
				Computed:    true,
			},
			"subcategory": schema.StringAttribute{
				Description: "Document subcategory.",
				Computed:    true,
			},
			"language": schema.StringAttribute{
				Description: "Documentation language (typically hcl).",
				Computed:    true,
			},
			"path": schema.StringAttribute{
				Description: "Original path in the provider repository.",
				Computed:    true,
			},
		},
	}
}

func (d *ProviderDocsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ProviderDocsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data ProviderDocsDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 1: Get provider ID from v2 API
	providerURL := fmt.Sprintf("https://%s/v2/providers/%s/%s",
		d.client.Host,
		url.PathEscape(data.Namespace.ValueString()),
		url.PathEscape(data.Name.ValueString()),
	)

	providerResp, err := d.fetchJSON(ctx, providerURL)
	if err != nil {
		resp.Diagnostics.AddError("Provider Lookup Error", err.Error())
		return
	}

	providerData, ok := providerResp["data"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid provider response format")
		return
	}
	providerID := getStringOrEmpty(providerData, "id")
	if providerID == "" {
		resp.Diagnostics.AddError("API Error", "Provider response is missing an id")
		return
	}

	// Step 2: Find the version ID
	versionsURL := fmt.Sprintf("https://%s/v2/providers/%s/provider-versions",
		d.client.Host,
		url.PathEscape(providerID),
	)

	versionsResp, err := d.fetchJSON(ctx, versionsURL)
	if err != nil {
		resp.Diagnostics.AddError("Versions Lookup Error", err.Error())
		return
	}

	versionsData, ok := versionsResp["data"].([]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid versions response format")
		return
	}

	var versionID string
	targetVersion := data.Version.ValueString()
	for _, v := range versionsData {
		vMap, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		attrs, ok := vMap["attributes"].(map[string]interface{})
		if !ok {
			continue
		}
		if getStringOrEmpty(attrs, "version") == targetVersion {
			versionID = getStringOrEmpty(vMap, "id")
			break
		}
	}

	if versionID == "" {
		resp.Diagnostics.AddError("Version Not Found",
			fmt.Sprintf("Version %s not found for provider %s/%s",
				targetVersion, data.Namespace.ValueString(), data.Name.ValueString()))
		return
	}

	// Step 3: Find the doc by slug
	docType := data.DocType.ValueString()
	docName := data.DocName.ValueString()

	docsURL := fmt.Sprintf("https://%s/v2/provider-docs?filter[provider-version]=%s&filter[category]=%s&filter[slug]=%s",
		d.client.Host,
		url.QueryEscape(versionID),
		url.QueryEscape(docType),
		url.QueryEscape(docName),
	)

	docsResp, err := d.fetchJSON(ctx, docsURL)
	if err != nil {
		resp.Diagnostics.AddError("Docs Lookup Error", err.Error())
		return
	}

	docsData, ok := docsResp["data"].([]interface{})
	if !ok || len(docsData) == 0 {
		resp.Diagnostics.AddError("Doc Not Found",
			fmt.Sprintf("Documentation '%s' not found in %s for %s/%s@%s",
				docName, docType, data.Namespace.ValueString(), data.Name.ValueString(), targetVersion))
		return
	}

	docMeta, ok := docsData[0].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid docs response format")
		return
	}
	docID := getStringOrEmpty(docMeta, "id")
	if docID == "" {
		resp.Diagnostics.AddError("API Error", "Docs response is missing an id")
		return
	}

	// Step 4: Fetch the full doc content
	docURL := fmt.Sprintf("https://%s/v2/provider-docs/%s", d.client.Host, url.PathEscape(docID))

	docResp, err := d.fetchJSON(ctx, docURL)
	if err != nil {
		resp.Diagnostics.AddError("Doc Content Error", err.Error())
		return
	}

	docData, ok := docResp["data"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Invalid doc response format")
		return
	}
	attrs, ok := docData["attributes"].(map[string]interface{})
	if !ok {
		resp.Diagnostics.AddError("API Error", "Doc response is missing attributes")
		return
	}

	// Map response to model
	data.ID = types.StringValue(docID)
	data.Content = types.StringValue(getStringOrEmpty(attrs, "content"))
	data.Title = types.StringValue(getStringOrEmpty(attrs, "title"))
	data.Category = types.StringValue(getStringOrEmpty(attrs, "category"))
	data.Subcategory = types.StringValue(getStringOrEmpty(attrs, "subcategory"))
	data.Language = types.StringValue(getStringOrEmpty(attrs, "language"))
	data.Path = types.StringValue(getStringOrEmpty(attrs, "path"))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (d *ProviderDocsDataSource) fetchJSON(ctx context.Context, url string) (map[string]interface{}, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	if d.client.Token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+d.client.Token)
	}

	httpResp, err := d.client.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = httpResp.Body.Close() }()

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	var result map[string]interface{}
	if err := json.NewDecoder(httpResp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

func getStringOrEmpty(m map[string]interface{}, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}
