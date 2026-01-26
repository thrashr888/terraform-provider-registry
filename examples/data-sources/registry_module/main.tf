# Get details about the VPC module
data "registry_module" "vpc" {
  namespace       = "terraform-aws-modules"
  name            = "vpc"
  target_provider = "aws"
}

output "module_info" {
  value = {
    id             = data.registry_module.vpc.id
    description    = data.registry_module.vpc.description
    source         = data.registry_module.vpc.source
    latest_version = data.registry_module.vpc.latest_version
    verified       = data.registry_module.vpc.verified
    downloads      = data.registry_module.vpc.downloads
  }
}

# Get details about the EKS module
data "registry_module" "eks" {
  namespace       = "terraform-aws-modules"
  name            = "eks"
  target_provider = "aws"
}

# Check if module is verified
output "eks_verified" {
  value = data.registry_module.eks.verified
}

# Use module source in another configuration
output "vpc_module_source" {
  description = "Use this source string in your module block"
  value       = "terraform-aws-modules/vpc/aws"
}
