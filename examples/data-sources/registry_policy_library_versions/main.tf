# Get all versions of a policy library
data "registry_policy_library_versions" "gcp_networking" {
  namespace = "hashicorp"
  name      = "gcp-networking-terraform"
}

output "versions" {
  value = [for v in data.registry_policy_library_versions.gcp_networking.versions : {
    version     = v.version
    tag         = v.tag
    published   = v.published_at
  }]
}

output "latest_version" {
  value = data.registry_policy_library_versions.gcp_networking.versions[0].version
}
