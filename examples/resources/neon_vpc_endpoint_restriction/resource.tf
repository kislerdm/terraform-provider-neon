resource "neon_vpc_endpoint_assignment" "example" {
  org_id          = "org-foo-bar-01234567"
  region_id       = "us-east-1"
  vpc_endpoint_id = "vpce-1234567890abcdef0"
  label           = "example"
}

resource "neon_vpc_endpoint_restriction" "example" {
  project_id      = "cold-bread-99644485"
  vpc_endpoint_id = neon_vpc_endpoint_assignment.example.vpc_endpoint_id
  label           = "example"
}
