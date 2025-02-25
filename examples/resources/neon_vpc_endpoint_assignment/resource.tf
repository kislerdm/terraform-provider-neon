resource "neon_vpc_endpoint_assignment" "example" {
  org_id          = "org-foo-bar-01234567"
  region_id       = "aws-us-east-1"
  vpc_endpoint_id = "vpce-1234567890abcdef0"
  label           = "example"
}
