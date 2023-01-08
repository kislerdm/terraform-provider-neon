package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceBranch(t *testing.T) {
	resource.UnitTest(
		t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					ResourceName: "",
					PreConfig:    nil,
					Taint:        nil,
					Config:       testAccResourceBranch,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_branch.foo", "name", "dev2"),
						resource.TestCheckResourceAttr("neon_branch.foo", "parent_lsn", "0/1DE2850"),
						resource.TestCheckResourceAttr("neon_branch.foo", "current_state", "ready"),
						resource.TestCheckResourceAttr(
							"neon_branch.foo", "host", "ep-shrill-thunder-454069.us-east-2.aws.neon.tech",
						),
					),
				},
			},
		},
	)
}

const testAccResourceBranch = `resource "neon_branch" "foo" {
	project_id = "spring-example-302709"
}`
