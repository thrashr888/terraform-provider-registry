// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccModuleDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModuleDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.registry_module.test", "namespace", "hashicorp"),
					resource.TestCheckResourceAttr("data.registry_module.test", "name", "consul"),
					resource.TestCheckResourceAttr("data.registry_module.test", "target_provider", "aws"),
					resource.TestCheckResourceAttrSet("data.registry_module.test", "description"),
				),
			},
		},
	})
}

const testAccModuleDataSourceConfig = `
data "registry_module" "test" {
  namespace = "hashicorp"
  name      = "consul"
  target_provider = "aws"
}
`

func TestAccModuleVersionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccModuleVersionsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.registry_module_versions.test", "namespace", "hashicorp"),
					resource.TestCheckResourceAttr("data.registry_module_versions.test", "name", "consul"),
					resource.TestCheckResourceAttr("data.registry_module_versions.test", "target_provider", "aws"),
					resource.TestCheckResourceAttrSet("data.registry_module_versions.test", "versions.#"),
				),
			},
		},
	})
}

const testAccModuleVersionsDataSourceConfig = `
data "registry_module_versions" "test" {
  namespace = "hashicorp"
  name      = "consul"
  target_provider = "aws"
}
`
