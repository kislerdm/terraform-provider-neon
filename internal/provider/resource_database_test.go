package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceDatabase(t *testing.T) {
	resource.UnitTest(
		t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					PreConfig: nil,
					Taint:     nil,
					Config:    testAccResourceDatabase,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_database.foo", "name", "mydb"),
						resource.TestCheckResourceAttr("neon_database.foo", "owner_name", "casey"),
						resource.TestCheckResourceAttr("neon_database.foo", "project_id", "shiny-wind-028834"),
						resource.TestCheckResourceAttr("neon_database.foo", "branch_id", "br-aged-salad-637688"),
					),
				},
			},
		},
	)
}

const testAccResourceDatabase = `resource "neon_database" "foo" {
	project_id = "shiny-wind-028834"
	branch_id  = "br-aged-salad-637688"
	name 	   = "mydb"
	owner_name = "casey"
}`
