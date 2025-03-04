---
page_title: "neon_endpoint Resource - terraform-provider-neon"
description: |-
  Project Endpoint. See details: https://neon.tech/docs/manage/endpoints/
---

# neon_endpoint (Resource)

Project Endpoint. See details: https://neon.tech/docs/manage/endpoints/

## Example Usage

```terraform
resource "neon_project" "example" {
  name = "foo"
}

resource "neon_branch" "example" {
  project_id = neon_project.example.id
  name       = "bar"
}

resource "neon_endpoint" "example" {
  project_id = neon_project.example.id
  branch_id  = neon_branch.example.id
  type       = "read_write"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `branch_id` (String) Branch ID.
- `project_id` (String) Project ID.

### Optional

- `autoscaling_limit_max_cu` (Number)
- `autoscaling_limit_min_cu` (Number)
- `compute_provisioner` (String) Provisioner The Neon compute provisioner.
Specify the k8s-neonvm provisioner to create a compute endpoint that supports Autoscaling.
- `disabled` (Boolean) Disable the endpoint.
- `pg_settings` (Map of String)
- `pooler_enabled` (Boolean) Activate connection pooling.
See details: https://neon.tech/docs/connect/connection-pooling
- `pooler_mode` (String) Mode of connections pooling.
See details: https://neon.tech/docs/connect/connection-pooling
- `region_id` (String) Deployment region: https://neon.tech/docs/introduction/regions
- `suspend_timeout_seconds` (Number) Duration of inactivity in seconds after which the compute endpoint is automatically suspended.
The value 0 means use the global default.
The value -1 means never suspend. The default value is 300 seconds (5 minutes).
The maximum value is 604800 seconds (1 week)
- `type` (String) Access type. **Note** that a single branch can have only one "read_write" endpoint.

### Read-Only

- `host` (String) Endpoint URI.
- `id` (String) Endpoint ID.
- `proxy_host` (String)



## Import

The Neon Endpoint can be imported to the terraform state by its identifier.

Import using the [import block](https://developer.hashicorp.com/terraform/language/import):

For example:

```hcl
import {
  to = neon_endpoint.example
  id = "ep-black-mouse-a64dr7wp"
}
```

Import using the command `terraform import`:

```commandline
terraform import neon_endpoint.example ep-black-mouse-a64dr7wp
```
