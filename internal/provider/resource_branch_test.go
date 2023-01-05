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
					Taint:  nil,
					Config: testAccResourceBranch,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_branch.foo", "name", "string"),
					),
					ExpectNonEmptyPlan:        false,
					PlanOnly:                  false,
					PreventDiskCleanup:        false,
					PreventPostDestroyRefresh: false,
					Destroy:                   false,
				},
			},
		},
	)
}

const testAccResourceBranch = `resource "neon_branch" "foo" {}`
