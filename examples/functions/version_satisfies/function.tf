# Warn when a new major version of a provider is out that your
# version constraint won't pick up.
check "aws_constraint_covers_latest" {
  data "registry_provider" "aws" {
    namespace = "hashicorp"
    name      = "aws"
  }

  assert {
    condition     = provider::registry::version_satisfies(data.registry_provider.aws.latest_version, "~> 6.0")
    error_message = "hashicorp/aws ${data.registry_provider.aws.latest_version} is out and \"~> 6.0\" won't pick it up."
  }
}

output "satisfied" {
  value = provider::registry::version_satisfies("6.2.0", ">= 6.0, < 7.0") # true
}
