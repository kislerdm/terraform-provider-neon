resource "neon_project" "example" {
  name = "foo"
}

# grant project access to the user with the email foo@bar.qux
resource "neon_project_permission" "share" {
  project_id = neon_project.example.id
  grantee    = "foo@bar.qux"
}
