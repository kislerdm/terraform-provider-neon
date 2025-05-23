---
page_title: "neon_project Resource - terraform-provider-neon"
description: |-
  Neon Project.

See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/
API: https://api-docs.neon.tech/reference/createproject
---

# neon_project (Resource)

Neon Project.

See details: https://neon.tech/docs/get-started-with-neon/setting-up-a-project/
API: https://api-docs.neon.tech/reference/createproject

## Example Usage

```terraform
### Default

resource "neon_project" "example" {
  name = "foo"
}

### Turn off data retention

resource "neon_project" "example" {
  name                      = "foo"
  history_retention_seconds = 0
}

### Set custom compute limits

resource "neon_project" "example" {
  name = "foo"

  default_endpoint_settings {
    autoscaling_limit_min_cu = 0.5
    autoscaling_limit_max_cu = 1
    suspend_timeout_seconds  = 10
  }
}

### Define custom default branch

resource "neon_project" "example" {
  name = "foo"

  branch {
    name          = "bar"
    database_name = "baz"
    role_name     = "qux"
  }
}

### Set the logical replication
# See: https://neon.tech/docs/guides/logical-replication-guide
resource "neon_project" "example_with_logical_replication" {
  name = "my-project-with-logical-replication"

  enable_logical_replication = "yes"
}

### Set the allow list of IP addresses
# Note that the feature is only available to the users of the Business plan:
# https://neon.tech/docs/introduction/ip-allow
resource "neon_project" "example_with_allowed_ips" {
  name = "my-project-with-allowed-list-of-ips"

  allowed_ips = ["1.2.3.4/24", "99.1.20.93"]
}

### Set the allow list of IP addresses for protected branches only
# Note that the feature is only available to the users of the Business, or Scale plans:
# https://neon.tech/docs/guides/protected-branches
resource "neon_project" "example_with_allowed_ips_protected_branch_only" {
  name = "my-project-with-allowed-list-of-ips-for-protected-branch"

  allowed_ips                         = ["1.2.3.4/24", "99.1.20.93"]
  allowed_ips_protected_branches_only = "yes"
}

### Create project in the organisation
resource "neon_project" "example_in_org" {
  name   = "myproject"
  org_id = "org-restless-silence-28866559"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Optional

- `allowed_ips` (List of String) A list of IP addresses that are allowed to connect to the endpoints.
Note that the feature is available to the Neon Scale plans only. Details: https://neon.tech/docs/manage/projects#configure-ip-allow
- `allowed_ips_protected_branches_only` (String) Set to 'yes' to activate, 'no' to deactivate explicitly, and omit to keep the default value.
Apply the allow-list to the protected branches only.
Note that the feature is available to the Neon Scale plans only.
- `branch` (Block List, Max: 1) (see [below for nested schema](#nestedblock--branch))
- `compute_provisioner` (String) Provisioner The Neon compute provisioner.
Specify the k8s-neonvm provisioner to create a compute endpoint that supports Autoscaling.
- `default_endpoint_settings` (Block List, Max: 1) (see [below for nested schema](#nestedblock--default_endpoint_settings))
- `enable_logical_replication` (String) Set to 'yes' to activate, 'no' to deactivate explicitly, and omit to keep the default value.
Sets wal_level=logical for all compute endpoints in this project.
All active endpoints will be suspended. Once enabled, logical replication cannot be disabled.
See details: https://neon.tech/docs/introduction/logical-replication
- `history_retention_seconds` (Number) The number of seconds to retain the point-in-time restore (PITR) backup history for this project.
Default: 1 day, see https://neon.tech/docs/reference/glossary#point-in-time-restore.
- `name` (String) Project name.
- `org_id` (String) Identifier of the organisation to which this project belongs.
- `pg_version` (Number) Postgres version
- `quota` (Block List, Max: 1) Per-project consumption quota. If the quota is exceeded, all active computes
are automatically suspended and it will not be possible to start them with
an API method call or incoming proxy connections. The only exception is
logical_size_bytes, which is applied on per-branch basis, i.e., only the
compute on the branch that exceeds the logical_size quota will be suspended.

Quotas are enforced based on per-project consumption metrics with the same names,
which are reset at the end of each billing period (the first day of the month).
Logical size is also an exception in this case, as it represents the total size
of data stored in a branch, so it is not reset.

The zero value per attributed means 'unlimited'. (see [below for nested schema](#nestedblock--quota))
- `region_id` (String) Deployment region: https://neon.tech/docs/introduction/regions
- `store_password` (String) Set to 'yes' to activate, 'no' to deactivate explicitly, and omit to keep the default value.
Whether or not passwords are stored for roles in the Neon project.
Storing passwords facilitates access to Neon features that require authorization.

### Read-Only

- `connection_uri` (String, Sensitive) Default connection uri. **Note** that it contains access credentials.
- `connection_uri_pooler` (String, Sensitive) Default connection uri with the traffic via pooler. **Note** that it contains access credentials.
- `database_host` (String) Default database host.
- `database_host_pooler` (String) Default endpoint host via pooler.
- `database_name` (String) Default database name.
- `database_password` (String, Sensitive) Default database access password.
- `database_user` (String) Default database role.
- `default_branch_id` (String) Default branch ID.
- `default_endpoint_id` (String) Default endpoint ID.
- `id` (String) Project ID.

<a id="nestedblock--branch"></a>
### Nested Schema for `branch`

Optional:

- `database_name` (String) The name of the default database provisioned upon creation of new project. It's owned by the default role (`role_name`).
If not specified, the default database name will be used.
- `name` (String) The name of the default branch provisioned upon creation of new project.
If not specified, the default branch name will be used.
- `role_name` (String) The name of the default role provisioned upon creation of new project.
If not specified, the default role name will be used.

Read-Only:

- `id` (String) Branch ID.


<a id="nestedblock--default_endpoint_settings"></a>
### Nested Schema for `default_endpoint_settings`

Optional:

- `autoscaling_limit_max_cu` (Number)
- `autoscaling_limit_min_cu` (Number)
- `suspend_timeout_seconds` (Number) Duration of inactivity in seconds after which the compute endpoint is automatically suspended.
The value 0 means use the global default.
The value -1 means never suspend. The default value is 300 seconds (5 minutes).
The maximum value is 604800 seconds (1 week)

Read-Only:

- `id` (String) Endpoint ID.


<a id="nestedblock--quota"></a>
### Nested Schema for `quota`

Optional:

- `active_time_seconds` (Number) The total amount of wall-clock time allowed to be spent by the project's compute endpoints.
- `compute_time_seconds` (Number) The total amount of CPU seconds allowed to be spent by the project's compute endpoints.
- `data_transfer_bytes` (Number) Total amount of data transferred from all of a project's branches using the proxy.
- `logical_size_bytes` (Number) Limit on the logical size of every project's branch.
- `written_data_bytes` (Number) Total amount of data written to all of a project's branches.




## Import

The Neon Project can be imported to the terraform state by its identifier.

Import using the [import block](https://developer.hashicorp.com/terraform/language/import):

For example:

```hcl
import {
  to = neon_project.example
  id = "shiny-cell-31746257"
}
```

Import using the command `terraform import`:

```commandline
terraform import neon_project.example shiny-cell-31746257
```
