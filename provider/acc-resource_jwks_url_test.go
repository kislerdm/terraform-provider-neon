package provider

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestAccJwksUrl(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix += "jwks-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	var newProjectName = func() string {
		return projectNamePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	// Note that Neon verifies the URL upon provisioning, hence the Stack project must exist.
	// Dmitry Kisler's Stack project ID.
	idpProjectID := "527b63cb-1552-429a-af47-29518c184629"
	wantJwksUrl := fmt.Sprintf("https://api.stack-auth.com/api/v1/projects/%s/.well-known/jwks.json", idpProjectID)
	wantRoleName := "foo"
	var resourceDefinition = func(projectName string) string {
		return fmt.Sprintf(`resource "neon_project" "_" { 
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
	}

	projectName := newProjectName()
	t.Run("Stack as IdP", func(t *testing.T) {
		const resourceName = "neon_jwks_url._"
		config := resourceDefinition(projectName)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: config,
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
						),
					},
					{
						Config:       config,
						ResourceName: resourceName,
						ImportState:  true,
						ExpectError: regexp.MustCompile(
							"the resource does not support import, please recreate it instead",
						),
					},
					// shall yield non-empty plan if the resource is deleted outside terraform
					// given that JWKs existed prior to deletion
					{
						PreConfig: func() {
							ref, err := readProjectInfo(client, projectName)
							if err != nil {
								panic(err)
							}
							resp, err := client.GetProjectJWKS(ref.ID)
							if err != nil {
								panic(err)
							}
							for _, jwk := range resp.Jwks {
								_, err = client.DeleteProjectJWKS(ref.ID, jwk.ID)
								if err != nil {
									panic(err)
								}
							}
						},
						RefreshState:       true,
						ExpectNonEmptyPlan: true,
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

	t.Run("shall destroy even if the resource was deleted outside of terraform,", func(t *testing.T) {
		projectName := newProjectName()
		config := resourceDefinition(projectName)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: config,
					},
					{
						PreConfig: func() {
							ref, err := readProjectInfo(client, projectName)
							if err != nil {
								panic(err)
							}
							resp, err := client.GetProjectJWKS(ref.ID)
							if err != nil {
								panic(err)
							}
							for _, jwk := range resp.Jwks {
								_, err = client.DeleteProjectJWKS(ref.ID, jwk.ID)
								if err != nil {
									panic(err)
								}
							}
						},
						Config:  config,
						Destroy: true,
						Check: func(s *terraform.State) error {
							_, ok := s.RootModule().Resources["neon_jwks_url._"]
							assert.False(t, ok, "resource neon_jwks_url._ should be destroyed")
							return nil
						},
					},
				},
			})
	})
}
