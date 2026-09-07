resource "neon_project" "example" {
  name = "foo"
}

resource "neon_branch_backup_schedule" "example" {
  project_id = neon_project.example.id
  branch_id  = neon_project.example.default_branch_id
  schedule = [
    {
      frequency         = "weekly"
      day               = 1
      hour              = 4
      retention_seconds = 86400 * 7
    },
    {
      frequency         = "daily"
      hour              = 4
      retention_seconds = 86400
    },
    {
      frequency         = "monthly"
      day               = 1
      hour              = 4
      retention_seconds = 86400 * 30
    },
    {
      frequency         = "monthly"
      day               = 2
      hour              = 4
      retention_seconds = 86400 * 30
    }
  ]
}
