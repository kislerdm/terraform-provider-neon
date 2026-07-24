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

	t.Run("shall update the state if the database was deleted outside of terraform", func(t *testing.T) {
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

								for _, db := range resp.Databases {
									if db.Name == "test" {
										_, err := client.DeleteProjectBranchDatabase(ref.ID, branchID, db.Name)
										if err != nil {
											return err
										}
									}
								}

								return nil
							},
						),
					},
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}`, projectName),
						Check: resource.ComposeTestCheckFunc(
							func(s *terraform.State) error {
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
								assert.Len(t, resp.Databases, 1)

								return nil
							}),
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

								for _, db := range resp.Databases {
									if db.Name == "test" {
										_, err := client.DeleteProjectBranchDatabase(ref.ID, branchID, db.Name)
										if err != nil {
											return err
										}
									}
								}

								return nil
							},
						),
					},
				},
			})
	})

	t.Run("shall recreate database upon read if it was deleted outside of terraform", func(t *testing.T) {
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

								for _, db := range resp.Databases {
									if db.Name == "test" {
										_, err := client.DeleteProjectBranchDatabase(ref.ID, branchID, db.Name)
										if err != nil {
											return err
										}
									}
								}

								return nil
							},
						),
					},
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
				},
			})
	})

	t.Run("shall recreate database upon update if it was deleted outside of terraform", func(t *testing.T) {
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
resource "neon_database" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	owner_name = neon_project.this.database_user
	name       = "foo"
}`, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_database.this",
								"name", "foo",
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

								for _, db := range resp.Databases {
									if db.Name == "foo" {
										_, err := client.DeleteProjectBranchDatabase(ref.ID, branchID, db.Name)
										if err != nil {
											return err
										}
									}
								}

								return nil
							},
						),
					},
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_database" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	owner_name = neon_project.this.database_user
	name       = "bar"
}`, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_database.this",
								"name", "bar",
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

								var found bool
								for _, db := range resp.Databases {
									if db.Name == "bar" {
										found = true
									}
								}
								assert.Truef(t, found, "database 'bar' is expected to be found after recreation")
								return nil
							},
						),
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
