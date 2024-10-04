resource "neon_project" "example" {
  name = "foo"
}

resource "neon_branch" "example" {
  project_id = neon_project.example.id
  name       = "bar"
}

### create a protected branch
resource "neon_branch" "example" {
  project_id = neon_project.example.id
  name       = "bar"
  protected  = "yes"
}

### create a branch off of a parent branch
resource "neon_branch" "parent" {
  project_id = neon_project.example.id
  name       = "foo"
}

resource "neon_branch" "child" {
  project_id = neon_project.example.id
  parent_id  = neon_branch.parent.id
  name       = "bar"
}
