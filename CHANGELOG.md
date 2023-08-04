# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.2.0] - 2023-08-04

The release follows update of the [Neon Go SDK](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.2.1).

### Fixed

- ([#25](https://github.com/kislerdm/terraform-provider-neon/issues/25)) Fixed branch import
- ([#26](https://github.com/kislerdm/terraform-provider-neon/issues/26)) Fixed database import
- ([#32](https://github.com/kislerdm/terraform-provider-neon/issues/32)) Data type to define autoscaling limits

## [0.1.0] - 2023-01-08

### Changed

- Fixed typo and indentation of documentation

## [0.0.9] - 2023-01-08

### Fixed

- Fixed `neon_branch` recourse by provisioning an endpoint attached to a newly created branch. It is required to permit
  interactions with the branch to manage associated roles and databases.

### Changed

- Improved documentation
- Added an end-to-end guide to provision resources for AWS application to communicate with the Neon database  

## [0.0.8] - 2023-01-08

### Added

- Database resource:

```terraform
resource "neon_database" "this" {
  project_id = "bitter-meadow-966132"
  branch_id  = "br-floral-mountain-251143"
  name       = "baz"
  owner_name = "qux"
}
```

## [0.0.7] - 2023-01-08

### Added

- Role resource:

```terraform
resource "neon_role" "this" {
  project_id = "bitter-meadow-966132"
  branch_id  = "br-floral-mountain-251143"
  name       = "qux"
}
```

## [0.0.6] - 2023-01-08

### Fixed

- Fixed `neon_endpoint` resource provisioning when the attribute `pg_settings` is not set. The bug was in the Neon SDK,
  see
  details in the [release notes](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.1.4).

### Changed

- Bumped [Neon Go SDK](https://pkg.go.dev/github.com/kislerdm/neon-sdk-go) to v0.1.4

## [0.0.5] - 2023-01-07

### Fixed

- Fixed the logic to import `neon_branch` resource by its ID.

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
