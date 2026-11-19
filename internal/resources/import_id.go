// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// parseImportID splits an import ID such as "hashicorp/aws" into exactly
// len(fields) non-empty parts separated by "/". fields names each part and
// is used to build the expected format in the error message.
func parseImportID(id string, fields ...string) ([]string, error) {
	parts := strings.Split(strings.TrimSpace(id), "/")
	if len(parts) != len(fields) {
		return nil, fmt.Errorf("expected import ID in the format %q, got %q", strings.Join(fields, "/"), id)
	}
	for i, p := range parts {
		if p == "" {
			return nil, fmt.Errorf("expected import ID in the format %q, got %q (%s is empty)", strings.Join(fields, "/"), id, fields[i])
		}
	}
	return parts, nil
}

// importByID parses req.ID into the named attributes and also stores the full
// ID, so Read has everything it needs to look the object up in the registry.
func importByID(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse, fields ...string) {
	parts, err := parseImportID(req.ID, fields...)
	if err != nil {
		resp.Diagnostics.AddError("Invalid Import ID", err.Error())
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), strings.Join(parts, "/"))...)
	for i, field := range fields {
		resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root(field), parts[i])...)
	}
}
