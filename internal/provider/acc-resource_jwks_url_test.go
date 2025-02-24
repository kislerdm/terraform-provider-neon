package provider

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

func TestAccJwksUrl(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectName := newProjectName()
	var projectID string
	t.Cleanup(func() {
		if _, err := client.DeleteProject(projectID); err != nil {
			log.Printf("could not clean project: %v\n", err)
		}
	})

	t.Run("Stack as IdP", func(t *testing.T) {
		// Note that Neon verifies the URL upon provisioning, hence the Stack project must exist.
		// Dmitry Kisler's Stack project ID.
		idpProjectID := "527b63cb-1552-429a-af47-29518c184629"
		wantJwksUrl := fmt.Sprintf("https://api.stack-auth.com/api/v1/projects/%s/.well-known/jwks.json", idpProjectID)
		wantRoleName := "foo"
		resourceDefinition := fmt.Sprintf(`resource "neon_project" "_" { 
	name = "%s"
	branch {role_name = "%s"}
}
resource "neon_jwks_url" "_" {
	project_id    = neon_project._.id
	role_names    = [neon_project._.database_user]
	provider_name = "Stack"
	jwks_url      = "%s"
	depends_on    = [neon_project._]
}`, projectName, wantRoleName, wantJwksUrl)
		const resourceName = "neon_jwks_url._"

		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: resourceDefinition,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(resourceName, "role_names.#", "1"),
							resource.TestCheckResourceAttr(resourceName, "role_names.0", wantRoleName),
							resource.TestCheckResourceAttr(resourceName, "jwks_url", wantJwksUrl),
							resource.TestCheckResourceAttr(resourceName, "provider_name", "Stack"),
							resource.TestCheckResourceAttrWith(resourceName, "project_id", func(v string) error {
								var er error
								if v == "" {
									er = errors.New("project_id must be set")
								}
								return er
							}),
							func(_ *terraform.State) error {
								resp, er := client.ListProjects(nil, nil, &projectName, nil, nil)
								if er == nil {
									for _, pr := range resp.Projects {
										if projectName == pr.Name {
											projectID = pr.ID
											break
										}
									}
									if projectID == "" {
										er = fmt.Errorf("no project found with name %s", projectName)
									}
								}
								return er
							},
						),
					},
					{
						Config:       resourceDefinition,
						ResourceName: resourceName,
						ImportState:  true,
						ExpectError: regexp.MustCompile(
							"the resource does not support import, please recreate it instead",
						),
					},
				},
			})
	})

	t.Run("Unknown IdP provider", func(t *testing.T) {
		resourceDefinition := fmt.Sprintf(`resource "neon_project" "_" {	name = "%s" }
resource "neon_jwks_url" "_" {
	project_id    = neon_project._.id
	role_names    = [neon_project._.database_user]
	provider_name = "foo"
	jwks_url      = "https://bar.com"
	depends_on    = [neon_project._]
}`, projectName)
		const resourceName = "neon_jwks_url._"

		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config:      resourceDefinition,
						ExpectError: regexp.MustCompile(`.*`),
					},
				},
			})
	})
}
