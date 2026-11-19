// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package functions

import (
	"context"

	goversion "github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/function"
)

var _ function.Function = &VersionSatisfiesFunction{}

// VersionSatisfiesFunction reports whether a version meets a Terraform-style
// version constraint. It is meant for assert conditions in check blocks.
type VersionSatisfiesFunction struct{}

// NewVersionSatisfiesFunction creates the version_satisfies function.
func NewVersionSatisfiesFunction() function.Function {
	return &VersionSatisfiesFunction{}
}

func (f *VersionSatisfiesFunction) Metadata(ctx context.Context, req function.MetadataRequest, resp *function.MetadataResponse) {
	resp.Name = "version_satisfies"
}

func (f *VersionSatisfiesFunction) Definition(ctx context.Context, req function.DefinitionRequest, resp *function.DefinitionResponse) {
	resp.Definition = function.Definition{
		Summary: "Check a version against a version constraint",
		MarkdownDescription: "Returns true when `version` meets `constraint`, using the same constraint syntax as " +
			"`required_providers` and module `version` arguments (for example `>= 5.0, < 7.0` or `~> 6.2`). " +
			"Use it in `check` block assertions with versions read from this provider's data sources.",
		Parameters: []function.Parameter{
			function.StringParameter{
				Name:                "version",
				MarkdownDescription: "Version to test, such as `6.2.0` or `v6.2.0`.",
			},
			function.StringParameter{
				Name:                "constraint",
				MarkdownDescription: "Version constraint, such as `~> 6.0`.",
			},
		},
		Return: function.BoolReturn{},
	}
}

func (f *VersionSatisfiesFunction) Run(ctx context.Context, req function.RunRequest, resp *function.RunResponse) {
	var version, constraint string
	resp.Error = function.ConcatFuncErrors(resp.Error, req.Arguments.Get(ctx, &version, &constraint))
	if resp.Error != nil {
		return
	}

	v, err := goversion.NewVersion(version)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(0, "Invalid version: "+err.Error())
		return
	}
	c, err := goversion.NewConstraint(constraint)
	if err != nil {
		resp.Error = function.NewArgumentFuncError(1, "Invalid version constraint: "+err.Error())
		return
	}

	resp.Error = function.ConcatFuncErrors(resp.Error, resp.Result.Set(ctx, c.Check(v)))
}
