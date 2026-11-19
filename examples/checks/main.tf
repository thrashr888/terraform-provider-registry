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
