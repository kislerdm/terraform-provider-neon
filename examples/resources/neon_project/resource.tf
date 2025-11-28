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

### Block public connections to the project's endpoints
# Note that the feature is only available to the users of the Scale plans:
# https://neon.tech/docs/introduction/ip-allow
resource "neon_project" "example_with_blocked_public_connections" {
  name = "my-project-with-blocked-public-connections"

  block_public_connections = "yes"
}

### Create project in the organisation
resource "neon_project" "example_in_org" {
  name   = "myproject"
  org_id = "org-restless-silence-28866559"
}

### Set custom maintenance window
# Note that the feature is only available to the users of non-Free plan
# https://neon.com/docs/manage/updates
resource "neon_project" "custom_maintenance_window" {
  name = "myproject"
  maintenance_window {
    weekdays   = [6, 7]
    start_time = "07:00"
    end_time   = "08:00"
  }
}
