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
