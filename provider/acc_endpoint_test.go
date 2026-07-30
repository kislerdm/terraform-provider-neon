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

func TestRecreateEndpointIfNotFound(t *testing.T) {
	// see: https://github.com/kislerdm/terraform-provider-neon/issues/209

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix += "endpointRecreation-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	var newProjectName = func() string {
		return projectNamePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	var preConfig = func(projectName string) {
		ref, err := readProjectInfo(client, projectName)
		if err != nil {
			panic(err)
		}

		resp, err := client.ListProjectEndpoints(ref.ID)
		if err != nil {
			panic(err)
		}
		for _, endpoint := range resp.Endpoints {
			if endpoint.Type == endpointTypeReadOnly {
				resp, err := client.DeleteProjectEndpoint(ref.ID, endpoint.ID)
				if err != nil {
					panic(err)
				}
				waitUnfinishedOperations(context.TODO(), client, resp.OperationsResponse.Operations)
			}
		}
	}

	t.Run("shall indicate non empty plan if the endpoint was deleted outside of terraform", func(t *testing.T) {
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
						Config: fmt.Sprintf(`
							resource "neon_project" "this" { name = "%s" }
							resource "neon_endpoint" "this" {
								project_id = neon_project.this.id
								branch_id  = neon_project.this.default_branch_id
								type       = "%s"
							}
						`, projectName, endpointTypeReadOnly),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_endpoint.this",
								"type", endpointTypeReadOnly,
							),
						),
					},
					{
						PreConfig: func() {
							preConfig(projectName)
						},
						RefreshState:       true,
						ExpectNonEmptyPlan: true,
					},
				},
			})
	})

	t.Run("shall destroy even if the endpoint was deleted outside of terraform", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" { name = "%s" }
resource "neon_endpoint" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	type       = "%s"
}`, projectName, endpointTypeReadOnly)
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
							preConfig(projectName)
						},
						Config:  config,
						Destroy: true,
						Check: func(s *terraform.State) error {
							_, ok := s.RootModule().Resources["neon_endpoint.this"]
							assert.False(t, ok, "resource neon_endpoint.this should be destroyed")
							return nil
						},
					},
				},
			})
	})

	t.Run("shall recreate endpoint upon update if it was deleted outside of terraform", func(t *testing.T) {
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
						Config: fmt.Sprintf(`
							resource "neon_project" "this" { name = "%s" }
							resource "neon_endpoint" "this" {
								project_id = neon_project.this.id
								branch_id  = neon_project.this.default_branch_id
								type       = "%s"
								disabled   = false
							}
						`, projectName, endpointTypeReadOnly),
						Check: resource.TestCheckResourceAttr(
							"neon_endpoint.this",
							"disabled", "false",
						),
					},
					{
						Config: fmt.Sprintf(`
							resource "neon_project" "this" { name = "%s" }
							resource "neon_endpoint" "this" {
								project_id = neon_project.this.id
								branch_id  = neon_project.this.default_branch_id
								type       = "%s"
								disabled   = true
							}
						`, projectName, endpointTypeReadOnly),
						PreConfig: func() {
							preConfig(projectName)
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_endpoint.this",
								"disabled", "true",
							),
							func(_ *terraform.State) error {
								ref, err := readProjectInfo(client, projectName)
								if err != nil {
									return err
								}

								resp, err := client.ListProjectEndpoints(ref.ID)
								if err != nil {
									return err
								}
								assert.Len(t, resp.Endpoints, 2,
									"2 endpoints are expected after recreation")
								for _, endpoint := range resp.Endpoints {
									if endpoint.Type == endpointTypeReadOnly {
										assert.True(t, endpoint.Disabled,
											"endpoint should be disabled after recreation")
									}
								}
								return nil
							},
						),
					},
				},
			})
	})

	t.Run("shall fail to import endpoint if it was deleted", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`
			resource "neon_project" "this" { name = "%s" }
			resource "neon_endpoint" "this" {
				project_id = neon_project.this.id
				branch_id  = neon_project.this.default_branch_id
				type       = "%s"
			}
		`, projectName, endpointTypeReadOnly)
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
						ResourceName: "neon_endpoint.this",
						ImportStateIdFunc: func(s *terraform.State) (string, error) {
							ref, err := readProjectInfo(client, projectName)
							if err != nil {
								return "", err
							}

							resp, err := client.ListProjectEndpoints(ref.ID)
							if err != nil {
								return "", err
							}
							var endpointID string
							for _, endpoint := range resp.Endpoints {
								if endpoint.Type == endpointTypeReadOnly {
									endpointID = endpoint.ID
									resp, err := client.DeleteProjectEndpoint(ref.ID, endpointID)
									if err != nil {
										return "", err
									}
									waitUnfinishedOperations(context.TODO(), client, resp.OperationsResponse.Operations)
								}
							}

							return fmt.Sprintf("%s/%s", ref.ID, endpointID), nil
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
