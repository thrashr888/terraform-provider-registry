// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// TestAccCheckBlocks applies check blocks that read registry data and assert
// with provider::registry::version_satisfies, and an output that proves the
// function returns the expected values.
func TestAccCheckBlocks(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_8_0),
		},
		Steps: []resource.TestStep{
			{
				Config: `
check "random_latest_is_v3_or_later" {
  data "registry_provider" "random" {
    namespace = "hashicorp"
    name      = "random"
  }

  assert {
    condition     = provider::registry::version_satisfies(data.registry_provider.random.latest_version, ">= 3.0.0")
    error_message = "hashicorp/random latest is ${data.registry_provider.random.latest_version}."
  }
}

check "pinned_module_version_published" {
  data "registry_module_versions" "consul" {
    namespace       = "hashicorp"
    name            = "consul"
    target_provider = "aws"
  }

  assert {
    condition     = contains(data.registry_module_versions.consul.versions[*].version, "0.1.0")
    error_message = "hashicorp/consul/aws 0.1.0 is no longer published."
  }
}

output "satisfied" {
  value = provider::registry::version_satisfies("6.2.0", "~> 6.0")
}

output "not_satisfied" {
  value = provider::registry::version_satisfies("7.0.0", "~> 6.0")
}
`,
				// Data sources scoped to a check block are always read during
				// apply ("config will be reloaded to verify a check block"), so
				// the plan after apply is never empty.
				ExpectNonEmptyPlan: true,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckOutput("satisfied", "true"),
					resource.TestCheckOutput("not_satisfied", "false"),
				),
			},
		},
	})
}
