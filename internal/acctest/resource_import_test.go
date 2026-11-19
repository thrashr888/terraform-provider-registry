// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Each test creates the resource, then imports it by its documented ID and
// checks the imported state matches. Download counts move between steps, so
// they are ignored.

func TestAccProviderResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "registry_provider" "test" {
  namespace = "hashicorp"
  name      = "random"
}
`,
				Check: resource.TestCheckResourceAttr("registry_provider.test", "id", "hashicorp/random"),
			},
			{
				ResourceName:            "registry_provider.test",
				ImportState:             true,
				ImportStateId:           "hashicorp/random",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"downloads"},
			},
		},
	})
}

func TestAccModuleResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "registry_module" "test" {
  namespace       = "hashicorp"
  name            = "consul"
  target_provider = "aws"
}
`,
				Check: resource.TestCheckResourceAttr("registry_module.test", "id", "hashicorp/consul/aws"),
			},
			{
				ResourceName:            "registry_module.test",
				ImportState:             true,
				ImportStateId:           "hashicorp/consul/aws",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"downloads"},
			},
		},
	})
}

func TestAccPolicyLibraryResource_import(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
resource "registry_policy_library" "test" {
  namespace = "hashicorp"
  name      = "gcp-networking-terraform"
}
`,
				Check: resource.TestCheckResourceAttr("registry_policy_library.test", "id", "hashicorp/gcp-networking-terraform"),
			},
			{
				ResourceName:            "registry_policy_library.test",
				ImportState:             true,
				ImportStateId:           "hashicorp/gcp-networking-terraform",
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"downloads"},
			},
		},
	})
}
