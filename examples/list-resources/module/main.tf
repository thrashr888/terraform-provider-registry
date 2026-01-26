# List all modules from terraform-aws-modules
list "registry_module" "aws_modules" {
  provider = registry

  config {
    namespace = "terraform-aws-modules"
  }
}

# List only AWS modules from terraform-aws-modules
list "registry_module" "aws_modules_aws" {
  provider = registry

  config {
    namespace       = "terraform-aws-modules"
    target_provider = "aws"
  }
}
