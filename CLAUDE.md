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

# Vet, lint and unit tests
go vet ./...
golangci-lint run ./...
go test ./...

# Acceptance tests (call the live registry; need Terraform 1.14+ on PATH)
TF_ACC=1 go test ./internal/acctest/... -v

# Generate documentation (tfplugindocs is pinned with the go.mod tool directive)
go generate ./...

# Install locally for testing
go install .
```

## Architecture Overview

```
terraform-provider-registry/
├── main.go                    # Provider entry point
├── internal/
│   ├── client/                # Shared HTTP client (timeout, User-Agent)
│   ├── provider/              # Provider configuration and registration
│   ├── datasources/           # registry_provider, _provider_versions, _provider_docs,
│   │                          # _module, _module_versions, _search,
│   │                          # _policy_library, _policy_library_versions
│   ├── resources/             # Read-only, importable registry_provider,
│   │                          # registry_module, registry_policy_library
│   ├── listresources/         # List resources for Terraform 1.14+ list blocks
│   └── acctest/               # Acceptance tests (TF_ACC=1)
├── examples/                  # Example configurations (used by tfplugindocs)
├── templates/                 # Doc templates
├── docs/                      # Generated documentation (do not edit by hand)
└── specs/                     # Specifications
```

## Data Sources

| Data Source | Description |
|-------------|-------------|
| `registry_provider` | Get provider details |
| `registry_provider_versions` | List provider versions |
| `registry_provider_docs` | Get documentation (markdown) |
| `registry_module` | Get module details |
| `registry_module_versions` | List module versions |
| `registry_search` | Search providers/modules |
| `registry_policy_library` | Get policy library info |
| `registry_policy_library_versions` | List policy library versions |

## Resources

`registry_provider`, `registry_module` and `registry_policy_library` are read-only. Import IDs are `namespace/name` (module: `namespace/name/provider`), parsed by `parseImportID` in `internal/resources/import_id.go`.

## Development Workflow

1. Make changes in `internal/`
2. Run `go vet ./...`, `golangci-lint run ./...` and `go test ./...`
3. Test manually with `terraform plan` in `examples/`
4. Regenerate docs with `go generate ./...` and commit the result
5. Build HTTP requests with `client.New` (via provider data), escape path segments and query values, and close response bodies with `defer func() { _ = resp.Body.Close() }()`

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

The provider is published on the Terraform Registry as `thrashr888/registry`.

1. Tag a release on main: `git tag vX.Y.Z && git push origin vX.Y.Z` (Paul decides when)
2. The Release workflow runs GoReleaser, which builds, signs and publishes the GitHub release
3. The Terraform Registry picks up the new GitHub release automatically


<!-- BEGIN BEADS INTEGRATION v:1 profile:minimal hash:7510c1e2 -->
## Beads Issue Tracker

This project uses **bd (beads)** for issue tracking. Run `bd prime` to see full workflow context and commands.

### Quick Reference

```bash
bd ready              # Find available work
bd show <id>          # View issue details
bd update <id> --claim  # Claim work
bd close <id>         # Complete work
```

### Rules

- Use `bd` for ALL task tracking — do NOT use TodoWrite, TaskCreate, or markdown TODO lists
- Run `bd prime` for detailed command reference and session close protocol
- Use `bd remember` for persistent knowledge — do NOT use MEMORY.md files

**Architecture in one line:** issues live in a local Dolt DB; sync uses `refs/dolt/data` on your git remote; `.beads/issues.jsonl` is a passive export. See https://github.com/gastownhall/beads/blob/main/docs/SYNC_CONCEPTS.md for details and anti-patterns.

## Session Completion

**When ending a work session**, you MUST complete ALL steps below. Work is NOT complete until `git push` succeeds.

**MANDATORY WORKFLOW:**

1. **File issues for remaining work** - Create issues for anything that needs follow-up
2. **Run quality gates** (if code changed) - Tests, linters, builds
3. **Update issue status** - Close finished work, update in-progress items
4. **PUSH TO REMOTE** - This is MANDATORY:
   ```bash
   git pull --rebase
   git push
   git status  # MUST show "up to date with origin"
   ```
5. **Clean up** - Clear stashes, prune remote branches
6. **Verify** - All changes committed AND pushed
7. **Hand off** - Provide context for next session

**CRITICAL RULES:**
- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
<!-- END BEADS INTEGRATION -->
