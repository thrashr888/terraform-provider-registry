# List all HashiCorp providers
list "registry_provider" "hashicorp" {
  provider = registry

  config {
    namespace = "hashicorp"
  }
}

# List only official HashiCorp providers
list "registry_provider" "hashicorp_official" {
  provider = registry

  config {
    namespace = "hashicorp"
    tier      = "official"
  }
}

# Search providers by text
list "registry_provider" "kubernetes" {
  provider = registry

  config {
    query = "kubernetes"
  }
}
