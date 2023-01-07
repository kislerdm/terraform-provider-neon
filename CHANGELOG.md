# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.0.4] - 2023-01-07

### Added

- Endpoint resource:

```terraform
resource "neon_endpoint" "this" {
  project_id = "bitter-meadow-966132"
  branch_id  = "br-floral-mountain-251143"
  type       = "read_write"
}
```

## [0.0.3] - 2023-01-07

### Added

- Added database connection details for the resource `neon_project`. **Note** that `database_password`
  and `connection_uri` read-only attributes are sensitive. Make sure that terraform backend is secured and
  terraform state is not exposed to prevent undesirable access to database.

## [0.0.2] - 2023-01-07

### Added

- Branch resource:

```terraform
resource "neon_project" "this" {
  name = "foo"
}

resource "neon_branch" "this" {
  project_id = neon_project.this.id
  name       = "bar"
}
```

- Backoff+retry mechanism: operation is retried after the delay of 5 sec. API response's HTTP codes are 500, or 429.
  Total number of attempts is limited to 120 per operation.

### Changed

- Bumped [Neon Go SDK](https://pkg.go.dev/github.com/kislerdm/neon-sdk-go) v0.1.3
- Added errors handling for the project resource

## [0.0.1] - 2023-01-05

### Added

- `Neon` Provider:

```terraform
terraform {
  required_providers {
    neon = {
      source = "kislerdm/neon"
    }
  }
}

provider "neon" {}
```

- Project resource:

```terraform
resource "neon_project" "this" {
  name = "foo"
}
```
