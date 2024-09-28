package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

func TestAccOrg(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	orgID := os.Getenv("ORG_ID")
	if orgID == "" {
		t.Skip("ORG_ID must be set")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	projectName := newProjectName()

	resourceDefinition := fmt.Sprintf(`resource "neon_project" "this" {
    org_id              = "%s"
	name                = "%s"
}`, orgID, projectName)

	resource.Test(
		t, resource.TestCase{
			ProviderFactories: map[string]func() (*schema.Provider, error){
				"neon": func() (*schema.Provider, error) {
					return New("0.6.0"), nil
				},
			},
			Steps: []resource.TestStep{
				{
					Config: resourceDefinition,
					Check: func(state *terraform.State) error {
						var (
							e    error
							resp neon.ListProjectsRespObj
						)
						resp, e = client.ListProjects(nil, nil, &projectName, &orgID)
						if e == nil {
							if len(resp.Projects) != 1 {
								e = fmt.Errorf(
									"project %s should have been creted in the org %s", projectName, orgID,
								)
							}
						}
						return e
					},
				},
			},
		},
	)
}
