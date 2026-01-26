# Get details about a CIS policy library for AWS
data "registry_policy_library" "cis_aws" {
  namespace = "hashicorp"
  name      = "CIS-Policy-Set-for-AWS-Terraform"
}

output "cis_policy_info" {
  value = {
    id        = data.registry_policy_library.cis_aws.id
    full_name = data.registry_policy_library.cis_aws.full_name
    title     = data.registry_policy_library.cis_aws.title
    source    = data.registry_policy_library.cis_aws.source
    downloads = data.registry_policy_library.cis_aws.downloads
    verified  = data.registry_policy_library.cis_aws.verified
  }
}

# Get AWS networking policy library
data "registry_policy_library" "aws_networking" {
  namespace = "hashicorp"
  name      = "aws-networking-terraform"
}

output "aws_networking_source" {
  description = "GitHub repo for the policy library"
  value       = data.registry_policy_library.aws_networking.source
}

# Get GCP compute policy library
data "registry_policy_library" "gcp_compute" {
  namespace = "hashicorp"
  name      = "gcp-compute-terraform"
}

output "gcp_compute_title" {
  value = data.registry_policy_library.gcp_compute.title
}

# Use policy library info to configure Terraform Cloud
output "policy_set_config" {
  description = "Example configuration for Terraform Cloud policy set"
  value = {
    name   = data.registry_policy_library.cis_aws.name
    source = data.registry_policy_library.cis_aws.source
  }
}
