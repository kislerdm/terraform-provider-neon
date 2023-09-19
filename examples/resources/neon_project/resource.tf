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
