package provider

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccResourceProject(t *testing.T) {
	resource.UnitTest(
		t, resource.TestCase{
			ProviderFactories: providerFactories,
			Steps: []resource.TestStep{
				{
					Config: testAccResourceProject,
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr("neon_project.this", "name", "foo"),
						resource.TestCheckResourceAttr("neon_project.this", "region_id", "aws-us-west-2"),
						resource.TestCheckResourceAttr("neon_project.this", "history_retention_seconds", "30"),
						resource.TestCheckResourceAttr("neon_project.this", "pg_version", "14"),
						resource.TestCheckResourceAttr("neon_project.this", "store_password", "true"),
						resource.TestCheckResourceAttr("neon_project.this", "compute_provisioner", "k8s-pod"),
					),
				},
			},
		},
	)
}

const testAccResourceProject = `
resource "neon_project" "this" {
	name      				  = "foo"
	region_id 				  = "aws-us-west-2"
	history_retention_seconds = 30
	pg_version				  = 14
}`
