# Track the AWS provider in state
# This is a read-only resource - it fetches metadata from the registry
resource "registry_provider" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "aws_provider_info" {
  value = {
    description    = registry_provider.aws.description
    tier           = registry_provider.aws.tier
    latest_version = registry_provider.aws.latest_version
    downloads      = registry_provider.aws.downloads
  }
}
