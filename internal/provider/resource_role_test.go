package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceRole(t *testing.T) {
	resource.UnitTest(
		t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					PreConfig: nil,
					Taint:     nil,
					Config:    testAccResourceRole,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_role.foo", "name", "qux"),
						resource.TestCheckResourceAttr("neon_role.foo", "password", "Onf1AjayKwe0"),
					),
				},
			},
		},
	)
}

const testAccResourceRole = `resource "neon_role" "foo" {
	project_id = "shiny-wind-028834"
	branch_id  = "br-noisy-sunset-458773"
	name 	   = "qux"
}`
