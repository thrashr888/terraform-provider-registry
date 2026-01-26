# terraform-provider-registry

A read-only Terraform provider for querying the [Terraform Registry API](https://registry.terraform.io).

## Overview

This provider enables Terraform configurations to query the Terraform Registry for:

- **Provider Details** - Get metadata about providers (tier, downloads, latest version)
- **Provider Versions** - List available versions with protocol support
- **Module Details** - Get module information (verified status, source)
- **Module Versions** - List versions with submodules and dependencies
- **Search** - Search for providers and modules with filters
- **Documentation** - Fetch provider docs as markdown
- **Policy Libraries** - Query Sentinel/OPA policy libraries

## Installation

```hcl
terraform {
  required_providers {
    registry = {
      source = "thrashr888/registry"
    }
  }
}
```

## Data Sources

| Data Source | Description | Status |
|-------------|-------------|--------|
| `registry_provider` | Get provider details | ✅ |
| `registry_provider_versions` | List provider versions | ✅ |
| `registry_provider_docs` | Get documentation markdown | ✅ |
| `registry_module` | Get module details | ✅ |
| `registry_module_versions` | List module versions | ✅ |
| `registry_search` | Search providers/modules | ✅ |
| `registry_policy_library` | Get policy library info | ✅ |
| `registry_provider_schema` | Get resource schemas | 🔜 |

## Quick Start

```hcl
provider "registry" {}

# Get provider details
data "registry_provider" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "aws_info" {
  value = {
    tier           = data.registry_provider.aws.tier
    latest_version = data.registry_provider.aws.latest_version
    downloads      = data.registry_provider.aws.downloads
  }
}

# Get module details
data "registry_module" "vpc" {
  namespace       = "terraform-aws-modules"
  name            = "vpc"
  target_provider = "aws"
}

output "vpc_verified" {
  value = data.registry_module.vpc.verified
}

# Search for providers
data "registry_search" "k8s" {
  query = "kubernetes"
  type  = "providers"
  tier  = "official"
}

output "k8s_providers" {
  value = [for r in data.registry_search.k8s.results : r.name]
}
```

## Provider Configuration

```hcl
provider "registry" {
  # Default: registry.terraform.io
  host = "registry.terraform.io"

  # For private registries (Terraform Cloud/Enterprise)
  # host  = "app.terraform.io"
  # token = var.tfc_token
}
```

## Examples

See the [examples](./examples) directory for complete usage examples:

- [Provider basics](./examples/provider) - Provider configuration
- [Complete example](./examples/complete) - All data sources working together
- Individual data source examples in [examples/data-sources](./examples/data-sources)

## Documentation

Full documentation is available in the [docs](./docs) directory and will be published to the Terraform Registry.

## Development

```bash
# Build
go build -o terraform-provider-registry

# Test
go test ./...

# Run acceptance tests
TF_ACC=1 go test ./... -v

# Generate documentation
go generate ./...
```

See `CLAUDE.md` for additional development guidance.

## License

MPL-2.0
