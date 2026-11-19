// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package functions

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func runVersionSatisfies(t *testing.T, version, constraint string) (bool, *function.FuncError) {
	t.Helper()
	ctx := context.Background()
	resp := function.RunResponse{Result: function.NewResultData(types.BoolUnknown())}
	NewVersionSatisfiesFunction().Run(ctx, function.RunRequest{
		Arguments: function.NewArgumentsData([]attr.Value{types.StringValue(version), types.StringValue(constraint)}),
	}, &resp)
	if resp.Error != nil {
		return false, resp.Error
	}
	got, ok := resp.Result.Value().(types.Bool)
	if !ok {
		t.Fatalf("result is %T", resp.Result.Value())
	}
	return got.ValueBool(), nil
}

func TestVersionSatisfies(t *testing.T) {
	cases := []struct {
		version, constraint string
		want                bool
	}{
		{"6.2.0", "~> 6.0", true},
		{"7.0.0", "~> 6.0", false},
		{"v6.2.0", ">= 6.0, < 7.0", true},
		{"5.9.1", ">= 6.0", false},
		{"6.0.0-beta1", ">= 6.0", false},
		{"1.2.3", "= 1.2.3", true},
	}
	for _, tc := range cases {
		got, err := runVersionSatisfies(t, tc.version, tc.constraint)
		if err != nil {
			t.Fatalf("%s %s: %v", tc.version, tc.constraint, err)
		}
		if got != tc.want {
			t.Errorf("version_satisfies(%q, %q) = %v, want %v", tc.version, tc.constraint, got, tc.want)
		}
	}
}

func TestVersionSatisfiesErrors(t *testing.T) {
	if _, err := runVersionSatisfies(t, "not-a-version", "~> 1.0"); err == nil || err.FunctionArgument == nil || *err.FunctionArgument != 0 {
		t.Fatalf("bad version: got %v, want an error on argument 0", err)
	}
	if _, err := runVersionSatisfies(t, "1.0.0", "~~ 1"); err == nil || err.FunctionArgument == nil || *err.FunctionArgument != 1 {
		t.Fatalf("bad constraint: got %v, want an error on argument 1", err)
	}
}
