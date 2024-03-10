# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.5.0] - 2024-03-10

### Added

- [[#22](https://github.com/kislerdm/terraform-provider-neon/issues/22)] Added the following data resources:
  - `neon_project`
  - `neon_branches`
  - `neon_branch_endpoints`
  - `neon_branch_roles`
  - `neon_branch_role_password`

- Added the read-only attribute `default_endpoint_id` to the resource `neon_project`.
- Added the retry logic to manage all supported resources:
  - `neon_project`
  - `neon_branch`
  - `neon_endpoint`
  - `neon_role`
  - `neon_database`
  - `neon_project_permission`

### Fixed

- [[#83](https://github.com/kislerdm/terraform-provider-neon/issues/83)] Fixed the state management of the project's 
default branch, role, database and endpoint.
- [[#88](https://github.com/kislerdm/terraform-provider-neon/issues/88)] Fixed import of the resource `neon_role`.

### Changed

- Updated dependencies:
  - Neon Go SDK [v0.4.7](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.4.7)
- Reduced the retry delay to _1 second_ from _5 seconds_.

## [v0.4.1] - 2024-02-28

### Changed

- Updated dependencies:
    - Neon Go SDK [v0.4.6](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.4.6)
    - terraform-plugin-sdk [v2.33.0](https://github.com/hashicorp/terraform-plugin-sdk/releases/tag/v2.33.0)
    - Google UUID module [v1.6.0](https://github.com/google/uuid/releases/tag/v1.6.0)

## [v0.4.0] - 2024-01-28

### Added

- Added the resource `neon_project_permission` to manage the project's permissions.

### Changed

- Updated dependencies: 
  - Neon Go SDK [v0.4.3](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.4.3)
  - Terraform docs [v0.18.0](https://github.com/hashicorp/terraform-plugin-docs/releases/tag/v0.18.0)

## [v0.3.2] - 2024-01-11

### Fixed

- [`resource_project>region_id`] Validation of the project deployment region was removed.

### Changed

- Updated dependencies: Neon Go SDK [v0.4.2](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.4.2)

## [v0.3.1] - 2023-12-22

### Added

- `resource_project` includes the attribute `enable_logical_replication` to configure the [logical replication](https://neon.tech/docs/introduction/logical-replication).

### Fixed

- PostgreSQL 16 is now supported.

### Changed

- Updated dependencies: Neon Go SDK [v0.4.1](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.4.1)

## [v0.3.0] - 2023-12-21

### Added

- `resource_project` includes two additional attributes to configure IP addresses allowed to connect to the project's 
endpoints:
  - `allowed_ips`
  - `allowed_ips_primary_branch_only`

### Fixed

- Schema is set on per resource basis now.

### Changed

- Updated dependencies: 
  - Go version to 1.21
  - Neon Go SDK [v0.4.0](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.4.0)
  - Terraform plugin SDK [v2.31.0](https://github.com/hashicorp/terraform-plugin-sdk/releases/tag/v2.31.0)

## [v0.2.5] - 2023-11-02

### Fixed

- Fixed validation of the [branch ID](https://neon.tech/docs/manage/branches#view-branches).

## [v0.2.4] - 2023-10-26

### Fixed

- Minor documentation fixes: 
  - The note in the `resource_role` is removed because it's not reflecting the provider's behaviour.
  - The logo is fixed.

### Changed

- Updated dependencies: Neon Go SDK [v0.3.0](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.3.0)

## [v0.2.3] - 2023-10-22

### Fixed

- [[#51](https://github.com/kislerdm/terraform-provider-neon/issues/51)] Fixed credentials content.
- Fixed management of the `resource_role` state:
  - Fixed password reading.
  - Removed the side effect upon the resource import: the role's password won't be reset now. 

### Changed

- Updated dependencies: Neon Go SDK [v0.2.5](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.2.5)

## [v0.2.2] - 2023-09-19

### Fixed

- `resource_project`:
    - ([#40](https://github.com/kislerdm/terraform-provider-neon/issues/40)) Fixed types conversion for the terraform
      resource attributes of the
      type [`TypeInt`](https://github.com/hashicorp/terraform-plugin-sdk/blob/af738e0d482f699504d9e35e134766da459ef1f6/helper/schema/schema.go#L55).
    - ([#42](https://github.com/kislerdm/terraform-provider-neon/issues/42)) Fixed default branch configuration.
    - ([#48](https://github.com/kislerdm/terraform-provider-neon/issues/48)) Fixed default endpoint settings
      configuration.
    - Fixed history retention configuration. Now, the retention period of 7 days will be set by default, and zero will
      be set if `history_retention_seconds` is set to zero explicitly.

**Note** web console will reflect the data retention period correctly only if `history_retention_seconds` was set to an
integral number of days because the web console shows the total number of full days only. Moreover, "7 days - default"
will be displayed for any `history_retention_seconds` value below 86400 (1 day).

| history_retention_seconds | web console      | human-friendly duration |
|:--------------------------|:-----------------|:------------------------|
| 0                         | 7 days - default | 0                       |
| 300                       | 7 days - default | 5 min                   |
| 3600                      | 7 days - default | 1 hour                  |
| 43200                     | 7 days - default | 12 hours                |
| 164160                    | 1 day            | 1.9 days                |
| 198720                    | 2 days           | 2.3 days                |
| 604800                    | 7 days - default | 7 days                  |

- Documentation:
    - Link to the Neon logo
    - End-to-end example

### Changed

- Documentation: examples of project provisioning
- Updated dependencies: Neon Go SDK [v0.2.2](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.2.2),
  terraform plugin SDK to 2.29.0

## [0.2.1] - 2023-08-07

### Added

- Acceptance e2e tests
- `resource_project`:
    - default_branch_id

- `resource_endpoint`:
    - id
    - compute_provisioner
    - suspend_timeout_seconds

### Removed

- `resource_role`:
    - created_at
    - updated_at

- `resource_endpoint`:
    - passwordless_access: it's not implemented yet by Neon
    - current_state
    - pending_state

- `resource_branch`:
    - connection_uri

### Changed

- `resource_endpoint`:
    - autoscaling_limit_min_cu set to 0.25 by default
    - autoscaling_limit_max_cu set to 0.25 by default
    - type set to "read_write" by default

## [0.2.0] - 2023-08-04

The release follows update of the [Neon Go SDK](https://github.com/kislerdm/neon-sdk-go/releases/tag/v0.2.1).

### Fixed

- ([#25](https://github.com/kislerdm/terraform-provider-neon/issues/25)) Fixed branch import
- ([#26](https://github.com/kislerdm/terraform-provider-neon/issues/26)) Fixed database import
- ([#32](https://github.com/kislerdm/terraform-provider-neon/issues/32)) Data type to define autoscaling limits
- Neon logo in documentation

### Added

- `resource_project`:
    - store_password (_Note that Neon does not support "false" value yet_)
    - history_retention_seconds
    - compute_provisioner
    - quota:
        - active_time_seconds
        - compute_time_seconds
        - written_data_bytes
        - data_transfer_bytes
        - logical_size_bytes
    - default_endpoint_settings:
        - autoscaling_limit_min_cu
        - autoscaling_limit_max_cu
        - suspend_timeout_seconds
    - branch:
        - id
        - name
        - role_name
        - database_name

- `resource_branch`:
    - id
    - connection_uri

### Removed

- `resource_project`:
    - pg_settings
    - cpu_quota_sec
    - autoscaling_limit_min_cu
    - autoscaling_limit_max_cu
    - branch_logical_size_limit
    - created_at
    - updated_at

- `resource_branch`:
    - physical_size_size
    - endpoint
    - host
    - current_state
    - pending_state
    - created_at
    - updated_at

- `resource_database`:
    - created_at
    - updated_at

- `resource_endpoint`:
    - created_at
    - updated_at

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
