package provider

import (
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

func TestRecreateDatabaseIfNotFound(t *testing.T) {
	// see: https://github.com/kislerdm/terraform-provider-neon/issues/209

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix += "databaseRecreation-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	var newProjectName = func() string {
		return projectNamePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	t.Run("shall yield non empty refresh plan if the database was deleted outside of terraform",
		func(t *testing.T) {
			projectName := newProjectName()
			config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_database" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	owner_name = neon_project.this.database_user
	name       = "test"
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
							Config: config,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_database.this",
									"name", "test",
								),
								func(_ *terraform.State) error {
									ref, err := readProjectInfo(client, projectName)
									if err != nil {
										return err
									}
									br, err := client.ListProjectBranches(ref.ID,
										nil, nil, nil, nil, nil)
									if err != nil {
										return err
									}

									var branchID string
									for _, branch := range br.Branches {
										if branch.Default {
											branchID = branch.ID
										}
									}

									resp, err := client.ListProjectBranchDatabases(ref.ID, branchID)
									if err != nil {
										return err
									}
									assert.Len(t, resp.Databases, 2)

									return nil
								},
							),
						},
						{
							PreConfig: func() {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									panic(err)
								}
								br, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									panic(err)
								}
								for _, branch := range br.Branches {
									if branch.Default {
										_, err = client.DeleteProjectBranchDatabase(ref.ID, branch.ID, "test")
										if err != nil {
											panic(err)
										}
									}
								}
							},
							RefreshState:       true,
							ExpectNonEmptyPlan: true,
						},
					},
				})
		})

	t.Run("shall destroy even if the database was deleted outside of terraform,", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_database" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	owner_name = neon_project.this.database_user
	name       = "test"
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
						Config: config,
					},
					{
						Config: config,
						PreConfig: func() {
							ref, err := readProjectInfo(client, projectName)
							if err != nil {
								panic(err)
							}
							br, err := client.ListProjectBranches(ref.ID,
								nil, nil, nil, nil, nil)
							if err != nil {
								panic(err)
							}

							for _, branch := range br.Branches {
								if branch.Default {
									_, err := client.DeleteProjectBranchDatabase(ref.ID, branch.ID,
										"test")
									if err != nil {
										panic(err)
									}
								}
							}
						},
						Destroy: true,
					},
				},
			})
	})

	t.Run("shall fail to import database if it was deleted", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_database" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	owner_name = neon_project.this.database_user
	name       = "test"
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
						Config: config,
					},
					{
						Config:       config,
						ImportState:  true,
						ResourceName: "neon_database.this",
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							ref, err := readProjectInfo(client, projectName)
							if err != nil {
								return "", err
							}

							br, err := client.ListProjectBranches(ref.ID,
								nil, nil, nil, nil, nil)
							if err != nil {
								return "", err
							}

							var branchID string
							for _, branch := range br.Branches {
								if branch.Default {
									branchID = branch.ID
								}
							}

							resp, err := client.ListProjectBranchDatabases(ref.ID, branchID)
							if err != nil {
								return "", err
							}

							for _, db := range resp.Databases {
								if db.Name == "test" {
									_, err := client.DeleteProjectBranchDatabase(ref.ID, branchID, db.Name)
									if err != nil {
										return "", err
									}
								}
							}
							return fmt.Sprintf("%s/%s/%s", ref.ID, branchID, "test"), nil
						},
						ExpectError: regexp.MustCompile("404"),
					},
					{
						// recreate the database to avoid errors on post-test destroy error
						Config: config,
					},
				},
			})
	})
}
