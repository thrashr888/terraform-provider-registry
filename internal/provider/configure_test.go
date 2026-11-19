// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/thrashr888/terraform-provider-registry/internal/client"
)

// TestConfigureSharesClient guards the terraform query crash: list resources
// need the client too, not just data sources and resources.
func TestConfigureSharesClient(t *testing.T) {
	ctx := context.Background()
	p := New("test")()

	var schemaResp provider.SchemaResponse
	p.Schema(ctx, provider.SchemaRequest{}, &schemaResp)
	typ := schemaResp.Schema.Type().TerraformType(ctx)
	cfg := tfsdk.Config{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, map[string]tftypes.Value{
		"host":  tftypes.NewValue(tftypes.String, nil),
		"token": tftypes.NewValue(tftypes.String, nil),
	})}

	var resp provider.ConfigureResponse
	p.Configure(ctx, provider.ConfigureRequest{Config: cfg}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("configure: %v", resp.Diagnostics)
	}
	for name, data := range map[string]any{
		"DataSourceData":   resp.DataSourceData,
		"ResourceData":     resp.ResourceData,
		"ListResourceData": resp.ListResourceData,
	} {
		c, ok := data.(*client.RegistryClient)
		if !ok || c == nil {
			t.Errorf("%s = %T, want *client.RegistryClient", name, data)
			continue
		}
		if c.Host != "registry.terraform.io" {
			t.Errorf("%s host = %q", name, c.Host)
		}
	}
}

// TestProviderRegistersFunctions makes sure version_satisfies is served.
func TestProviderRegistersFunctions(t *testing.T) {
	p, ok := New("test")().(provider.ProviderWithFunctions)
	if !ok {
		t.Fatal("provider does not implement ProviderWithFunctions")
	}
	if n := len(p.Functions(context.Background())); n != 1 {
		t.Fatalf("got %d functions, want 1", n)
	}
}
