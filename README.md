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
| `registry_policy_library_versions` | List policy library versions | ✅ |

## Resources

Read-only resources for tracking registry objects in Terraform state. Creating one only reads the registry; destroying one only removes it from state.

| Resource | Import ID |
|----------|-----------|
| `registry_provider` | `namespace/name`, e.g. `hashicorp/aws` |
| `registry_module` | `namespace/name/provider`, e.g. `terraform-aws-modules/vpc/aws` |
| `registry_policy_library` | `namespace/name`, e.g. `hashicorp/gcp-networking-terraform` |

```shell
terraform import registry_provider.aws hashicorp/aws
```

## Terraform Search

With Terraform 1.14 or later, `registry_provider`, `registry_module` and `registry_policy_library` work in `list` blocks with `terraform query`. Search providers and modules by text with `query`, or list a `namespace`. `terraform query -generate-config-out=generated.tf` writes import blocks for the results. See the [Terraform Search guide](./docs/guides/terraform-search.md) and [examples/search](./examples/search).

```hcl
# search.tfquery.hcl
list "registry_module" "aws_vpc" {
  provider = registry
  config {
    query           = "vpc"
    target_provider = "aws"
  }
}
```

## Checks

Data sources and the `provider::registry::version_satisfies(version, constraint)` function work in `check` blocks, so `plan`, `apply` and HCP Terraform health assessments warn you about registry changes. See the [check blocks guide](./docs/guides/check-blocks.md) and [examples/checks](./examples/checks).

```hcl
check "aws_constraint_covers_latest" {
  data "registry_provider" "aws" {
    namespace = "hashicorp"
    name      = "aws"
  }

  assert {
    condition     = provider::registry::version_satisfies(data.registry_provider.aws.latest_version, "~> 6.0")
    error_message = "hashicorp/aws ${data.registry_provider.aws.latest_version} is out and ~> 6.0 won't pick it up."
  }
}
```

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

Full documentation is published on the [Terraform Registry](https://registry.terraform.io/providers/thrashr888/registry/latest/docs) and lives in the [docs](./docs) directory, generated from [templates](./templates) and [examples](./examples).

## Development

```bash
# Build
go build -o terraform-provider-registry

# Test
go test ./...

# Lint
golangci-lint run ./...

# Run acceptance tests (they call the live registry)
TF_ACC=1 go test ./internal/acctest/... -v

# Generate documentation (tfplugindocs is pinned with the go.mod tool directive)
go generate ./...
```

See `CLAUDE.md` for additional development guidance.

## License

MPL-2.0
