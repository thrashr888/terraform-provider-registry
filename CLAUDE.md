# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**terraform-provider-registry** is a read-only Terraform provider that queries the [Terraform Registry API](https://registry.terraform.io) to retrieve metadata about providers, modules, and policies.

### Use Cases

- **Documentation Generation** - Pull resource schemas and docs
- **Compliance Checking** - Verify provider/module versions
- **Dependency Analysis** - Understand available resources
- **Tooling Integration** - Build tools needing registry metadata

## Build Commands

```bash
# Build the provider
go build -o terraform-provider-registry

# Run tests
go test ./...

# Run acceptance tests (requires TF_ACC=1)
TF_ACC=1 go test ./... -v

# Generate documentation
go generate ./...

# Install locally for testing
go install .
```

## Architecture Overview

```
terraform-provider-registry/
├── main.go                    # Provider entry point
├── internal/
│   ├── provider/
│   │   └── provider.go        # Provider configuration
│   └── datasources/
│       ├── provider.go        # registry_provider
│       ├── provider_versions.go
│       ├── provider_schema.go
│       ├── module.go          # registry_module
│       ├── module_versions.go
│       ├── search.go          # registry_search
│       ├── provider_docs.go   # registry_provider_docs
│       └── policy_library.go  # registry_policy_library
├── examples/                  # Example configurations
├── docs/                      # Generated documentation
└── specs/                     # Specifications
```

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `registry_provider` | Get provider details |
| `registry_provider_versions` | List provider versions |
| `registry_provider_schema` | Get resource/data source schemas |
| `registry_module` | Get module details |
| `registry_module_versions` | List module versions |
| `registry_search` | Search providers/modules |
| `registry_provider_docs` | Get documentation (markdown) |
| `registry_policy_library` | Get policy library info |

## Development Workflow

1. Make changes to data sources in `internal/datasources/`
2. Run `go test ./...` to verify
3. Test manually with `terraform plan` in `examples/`
4. Generate docs with `go generate ./...`

## Terraform Plugin Framework

This provider uses the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

Key patterns:
- Data sources implement `datasource.DataSource` interface
- Use `schema.Schema` for defining attributes
- Implement `Read` method for fetching data
- Use `diag.Diagnostics` for error reporting

## Registry API

Base URL: `https://registry.terraform.io/v1/`

Key endpoints:
- `GET /providers/{namespace}/{name}` - Provider details
- `GET /providers/{namespace}/{name}/versions` - Versions
- `GET /modules/{namespace}/{name}/{provider}` - Module details
- `GET /search` - Search endpoint

## Beads Issue Tracking

This repository uses `bd` (beads) for issue tracking.

### Essential Commands

```bash
# Create issues
bd create --title="Implement feature X" --type=feature --priority=1

# List and filter
bd list --status=open
bd ready                    # Show unblocked work

# Update work
bd update <id> --status=in_progress
bd close <id>

# Dependencies
bd dep add <issue> <depends-on>
```

## Release Process

1. Tag release: `git tag v0.1.0`
2. Push tags: `git push --tags`
3. GoReleaser builds binaries
4. Publish to Terraform Registry
