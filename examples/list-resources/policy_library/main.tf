# List all HashiCorp policy libraries
list "registry_policy_library" "hashicorp" {
  provider = registry

  config {
    namespace = "hashicorp"
  }
}
