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
					),
				},
			},
		},
	)
}

const testAccResourceEndpoint = `resource "neon_endpoint" "foo" {
	project_id = "spring-example-302709"
	branch_id  = "foo"
	type 	   = "read_write"
}`
