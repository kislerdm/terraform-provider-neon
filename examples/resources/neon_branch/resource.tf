resource "neon_project" "example" {
  name = "foo"
}

resource "neon_branch" "example" {
  project_id = neon_project.example.id
  name       = "bar"
}
