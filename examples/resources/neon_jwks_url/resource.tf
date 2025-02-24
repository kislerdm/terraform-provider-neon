# provision Stack's JWKs URL for the default role.
resource "neon_project" "example" {
  name = "foo"
}

resource "neon_jwks_url" "default" {
  project_id    = neon_project.example.id
  role_names    = [neon_project.example.database_user]
  provider_name = "Stack"
  # replace the URL with the one which corresponds to your Stack project
  # see details: https://neon.tech/docs/guides/neon-rls-authorize-stack-auth
  jwks_url   = "https://api.stack-auth.com/api/v1/projects/e3475923-a0b3-4cbb-a70f-b3071985a11d/.well-known/jwks.json"
  depends_on = [neon_project.example]
}

# provision Stack's JWKs URL for the custom branch and the custom roles.
resource "neon_branch" "custom" {
  project_id = neon_project.example.id
  name       = "bar"
}

locals {
  custom_roles = ["r0", "r1", "r2"]
}

resource "neon_role" "custom" {
  for_each   = toset(local.custom_roles)
  project_id = neon_project.example.id
  branch_id  = neon_branch.custom.id
  name       = each.key
}

resource "neon_jwks_url" "custom" {
  project_id    = neon_project.example.id
  branch_id     = neon_branch.custom.id
  role_names    = local.custom_roles
  provider_name = "Stack"
  # replace the URL with the one which corresponds to your Stack project
  # see details: https://neon.tech/docs/guides/neon-rls-authorize-stack-auth
  jwks_url   = "https://api.stack-auth.com/api/v1/projects/e3475923-a0b3-4cbb-a70f-b3071985a11d/.well-known/jwks.json"
  depends_on = [neon_project.example, neon_branch.custom, neon_role.custom]
}
