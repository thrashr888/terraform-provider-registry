// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/querycheck"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccTerraformSearch runs `terraform query` (Terraform Search) against the
// live registry with every list resource, by namespace and by text search.
func TestAccTerraformSearch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_14_0),
		},
		Steps: []resource.TestStep{
			{
				// Query mode needs an earlier config step.
				Config: `
data "registry_provider" "random" {
  namespace = "hashicorp"
  name      = "random"
}
`,
			},
			{
				Query: true,
				Config: `
provider "registry" {}

list "registry_provider" "by_namespace" {
  provider = registry
  config {
    namespace = "hashicorp"
    tier      = "official"
  }
}

list "registry_provider" "by_text" {
  provider = registry
  config {
    query = "random"
  }
}

list "registry_module" "by_text" {
  provider = registry
  config {
    query           = "vpc"
    namespace       = "terraform-aws-modules"
    target_provider = "aws"
  }
}

list "registry_module" "by_namespace" {
  provider = registry
  config {
    namespace = "terraform-aws-modules"
  }
}

list "registry_policy_library" "by_namespace" {
  provider = registry
  config {
    namespace = "hashicorp"
  }
}
`,
				QueryResultChecks: []querycheck.QueryResultCheck{
					querycheck.ExpectIdentity("registry_provider.by_namespace", map[string]knownvalue.Check{
						"namespace": knownvalue.StringExact("hashicorp"),
						"name":      knownvalue.StringExact("aws"),
					}),
					querycheck.ExpectIdentity("registry_provider.by_text", map[string]knownvalue.Check{
						"namespace": knownvalue.StringExact("hashicorp"),
						"name":      knownvalue.StringExact("random"),
					}),
					querycheck.ExpectIdentity("registry_module.by_text", map[string]knownvalue.Check{
						"namespace":       knownvalue.StringExact("terraform-aws-modules"),
						"name":            knownvalue.StringExact("vpc"),
						"target_provider": knownvalue.StringExact("aws"),
					}),
					querycheck.ExpectLengthAtLeast("registry_module.by_namespace", 1),
					querycheck.ExpectLengthAtLeast("registry_policy_library.by_namespace", 1),
				},
			},
		},
	})
}
