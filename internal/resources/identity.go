// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource/identityschema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Resource identities let Terraform 1.12+ import by identity and let
// `terraform query` (Terraform Search, 1.14+) return list results. The list
// resources in internal/listresources set these same models.

// ProviderIdentityModel identifies a registry provider.
type ProviderIdentityModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
}

// ModuleIdentityModel identifies a registry module.
type ModuleIdentityModel struct {
	Namespace      types.String `tfsdk:"namespace"`
	Name           types.String `tfsdk:"name"`
	TargetProvider types.String `tfsdk:"target_provider"`
}

// PolicyLibraryIdentityModel identifies a registry policy library.
type PolicyLibraryIdentityModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Name      types.String `tfsdk:"name"`
}

func identityAttr(description string) identityschema.StringAttribute {
	return identityschema.StringAttribute{Description: description, RequiredForImport: true}
}

func namespaceNameIdentitySchema(kind string) identityschema.Schema {
	return identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"namespace": identityAttr(kind + " namespace."),
			"name":      identityAttr(kind + " name."),
		},
	}
}

func moduleIdentitySchema() identityschema.Schema {
	return identityschema.Schema{
		Attributes: map[string]identityschema.Attribute{
			"namespace":       identityAttr("Module namespace."),
			"name":            identityAttr("Module name."),
			"target_provider": identityAttr("Target provider of the module (e.g., aws)."),
		},
	}
}

// setIdentity stores the identity when Terraform supports it (identity is nil
// on Terraform versions before 1.12).
func setIdentity(ctx context.Context, identity *tfsdk.ResourceIdentity, value any) diag.Diagnostics {
	if identity == nil {
		return nil
	}
	return identity.Set(ctx, value)
}
