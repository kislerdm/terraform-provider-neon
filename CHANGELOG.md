# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v0.13.0] - 2026-01-02

### Added

- Added the following attributes to the resource `neon_project`: 
  - `block_public_connections`;
  - `block_vpc_connections`.

### Fixed

- [[191](https://github.com/kislerdm/terraform-provider-neon/issues/191)] Fixed the import logic for the resource 
  `neon_vpc_endpoint_assignment`.
- Fixed documentation for the resource `neon_project_permission`.
- [[193](https://github.com/kislerdm/terraform-provider-neon/issues/193)] Fixed the import logic for the resource
  `neon_endpoint`.
- [[198](https://github.com/kislerdm/terraform-provider-neon/issues/198)] Fixed the import logic for the resource
  `neon_branch`. 

### Changed

- **[BREAKING]** Changed the `id` attribute of the `neon_project_permission`. It's identical to the identifier assigned 
  by Neon now.

## [v0.12.0] - 2025-11-04

### Added

- [[#126](https://github.com/kislerdm/terraform-provider-neon/issues/126)] Added the block `maintenance_window` to the `neon_project` resource to configure the [Neon maintenance window](https://neon.com/docs/manage/updates#updates-on-paid-plans).
- Added examples to the documentation of the resource `neon_org_api_key`.

## [v0.11.0] - 2025-10-26

### Added

- [[#148](https://github.com/kislerdm/terraform-provider-neon/issues/148)] Added the resource `neon_org_api_key` to manage API keys on per org. level.

### Changed

- Updated dependencies:
  - Neon Go SDK: [v0.16.0](https://github.com/kislerdm/neon-sdk-go/compare/v0.13.0...v0.16.0)

## [v0.10.0] - 2025-10-12

### Removed

- [[#157](https://github.com/kislerdm/terraform-provider-neon/issues/157)] Removed the client-side validation of the autoscaling limit input.
- Removed client-side validation of the Postgres version input.

### Fixed

- [[#154](https://github.com/kislerdm/terraform-provider-neon/issues/154)], [[#156](https://github.com/kislerdm/terraform-provider-neon/issues/156)] Fixed projects listing when importing resources.   
- [[#166](https://github.com/kislerdm/terraform-provider-neon/issues/166)] Fixed the `neon_project` resource validation to configure the default endpoint to never suspend.
- [[#179](https://github.com/kislerdm/terraform-provider-neon/issues/179)] Fixed the `neon_project` resource diff management when the `org_id` is attribute is not set in the terraform module.

## [v0.9.0] - 2025-02-25

### Added

- Added the resources `neon_vpc_endpoint_assignment` and `neon_vpc_endpoint_restriction` to manage Neon Private Networking.

## [v0.8.0] - 2025-02-24

### Added

- [[#144](https://github.com/kislerdm/terraform-provider-neon/issues/144)] Added the resource `neon_jwks_url` 
  to manage the JWKs URL provided by the 3rd-party IdP required to establish [Neon RLS authorization](https://neon.tech/docs/guides/neon-rls-authorize).

### Changed

- Updated the Go version: 1.23.6 from 1.22.7.

## [v0.7.1] - 2025-02-19

### Fixed

- Fixed the note about the endpoints types.

### Changed

- Updated dependencies:
  - Neon Go SDK: [v0.13.0](https://github.com/kislerdm/neon-sdk-go/compare/v0.12.0...v0.13.0)

## [v0.7.0] - 2025-02-19

### Added

- Added the "User-Agent" header injected to every request to the Neon API for tracking purposes as agreed with
  James Broadhead from Neon.
- Added the resource `neon_api_key` to manage Neon API keys.
- Added the output attributes `connection_uri_pooler` and `database_host_pooler` to the resource `neon_project` to connect to the default database in the pooler mode. 

### Fixed

- [[#119](https://github.com/kislerdm/terraform-provider-neon/issues/119)] Fixed the output attribute `host` of the
  resource `neon_endpoint`: it will yield the correct URI for the endpoints with the
  [pooled mode](https://neon.tech/docs/connect/connection-pooling#how-to-use-connection-pooling) activated.
- [[#137](https://github.com/kislerdm/terraform-provider-neon/issues/137)] Fixed operations execution management 
  by introducing await mechanism to wait until the running operations finish.
- [[#133](https://github.com/kislerdm/terraform-provider-neon/issues/133)] Fixed state management when configuring network security.
- Documentation improvements:
  - Removed unclear warning from the page for the `neon_endpoint` resource.

### Removed

- Removed the outdated warning which was showing upon creation of the endpoint of the type "read_only",
- Removed the attributes "allowed_ips_primary_branch_only" from the resource `neon_project` because it's no longer supported by Neon.

### Changed

- Updated dependencies:
  - Neon Go SDK: [v0.12.0](https://github.com/kislerdm/neon-sdk-go/compare/v0.6.1...v0.12.0)
  - github.com/hashicorp/terraform-plugin-docs: [v0.20.1](https://github.com/hashicorp/terraform-plugin-docs/compare/v0.19.4...v0.20.1)
  - github.com/hashicorp/terraform-plugin-sdk: [v2.36.0](https://github.com/hashicorp/terraform-plugin-sdk/compare/v2.33.0...v2.36.0)
- **[BREAKING]** [#113](https://github.com/kislerdm/terraform-provider-neon/issues/113)] Set the default retention window to 1 day to avoid inconsistency with Neon.

## [v0.6.3] - 2024-10-05

### Fixed

- Fixed docu; non-functional change.

## [v0.6.2] - 2024-10-04

### Added

- Added the attribute `protected` for the resource `neon_branch` to provision protected branches.

### Fixed

- [[#108](https://github.com/kislerdm/terraform-provider-neon/issues/108)] Fixed the import behaviour for the resource `neon_role`.
- Fixed mutability of the default branch by adjusting the behaviour for the `branch` state of the resource `neon_project`.

### Changed

- Updated dependencies:
  - Neon Go SDK: [v0.6.1](https://github.com/kislerdm/neon-sdk-go/compare/v0.5.0...v0.6.1)

## [v0.6.1] - 2024-09-28

### Added

- Added support of Postgres 17. See the Neon [announcement](https://neon.tech/blog/postgres-17) and the Postgres 
  [announcement](https://www.postgresql.org/about/news/postgresql-17-released-2936/).

### Fixed

- Fixed validation of the autoscalling limits. You can now set the maximum compute size up to `10`.

## [v0.6.0] - 2024-09-23

### Added

- [[#99](https://github.com/kislerdm/terraform-provider-neon/issues/99)] Added the attribute `org_id` to
  the resource `neon_project` to create project in the organisation.

### Fixed

- **[BREAKING]** [[#96](https://github.com/kislerdm/terraform-provider-neon/issues/96)] The boolean attributes of the 
  resource `neon_project` will be treated as strings to work around the 
  [issue](https://github.com/hashicorp/terraform-plugin-sdk/issues/817) with state management when the attribute gets
  removed from the manifest.

**Examples**

- Set allowed_ips to be applicable only to the primary branch:
  ```terraform
   resource "neon_project" "this" {
      name = "myproject"
      
      allowed_ips = ["1.2.3.4/24"]

      allowed_ips_primary_branch_only = "yes"
   }
  ``` 
- Set allowed_ips to be applicable to all branches, explicitly:
  ```terraform
   resource "neon_project" "this" {
      name = "myproject"
      
      allowed_ips = ["1.2.3.4/24"]

      allowed_ips_primary_branch_only = "no"
   }
  ```
- Set allowed_ips to be applicable to all branches, implicitly:
  ```terraform
   resource "neon_project" "this" {
      name = "myproject"
      
      allowed_ips = ["1.2.3.4/24"]
   }
  ```

### Changed

- Updated dependencies:
  - Neon Go SDK: [v0.5.0](https://github.com/kislerdm/neon-sdk-go/compare/v0.4.7...v0.5.0)
  - github.com/hashicorp/terraform-plugin-docs: [v0.19.4](https://github.com/hashicorp/terraform-plugin-docs/compare/v0.19.0...v0.19.4)

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
