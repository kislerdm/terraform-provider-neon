resource "neon_org_api_key" "example" {
  name   = "foo"
  org_id = "org-foo-bar-12345678"
}


### Create API key that could only be used to manage the project with the ID baz-qux-12345678
resource "neon_org_api_key" "limited_to_project" {
  name       = "foo"
  org_id     = "org-foo-bar-12345678"
  project_id = "baz-qux-12345678"
}
