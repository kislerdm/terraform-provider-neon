package provider

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceProject(t *testing.T) {
	t.Skip("resource not yet implemented, remove this once you add your own code")

	resource.UnitTest(
		t, resource.TestCase{
			PreCheck:          func() { testAccPreCheck(t) },
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceProject,
					Check: resource.ComposeTestCheckFunc(
						resource.TestMatchResourceAttr(
							"neon_project.example", "name", regexp.MustCompile("^fo"),
						),
					),
				},
			},
		},
	)
}

const testAccResourceProject = `
resource "neon_project" "example" {
  name = "foo"
}
`
