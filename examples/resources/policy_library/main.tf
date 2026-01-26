# Track the GCP Networking policy library in state
resource "registry_policy_library" "gcp_networking" {
  namespace = "hashicorp"
  name      = "gcp-networking-terraform"
}

output "policy_library_info" {
  value = {
    title     = registry_policy_library.gcp_networking.title
    source    = registry_policy_library.gcp_networking.source
    verified  = registry_policy_library.gcp_networking.verified
    downloads = registry_policy_library.gcp_networking.downloads
  }
}
