# Search for official Kubernetes providers
data "registry_search" "k8s_providers" {
  query = "kubernetes"
  type  = "providers"
  tier  = "official"
}

output "k8s_providers" {
  value = [for r in data.registry_search.k8s_providers.results : {
    id          = r.id
    name        = r.name
    namespace   = r.namespace
    description = r.description
    tier        = r.tier
  }]
}

# Search for AWS modules
data "registry_search" "aws_modules" {
  query     = "vpc"
  type      = "modules"
  namespace = "terraform-aws-modules"
  limit     = 10
}

output "aws_modules" {
  value = [for r in data.registry_search.aws_modules.results : {
    id          = r.id
    name        = r.name
    description = r.description
    downloads   = r.downloads
  }]
}

# Search for database providers
data "registry_search" "database" {
  query = "database"
  type  = "providers"
  limit = 5
}

output "database_provider_count" {
  value = data.registry_search.database.total_count
}

# Search with pagination
data "registry_search" "providers_page2" {
  query  = "cloud"
  type   = "providers"
  limit  = 10
  offset = 10
}

output "providers_page2" {
  value = [for r in data.registry_search.providers_page2.results : r.name]
}
