package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceEndpoint(t *testing.T) {
	resource.UnitTest(
		t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					ResourceName: "",
					PreConfig:    nil,
					Taint:        nil,
					Config:       testAccResourceEndpoint,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_endpoint.foo", "type", "read_write"),
						resource.TestCheckResourceAttr("neon_endpoint.foo", "type", "read_write"),
						resource.TestCheckResourceAttr("neon_endpoint.foo", "region_id", "aws-us-east-2"),
						resource.TestCheckResourceAttr("neon_endpoint.foo", "proxy_host", "us-east-2.aws.neon.tech"),
					),
				},
			},
		},
	)
}

const testAccResourceEndpoint = `resource "neon_endpoint" "foo" {
	project_id = "shiny-wind-028834"
	branch_id  = "br-aged-salad-637688"
	type 	   = "read_write"
}`
