package provider

import (
	"context"
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

	var preConfig = func(projectName string, branchName string) {
		ref, err := readProjectInfo(client, projectName)
		if err != nil {
			panic(err)
		}

		resp, err := client.ListProjectBranches(ref.ID,
			nil, nil, nil, nil, nil)
		if err != nil {
			panic(err)
		}
		for _, branch := range resp.Branches {
			if branch.Name == branchName {
				resp, err := client.DeleteProjectBranch(ref.ID, branch.ID)
				if err != nil {
					panic(err)
				}
				waitUnfinishedOperations(context.TODO(), client, resp.OperationsResponse.Operations)
			}
		}
	}

	t.Run("shall indicate non empty plan if the branch was deleted outside of terraform", func(t *testing.T) {
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
	name       = "test"
}`, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_branch.this",
								"name", "test",
							),
						),
					},
					{
						PreConfig: func() {
							preConfig(projectName, "test")
						},
						RefreshState:       true,
						ExpectNonEmptyPlan: true,
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
						),
					},
					{
						PreConfig: func() {
							preConfig(projectName, "test")
						},
						Config:  config,
						Destroy: true,
						Check: func(s *terraform.State) error {
							_, ok := s.RootModule().Resources["neon_branch.this"]
							assert.False(t, ok, "resource neon_branch.this should be destroyed")
							return nil
						},
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
						),
					},
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}
resource "neon_branch" "this" {
	project_id = neon_project.this.id 
	name       = "bar"
}`, projectName),
						PreConfig: func() {
							preConfig(projectName, "foo")
						},
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
								var oldFound bool
								for _, branch := range resp.Branches {
									if branch.Name == "bar" {
										found = true
									}
									if branch.Name == "foo" {
										oldFound = true
									}
								}
								assert.Truef(t, found, "branch 'bar' is expected to be found after recreation")
								assert.Falsef(t, oldFound, "branch 'foo' is not expected to be found")
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
					// to avoid dangling resources on post-test destroy
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" { name = "%s" }`, projectName),
					},
				},
			})
	})
}
