# Terraform Provider for Terraform Registry

## Overview

A read-only Terraform provider that queries the [Terraform Registry API](https://registry.terraform.io) to retrieve metadata about providers, modules, and policies. This enables infrastructure-as-code workflows to introspect the registry itself.

## Use Cases

1. **Documentation Generation** - Pull resource schemas and docs into your own systems
2. **Compliance Checking** - Verify which provider/module versions are available
3. **Dependency Analysis** - Understand what resources a provider offers
4. **Tooling Integration** - Build tools that need registry metadata
5. **Version Pinning Automation** - Discover latest versions programmatically

## Provider Configuration

```hcl
provider "registry" {
  # Optional: Custom registry URL (default: registry.terraform.io)
  host = "registry.terraform.io"

  # Optional: API token for private registry
  token = var.registry_token
}
```

## Data Sources

### 1. `registry_provider`

Get details about a provider.

```hcl
data "registry_provider" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "aws_provider_id" {
  value = data.registry_provider.aws.id
}

output "aws_description" {
  value = data.registry_provider.aws.description
}
```

**Attributes:**
- `id` - Full provider ID (namespace/name)
- `namespace` - Provider namespace
- `name` - Provider name
- `description` - Provider description
- `source` - Source repository URL
- `tier` - Provider tier (official, partner, community)
- `published_at` - Publication date
- `downloads` - Total download count
- `latest_version` - Latest version string

### 2. `registry_provider_versions`

List available versions for a provider.

```hcl
data "registry_provider_versions" "aws" {
  namespace = "hashicorp"
  name      = "aws"
}

output "aws_versions" {
  value = data.registry_provider_versions.aws.versions
}

output "latest_aws" {
  value = data.registry_provider_versions.aws.versions[0].version
}
```

**Attributes:**
- `versions` - List of version objects:
  - `version` - Version string
  - `protocols` - Supported protocol versions
  - `platforms` - Available platforms (os/arch)
  - `published_at` - Publication date

### 3. `registry_provider_schema`

Get the schema for a specific provider version.

```hcl
data "registry_provider_schema" "aws" {
  namespace = "hashicorp"
  name      = "aws"
  version   = "5.0.0"
}

# Get all resource types
output "aws_resources" {
  value = keys(data.registry_provider_schema.aws.resource_schemas)
}

# Get specific resource schema
output "aws_instance_schema" {
  value = data.registry_provider_schema.aws.resource_schemas["aws_instance"]
}
```

**Attributes:**
- `provider_schema` - Provider block schema
- `resource_schemas` - Map of resource type to schema
- `data_source_schemas` - Map of data source type to schema
- `functions` - Map of function name to signature (TF 1.8+)

**Schema Object:**
- `attributes` - Map of attribute name to attribute schema
- `block_types` - Map of nested block name to block schema
- `description` - Block description
- `description_kind` - Description format (plain, markdown)

**Attribute Schema:**
- `type` - Attribute type (string, number, bool, list, map, set, object)
- `description` - Attribute description
- `required` - Is attribute required
- `optional` - Is attribute optional
- `computed` - Is attribute computed
- `sensitive` - Is attribute sensitive
- `deprecated` - Deprecation message if deprecated

### 4. `registry_module`

Get details about a module.

```hcl
data "registry_module" "vpc" {
  namespace = "terraform-aws-modules"
  name      = "vpc"
  provider  = "aws"
}

output "vpc_source" {
  value = data.registry_module.vpc.source
}
```

**Attributes:**
- `id` - Full module ID
- `namespace` - Module namespace
- `name` - Module name
- `provider` - Target provider
- `description` - Module description
- `source` - Source repository
- `version` - Latest version
- `published_at` - Publication date
- `downloads` - Download count
- `verified` - Is module verified

### 5. `registry_module_versions`

List module versions.

```hcl
data "registry_module_versions" "vpc" {
  namespace = "terraform-aws-modules"
  name      = "vpc"
  provider  = "aws"
}
```

### 6. `registry_search`

Search the registry for providers or modules.

```hcl
data "registry_search" "kubernetes" {
  query = "kubernetes"
  type  = "providers"  # or "modules"

  # Optional filters
  namespace = "hashicorp"
  tier      = "official"
}

output "k8s_providers" {
  value = data.registry_search.kubernetes.results
}
```

**Arguments:**
- `query` - Search query string
- `type` - Search type: "providers" or "modules"
- `namespace` - Filter by namespace
- `tier` - Filter by tier (providers only)
- `provider` - Filter by provider (modules only)
- `verified` - Filter verified only (modules only)
- `limit` - Max results (default: 20)
- `offset` - Pagination offset

**Attributes:**
- `results` - List of matching items
- `total_count` - Total matching results

### 7. `registry_provider_docs`

Get documentation for a provider resource.

```hcl
data "registry_provider_docs" "aws_instance" {
  namespace = "hashicorp"
  name      = "aws"
  version   = "5.0.0"

  doc_type  = "resources"  # resources, data-sources, guides
  doc_name  = "instance"
}

output "aws_instance_docs" {
  value = data.registry_provider_docs.aws_instance.content
}
```

**Attributes:**
- `content` - Documentation content (markdown)
- `title` - Document title
- `category` - Document category
- `subcategory` - Document subcategory

### 8. `registry_policy_library` (Sentinel/OPA)

Get policy library information.

```hcl
data "registry_policy_library" "cis" {
  namespace = "hashicorp"
  name      = "cis-policy-library"
}
```

## Implementation Notes

### API Endpoints

Base URL: `https://registry.terraform.io/v1/` or `/v2/`

Key endpoints:
- `GET /providers/{namespace}/{name}` - Provider details
- `GET /providers/{namespace}/{name}/versions` - Provider versions
- `GET /providers/{namespace}/{name}/{version}/download/{os}/{arch}` - Download info
- `GET /modules/{namespace}/{name}/{provider}` - Module details
- `GET /modules/{namespace}/{name}/{provider}/versions` - Module versions
- `GET /search` - Search endpoint

### Schema API

Provider schemas require fetching the provider binary and extracting schema via `terraform providers schema -json`. Alternative: use the registry's schema endpoint if available.

### Rate Limiting

- Implement request throttling
- Cache responses where appropriate
- Support conditional requests (ETag/If-None-Match)

### Error Handling

- 404: Provider/module not found
- 429: Rate limited
- 503: Registry unavailable

## Project Structure

```
terraform-provider-registry/
├── main.go
├── internal/
│   ├── provider/
│   │   └── provider.go
│   └── datasources/
│       ├── provider.go
│       ├── provider_versions.go
│       ├── provider_schema.go
│       ├── module.go
│       ├── module_versions.go
│       ├── search.go
│       ├── provider_docs.go
│       └── policy_library.go
├── examples/
│   ├── provider/
│   ├── data-sources/
│   └── README.md
├── docs/
│   ├── index.md
│   └── data-sources/
├── .goreleaser.yml
├── .github/
│   └── workflows/
├── CLAUDE.md
└── .beads/
```

## Development Phases

### Phase 1: Core Provider Infrastructure
- Provider configuration
- HTTP client with rate limiting
- Basic error handling

### Phase 2: Provider Data Sources
- `registry_provider`
- `registry_provider_versions`
- `registry_search` (providers)

### Phase 3: Module Data Sources
- `registry_module`
- `registry_module_versions`
- `registry_search` (modules)

### Phase 4: Schema Support
- `registry_provider_schema`
- Schema parsing and normalization

### Phase 5: Documentation
- `registry_provider_docs`
- Markdown content retrieval

### Phase 6: Policy Libraries
- `registry_policy_library`
- Sentinel/OPA support

## Success Metrics

- All data sources implemented with tests
- Published to Terraform Registry
- Documentation complete
- Examples for each data source
