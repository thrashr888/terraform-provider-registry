---
page_title: "Registry checks with check blocks"
subcategory: ""
description: |-
  Use registry data and the version_satisfies function in Terraform check blocks.
---

# Registry checks with check blocks

[Check blocks](https://developer.hashicorp.com/terraform/language/block/check) validate your infrastructure on every `plan` and `apply` without blocking them. A failed assertion is a warning. On HCP Terraform, health assessments run them continuously and alert you when one fails.

This provider makes registry facts available to check blocks:

- A data source scoped inside a `check` block reads the registry each run. If the read fails, Terraform shows a warning instead of failing the run.
- `provider::registry::version_satisfies(version, constraint)` tests a version against the same constraint syntax as `required_providers` (Terraform 1.8+).

## Examples

```terraform
terraform {
  required_providers {
    registry = {
      source = "thrashr888/registry"
    }
  }
}

provider "registry" {}

# A new major version is out that our constraint won't pick up.
check "aws_constraint_covers_latest" {
  data "registry_provider" "aws" {
    namespace = "hashicorp"
    name      = "aws"
  }

  assert {
    condition     = provider::registry::version_satisfies(data.registry_provider.aws.latest_version, "~> 6.0")
    error_message = "hashicorp/aws ${data.registry_provider.aws.latest_version} is out and \"~> 6.0\" won't pick it up."
  }
}

# The module version we pin is still published.
check "vpc_module_pin_published" {
  data "registry_module_versions" "vpc" {
    namespace       = "terraform-aws-modules"
    name            = "vpc"
    target_provider = "aws"
  }

  assert {
    condition     = contains(data.registry_module_versions.vpc.versions[*].version, "6.0.1")
    error_message = "terraform-aws-modules/vpc/aws 6.0.1 is no longer published."
  }
}

# We only use verified modules.
check "vpc_module_verified" {
  data "registry_module" "vpc" {
    namespace       = "terraform-aws-modules"
    name            = "vpc"
    target_provider = "aws"
  }

  assert {
    condition     = data.registry_module.vpc.verified
    error_message = "terraform-aws-modules/vpc/aws is not a verified module."
  }
}

# A provider we depend on is still an official or partner provider.
check "kubernetes_provider_tier" {
  data "registry_provider" "kubernetes" {
    namespace = "hashicorp"
    name      = "kubernetes"
  }

  assert {
    condition     = contains(["official", "partner"], data.registry_provider.kubernetes.tier)
    error_message = "hashicorp/kubernetes is now ${data.registry_provider.kubernetes.tier}."
  }
}
```

## Ideas

- Warn when a new major version of a provider is out that your constraint won't pick up.
- Warn when a pinned module version is withdrawn from the registry.
- Require verified modules, or official and partner providers.
- Use the `registry_search` data source to check that an internal module is published in your private registry (set `host` and `token` on the provider).
