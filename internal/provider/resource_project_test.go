package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceProject(t *testing.T) {
	resource.UnitTest(
		t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Taint:  nil,
					Config: testAccResourceProject,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_project.foo", "id", "string"),
						resource.TestCheckResourceAttr("neon_project.foo", "name", "string"),
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

const testAccResourceProject = `resource "neon_project" "foo" {}`
