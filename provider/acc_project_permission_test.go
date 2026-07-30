package provider

import (
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestProjectPermissionFSMIfResourceDeletedOutsideTerraform(t *testing.T) {
	// see: https://github.com/kislerdm/terraform-provider-neon/issues/209

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix += "projectPermissionRecreation-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	var newProjectName = func() string {
		return projectNamePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	var preConfig = func(projectName, email string) {
		ref, err := readProjectInfo(client, projectName)
		if err != nil {
			panic(err)
		}

		resp, err := client.ListProjectPermissions(ref.ID)
		if err != nil {
			panic(err)
		}
		for _, permission := range resp.ProjectPermissions {
			if permission.GrantedToEmail == email {
				_, err = client.RevokePermissionFromProject(ref.ID, permission.ID)
				if err != nil {
					panic(err)
				}
			}
		}
	}

	t.Run("shall indicate non empty plan if the project permission was deleted outside of terraform",
		func(t *testing.T) {
			email := "foo@bar.baz"
			projectName := newProjectName()
			resource.Test(
				t, resource.TestCase{
					ProviderFactories: map[string]func() (*schema.Provider, error){
						"neon": func() (*schema.Provider, error) {
							return newAccTest(), nil
						},
					},
					Steps: []resource.TestStep{
						{
							Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_project_permission" "this" {
	project_id = neon_project.this.id 
	grantee    = "%s"
}`, projectName, email),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_project_permission.this",
									"grantee", email,
								),
							),
						},
						{
							PreConfig: func() {
								preConfig(projectName, email)
							},
							RefreshState:       true,
							ExpectNonEmptyPlan: true,
						},
					},
				})
		})

	t.Run("shall destroy even if the project permission was deleted outside of terraform,", func(t *testing.T) {
		email := "foo@bar.baz"
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_project_permission" "this" {
	project_id = neon_project.this.id 
	grantee    = "%s"
}`, projectName, email)
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
							resource.TestCheckResourceAttr(
								"neon_project_permission.this",
								"grantee", email,
							),
						),
					},
					{
						PreConfig: func() {
							preConfig(projectName, email)
						},
						Config:  config,
						Destroy: true,
						Check: func(s *terraform.State) error {
							_, ok := s.RootModule().Resources["neon_project_permission.this"]
							assert.False(t, ok, "resource neon_project_permission.this should be destroyed")
							return nil
						},
					},
					// to avoid dangling resources on post-test destroy
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" { name = "%s" }`, projectName),
					},
				},
			})
	})
}
