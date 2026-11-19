// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"reflect"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

func TestParseImportID(t *testing.T) {
	cases := []struct {
		name    string
		id      string
		fields  []string
		want    []string
		wantErr bool
	}{
		{"provider", "hashicorp/aws", []string{"namespace", "name"}, []string{"hashicorp", "aws"}, false},
		{"module", "terraform-aws-modules/vpc/aws", []string{"namespace", "name", "target_provider"}, []string{"terraform-aws-modules", "vpc", "aws"}, false},
		{"policy library", "hashicorp/CIS-Policy-Set-for-AWS-Terraform", []string{"namespace", "name"}, []string{"hashicorp", "CIS-Policy-Set-for-AWS-Terraform"}, false},
		{"surrounding space", " hashicorp/aws ", []string{"namespace", "name"}, []string{"hashicorp", "aws"}, false},
		{"too few parts", "hashicorp", []string{"namespace", "name"}, nil, true},
		{"too many parts", "hashicorp/aws/extra", []string{"namespace", "name"}, nil, true},
		{"module missing provider", "terraform-aws-modules/vpc", []string{"namespace", "name", "target_provider"}, nil, true},
		{"empty part", "hashicorp/", []string{"namespace", "name"}, nil, true},
		{"empty", "", []string{"namespace", "name"}, nil, true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseImportID(tc.id, tc.fields...)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err = %v, wantErr %v", err, tc.wantErr)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("got %v, want %v", got, tc.want)
			}
		})
	}
}

// TestImportStateSetsLookupAttributes runs each resource's ImportState and
// checks that the attributes Read needs are set from the import ID.
func TestImportStateSetsLookupAttributes(t *testing.T) {
	ctx := context.Background()
	cases := []struct {
		name string
		res  resource.Resource
		id   string
		want map[string]string
	}{
		{"provider", NewProviderResource(), "hashicorp/aws", map[string]string{"id": "hashicorp/aws", "namespace": "hashicorp", "name": "aws"}},
		{"module", NewModuleResource(), "terraform-aws-modules/vpc/aws", map[string]string{"id": "terraform-aws-modules/vpc/aws", "namespace": "terraform-aws-modules", "name": "vpc", "target_provider": "aws"}},
		{"policy library", NewPolicyLibraryResource(), "hashicorp/cis", map[string]string{"id": "hashicorp/cis", "namespace": "hashicorp", "name": "cis"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var schemaResp resource.SchemaResponse
			tc.res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
			if schemaResp.Diagnostics.HasError() {
				t.Fatalf("schema: %v", schemaResp.Diagnostics)
			}
			typ := schemaResp.Schema.Type().TerraformType(ctx)
			resp := resource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, nil)}}

			tc.res.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: tc.id}, &resp)
			if resp.Diagnostics.HasError() {
				t.Fatalf("import: %v", resp.Diagnostics)
			}
			for attr, want := range tc.want {
				var got string
				diags := resp.State.GetAttribute(ctx, pathOf(attr), &got)
				if diags.HasError() {
					t.Fatalf("get %s: %v", attr, diags)
				}
				if got != want {
					t.Errorf("%s = %q, want %q", attr, got, want)
				}
			}
		})
	}

	t.Run("bad id", func(t *testing.T) {
		res := NewModuleResource()
		var schemaResp resource.SchemaResponse
		res.Schema(ctx, resource.SchemaRequest{}, &schemaResp)
		typ := schemaResp.Schema.Type().TerraformType(ctx)
		resp := resource.ImportStateResponse{State: tfsdk.State{Schema: schemaResp.Schema, Raw: tftypes.NewValue(typ, nil)}}
		res.(resource.ResourceWithImportState).ImportState(ctx, resource.ImportStateRequest{ID: "terraform-aws-modules/vpc"}, &resp)
		if !resp.Diagnostics.HasError() {
			t.Fatal("expected an error for a module ID without a target provider")
		}
	})
}

func pathOf(attr string) path.Path { return path.Root(attr) }
