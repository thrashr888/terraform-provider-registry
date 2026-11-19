---
page_title: "Find and import registry objects with Terraform Search"
subcategory: ""
description: |-
  Use list blocks and terraform query to search the registry and generate import configuration.
---

# Find and import registry objects with Terraform Search

Terraform 1.14 adds Terraform Search: `list` blocks in `.tfquery.hcl` files, run with `terraform query`. This provider has list resources for `registry_provider`, `registry_module` and `registry_policy_library`.

- `registry_provider` and `registry_module` take `query` for a text search, `namespace` to list a namespace, or both. Set at least one.
- `registry_policy_library` lists a `namespace`.

## Example

`main.tf` declares the provider requirement and `search.tfquery.hcl` holds the list blocks:

```terraform
provider "registry" {}

# Every official HashiCorp provider.
list "registry_provider" "hashicorp_official" {
  provider = registry

  config {
    namespace = "hashicorp"
    tier      = "official"
  }
}

# Text search across all providers.
list "registry_provider" "kubernetes" {
  provider = registry

  config {
    query = "kubernetes"
  }
}

# Text search for AWS VPC modules.
list "registry_module" "aws_vpc" {
  provider = registry

  config {
    query           = "vpc"
    target_provider = "aws"
  }
}

# Every policy library in a namespace.
list "registry_policy_library" "hashicorp" {
  provider = registry

  config {
    namespace = "hashicorp"
  }
}
```

Run the search:

```shell
terraform init
terraform query
```

Each result shows its identity, such as `name=aws,namespace=hashicorp`.

## Generate import configuration

`terraform query -generate-config-out=generated.tf` writes a `resource` block and an `import` block (by identity) for every result. Review the file, keep what you want, then run `terraform apply` to bring those objects into state as read-only `registry_*` resources.

You can also import one object by ID, for example `terraform import registry_provider.aws hashicorp/aws`, or with an `import` block that uses `identity`.
