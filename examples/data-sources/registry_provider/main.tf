# Get details about the AWS provider
data "registry_provider" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "provider_info" {
  value = {
    id             = data.registry_provider.aws.id
    description    = data.registry_provider.aws.description
    tier           = data.registry_provider.aws.tier
    latest_version = data.registry_provider.aws.latest_version
    downloads      = data.registry_provider.aws.downloads
    source         = data.registry_provider.aws.source
  }
}

# Get details about the Google provider
data "registry_provider" "google" {
  namespace = "hashicorp"
  name      = "google"
}

# Check if a provider is official tier
output "is_official" {
  value = data.registry_provider.aws.tier == "official"
}
