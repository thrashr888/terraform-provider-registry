# Track the VPC module in state
# This is a read-only resource - it fetches metadata from the registry
resource "registry_module" "vpc" {
  namespace       = "terraform-aws-modules"
  name            = "vpc"
  target_provider = "aws"
}

output "vpc_module_info" {
  value = {
    description = registry_module.vpc.description
    source      = registry_module.vpc.source
    verified    = registry_module.vpc.verified
    downloads   = registry_module.vpc.downloads
  }
}
