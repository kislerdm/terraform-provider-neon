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
					Config: testAccResourceProject,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_project.foo", "id", "string"),
						resource.TestCheckResourceAttr("neon_project.foo", "name", "string"),
					),
					Destroy:                 false,
					ExpectNonEmptyPlan:      true,
					PlanOnly:                true,
					ImportStateVerifyIgnore: nil,
					ProviderFactories:       nil,
				},
			},
		},
	)
}

const testAccResourceProject = `resource "neon_project" "foo" {}`
