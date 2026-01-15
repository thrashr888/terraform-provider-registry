# terraform-provider-registry

A read-only Terraform provider for querying the [Terraform Registry API](https://registry.terraform.io).

## Status

**Planning Phase** - This project is being planned. See the beads issues for implementation roadmap.

```bash
bd list  # View planned work
bd ready # See what's ready to implement
```

## Overview

This provider enables Terraform configurations to query the Terraform Registry for:

- **Provider Details** - Get metadata about providers
- **Provider Versions** - List available versions
- **Provider Schemas** - Retrieve resource/data source schemas
- **Module Details** - Get module information
- **Search** - Search for providers and modules
- **Documentation** - Fetch provider docs as markdown
- **Policy Libraries** - Query Sentinel/OPA policy libraries

## Planned Data Sources

| Data Source | Description | Status |
|-------------|-------------|--------|
| `registry_provider` | Get provider details | Planned |
| `registry_provider_versions` | List provider versions | Planned |
| `registry_provider_schema` | Get resource schemas | Planned |
| `registry_module` | Get module details | Planned |
| `registry_module_versions` | List module versions | Planned |
| `registry_search` | Search providers/modules | Planned |
| `registry_provider_docs` | Get documentation | Planned |
| `registry_policy_library` | Get policy info | Planned |

## Example Usage (Planned)

```hcl
provider "registry" {}

data "registry_provider" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "aws_latest_version" {
  value = data.registry_provider.aws.latest_version
}

data "registry_search" "kubernetes" {
  query = "kubernetes"
  type  = "providers"
}
```

## Development

See `CLAUDE.md` for development guidance and `specs/` for detailed specifications.

## License

MPL-2.0
