terraform {
  required_providers {
    registry = {
      source = "thrashr888/registry"
    }
  }
}

provider "registry" {
  # Uses default registry.terraform.io
  # For private registries:
  # host  = "my-registry.example.com"
  # token = var.registry_token
}

# =============================================================================
# Provider Discovery
# =============================================================================

# Get AWS provider details
data "registry_provider" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

# Get available versions
data "registry_provider_versions" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

# =============================================================================
# Module Discovery
# =============================================================================

# Get VPC module info
data "registry_module" "vpc" {
  namespace       = "terraform-aws-modules"
  name            = "vpc"
  target_provider = "aws"
}

# Get module versions
data "registry_module_versions" "vpc" {
  namespace       = "terraform-aws-modules"
  name            = "vpc"
  target_provider = "aws"
}

# =============================================================================
# Registry Search
# =============================================================================

# Find all official networking-related providers
data "registry_search" "networking" {
  query = "network"
  type  = "providers"
  tier  = "official"
  limit = 5
}

# =============================================================================
# Documentation Retrieval
# =============================================================================

# Get aws_vpc resource documentation
data "registry_provider_docs" "vpc_resource" {
  namespace = "hashicorp"
  name      = "aws"
  version   = data.registry_provider.aws.latest_version
  doc_type  = "resources"
  doc_name  = "vpc"
}

# =============================================================================
# Policy Libraries
# =============================================================================

# Get AWS CIS policy library
data "registry_policy_library" "cis" {
  namespace = "hashicorp"
  name      = "CIS-Policy-Set-for-AWS-Terraform"
}

# =============================================================================
# Outputs - Provider Info
# =============================================================================

output "provider_summary" {
  description = "AWS provider summary"
  value = {
    name           = "${data.registry_provider.aws.namespace}/${data.registry_provider.aws.name}"
    tier           = data.registry_provider.aws.tier
    latest_version = data.registry_provider.aws.latest_version
    total_versions = length(data.registry_provider_versions.aws.versions)
    downloads      = data.registry_provider.aws.downloads
  }
}

# =============================================================================
# Outputs - Module Info
# =============================================================================

output "module_summary" {
  description = "VPC module summary"
  value = {
    name           = data.registry_module.vpc.id
    verified       = data.registry_module.vpc.verified
    latest_version = data.registry_module.vpc.latest_version
    total_versions = length(data.registry_module_versions.vpc.versions)
    source         = data.registry_module.vpc.source
  }
}

# =============================================================================
# Outputs - Search Results
# =============================================================================

output "networking_providers" {
  description = "Official networking providers"
  value = [for p in data.registry_search.networking.results : {
    name        = p.name
    namespace   = p.namespace
    description = substr(p.description, 0, 100)
  }]
}

# =============================================================================
# Outputs - Documentation
# =============================================================================

output "vpc_docs_info" {
  description = "VPC resource documentation metadata"
  value = {
    title       = data.registry_provider_docs.vpc_resource.title
    subcategory = data.registry_provider_docs.vpc_resource.subcategory
    has_content = length(data.registry_provider_docs.vpc_resource.content) > 0
  }
}

# =============================================================================
# Outputs - Policy Library
# =============================================================================

output "policy_summary" {
  description = "CIS policy library info"
  value = {
    title  = data.registry_policy_library.cis.title
    source = data.registry_policy_library.cis.source
  }
}

# =============================================================================
# Use Case: Version Constraint Validation
# =============================================================================

locals {
  # Check if a specific version exists
  target_version   = "5.0.0"
  version_exists   = contains([for v in data.registry_provider_versions.aws.versions : v.version], local.target_version)

  # Get recent versions (last 10)
  recent_versions = slice([for v in data.registry_provider_versions.aws.versions : v.version], 0, min(10, length(data.registry_provider_versions.aws.versions)))
}

output "version_check" {
  description = "Version availability check"
  value = {
    target_version  = local.target_version
    exists          = local.version_exists
    recent_versions = local.recent_versions
  }
}
