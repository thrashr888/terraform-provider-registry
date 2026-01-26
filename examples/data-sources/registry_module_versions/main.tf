# List available versions for the VPC module
data "registry_module_versions" "vpc" {
  namespace       = "terraform-aws-modules"
  name            = "vpc"
  target_provider = "aws"
}

# Output all version strings
output "all_versions" {
  value = [for v in data.registry_module_versions.vpc.versions : v.version]
}

# Get the latest version
output "latest_version" {
  value = length(data.registry_module_versions.vpc.versions) > 0 ? data.registry_module_versions.vpc.versions[0].version : "none"
}

# Count total available versions
output "version_count" {
  value = length(data.registry_module_versions.vpc.versions)
}

# Get submodules for the latest version (if any)
output "latest_submodules" {
  value = length(data.registry_module_versions.vpc.versions) > 0 ? [
    for sm in data.registry_module_versions.vpc.versions[0].submodules : sm.path
  ] : []
}
