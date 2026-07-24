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

func TestRecreateBranchIfNotFound(t *testing.T) {
	// see: https://github.com/kislerdm/terraform-provider-neon/issues/209

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix += "branchRecreation-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	var newProjectName = func() string {
		return projectNamePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	t.Run("shall update the state if the branch was deleted outside of terraform", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_branch" "this" {
	project_id = neon_project.this.id 
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
								"neon_branch.this",
								"name", "test",
							),
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									return err
								}
								for _, branch := range resp.Branches {
									if branch.Name == "test" {
										_, err := client.DeleteProjectBranch(ref.ID, branch.ID)
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
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									return err
								}
								assert.Len(t, resp.Branches, 1,
									"1 branch is expected after deletion")
								return nil
							}),
					},
				},
			})
	})

	t.Run("shall destroy even if the branch was deleted outside of terraform,", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_branch" "this" {
	project_id = neon_project.this.id 
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
								"neon_branch.this",
								"name", "test",
							),
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									return err
								}
								for _, branch := range resp.Branches {
									if branch.Name == "test" {
										_, err := client.DeleteProjectBranch(ref.ID, branch.ID)
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

	t.Run("shall recreate branch upon read if it was deleted outside of terraform", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_branch" "this" {
	project_id = neon_project.this.id 
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
								"neon_branch.this",
								"name", "test",
							),
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									return err
								}
								for _, branch := range resp.Branches {
									if branch.Name == "test" {
										_, err := client.DeleteProjectBranch(ref.ID, branch.ID)
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
								"neon_branch.this",
								"name", "test",
							),
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									return err
								}
								assert.Len(t, resp.Branches, 2,
									"2 branches are expected after recreation")
								return nil
							},
						),
					},
				},
			})
	})

	t.Run("shall recreate branch upon update if it was deleted outside of terraform", func(t *testing.T) {
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
resource "neon_branch" "this" {
	project_id = neon_project.this.id 
	name       = "foo"
}`, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_branch.this",
								"name", "foo",
							),
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									return err
								}
								for _, branch := range resp.Branches {
									if branch.Name == "foo" {
										_, err := client.DeleteProjectBranch(ref.ID, branch.ID)
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
resource "neon_branch" "this" {
	project_id = neon_project.this.id 
	name       = "bar"
}`, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_branch.this",
								"name", "bar",
							),
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectBranches(ref.ID,
									nil, nil, nil, nil, nil)
								if err != nil {
									return err
								}
								assert.Len(t, resp.Branches, 2,
									"2 branches are expected after recreation")
								var found bool
								for _, branch := range resp.Branches {
									if branch.Name == "bar" {
										found = true
									}
								}
								assert.Truef(t, found, "branch 'bar' is expected to be found after recreation")
								return nil
							},
						),
					},
				},
			})
	})

	t.Run("shall fail to import branch if it was deleted", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_branch" "this" {
	project_id = neon_project.this.id 
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
						ResourceName: "neon_branch.this",
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							ref, err := readProjectInfo(client, projectName)
							if err != nil {
								return "", err
							}

							resp, err := client.ListProjectBranches(ref.ID,
								nil, nil, nil, nil, nil)
							if err != nil {
								return "", err
							}
							var branchID string
							for _, branch := range resp.Branches {
								if branch.Name == "test" {
									branchID = branch.ID
									_, err := client.DeleteProjectBranch(ref.ID, branch.ID)
									if err != nil {
										return "", err
									}
								}
							}
							return fmt.Sprintf("%s/%s", ref.ID, branchID), nil
						},
						ExpectError: regexp.MustCompile("404"),
					},
				},
			})
	})
}
