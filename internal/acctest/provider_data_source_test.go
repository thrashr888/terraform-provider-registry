// Copyright (c) Paul Thrasher
// SPDX-License-Identifier: MPL-2.0

package acctest

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccProviderDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.registry_provider.test", "namespace", "hashicorp"),
					resource.TestCheckResourceAttr("data.registry_provider.test", "name", "aws"),
					resource.TestCheckResourceAttrSet("data.registry_provider.test", "description"),
					resource.TestCheckResourceAttrSet("data.registry_provider.test", "tier"),
					resource.TestCheckResourceAttrSet("data.registry_provider.test", "latest_version"),
				),
			},
		},
	})
}

const testAccProviderDataSourceConfig = `
data "registry_provider" "test" {
  namespace = "hashicorp"
  name      = "aws"
}
`

func TestAccProviderVersionsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderVersionsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.registry_provider_versions.test", "namespace", "hashicorp"),
					resource.TestCheckResourceAttr("data.registry_provider_versions.test", "name", "aws"),
					resource.TestCheckResourceAttrSet("data.registry_provider_versions.test", "versions.#"),
				),
			},
		},
	})
}

const testAccProviderVersionsDataSourceConfig = `
data "registry_provider_versions" "test" {
  namespace = "hashicorp"
  name      = "aws"
}
`

func TestAccSearchDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSearchDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.registry_search.test", "query", "aws"),
					resource.TestCheckResourceAttr("data.registry_search.test", "type", "providers"),
					resource.TestCheckResourceAttrSet("data.registry_search.test", "results.#"),
				),
			},
		},
	})
}

const testAccSearchDataSourceConfig = `
data "registry_search" "test" {
  query = "aws"
  type  = "providers"
}
`

func TestAccProviderDocsDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderDocsDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.registry_provider_docs.test", "namespace", "hashicorp"),
					resource.TestCheckResourceAttr("data.registry_provider_docs.test", "name", "aws"),
					resource.TestCheckResourceAttr("data.registry_provider_docs.test", "version", "5.0.0"),
					resource.TestCheckResourceAttr("data.registry_provider_docs.test", "doc_type", "resources"),
					resource.TestCheckResourceAttr("data.registry_provider_docs.test", "doc_name", "instance"),
					resource.TestCheckResourceAttrSet("data.registry_provider_docs.test", "content"),
					resource.TestCheckResourceAttrSet("data.registry_provider_docs.test", "title"),
				),
			},
		},
	})
}

const testAccProviderDocsDataSourceConfig = `
data "registry_provider_docs" "test" {
  namespace = "hashicorp"
  name      = "aws"
  version   = "5.0.0"
  doc_type  = "resources"
  doc_name  = "instance"
}
`

func TestAccPolicyLibraryDataSource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: ProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccPolicyLibraryDataSourceConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("data.registry_policy_library.test", "namespace", "hashicorp"),
					resource.TestCheckResourceAttr("data.registry_policy_library.test", "name", "CIS-Policy-Set-for-AWS-Terraform"),
					resource.TestCheckResourceAttrSet("data.registry_policy_library.test", "title"),
					resource.TestCheckResourceAttrSet("data.registry_policy_library.test", "source"),
				),
			},
		},
	})
}

const testAccPolicyLibraryDataSourceConfig = `
data "registry_policy_library" "test" {
  namespace = "hashicorp"
  name      = "CIS-Policy-Set-for-AWS-Terraform"
}
`
