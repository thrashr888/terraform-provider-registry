terraform {
  required_providers {
    registry = {
      source = "thrashr888/registry"
    }
  }
}

provider "registry" {
  # Using default registry.terraform.io
}

# Get AWS provider details
data "registry_provider" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "aws_provider" {
  value = {
    id             = data.registry_provider.aws.id
    description    = data.registry_provider.aws.description
    tier           = data.registry_provider.aws.tier
    latest_version = data.registry_provider.aws.latest_version
    downloads      = data.registry_provider.aws.downloads
  }
}

# Get AWS provider versions
data "registry_provider_versions" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "aws_versions" {
  value = [for v in data.registry_provider_versions.aws.versions : v.version]
}

# Search for kubernetes providers
data "registry_search" "k8s" {
  query = "kubernetes"
  type  = "providers"
  tier  = "official"
}

output "k8s_providers" {
  value = [for r in data.registry_search.k8s.results : {
    name        = r.name
    namespace   = r.namespace
    description = r.description
  }]
}
