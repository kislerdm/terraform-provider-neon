resource "neon_project" "example" {
  name = "foo"
}

resource "neon_branch" "example" {
  project_id = neon_project.example.id
  name       = "foo"
}

resource "neon_endpoint" "example" {
  project_id = neon_project.example.id
  branch_id  = neon_branch.example.id
  type       = "read_write"
}
