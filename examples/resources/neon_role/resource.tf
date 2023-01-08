resource "neon_project" "example" {
  name = "foo"
}

resource "neon_branch" "example" {
  project_id = neon_project.example.id
  name       = "bar"
}

resource "neon_role" "example" {
  project_id = neon_project.example.id
  branch_id  = neon_branch.example.id
  name       = "qux"
}
