# List available versions for the AWS provider
data "registry_provider_versions" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

# Output all version strings
output "all_versions" {
  value = [for v in data.registry_provider_versions.aws.versions : v.version]
}

# Get the latest version
output "latest_version" {
  value = data.registry_provider_versions.aws.versions[0].version
}

# Find versions that support protocol 6.0
output "protocol_6_versions" {
  value = [
    for v in data.registry_provider_versions.aws.versions : v.version
    if contains(v.protocols, "6.0")
  ]
}

# Count total available versions
output "version_count" {
  value = length(data.registry_provider_versions.aws.versions)
}
