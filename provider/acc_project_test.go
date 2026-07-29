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

func TestRecreateProjectIfNotFound(t *testing.T) {
	// see: https://github.com/kislerdm/terraform-provider-neon/issues/209

	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix += "projectRecreation-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	var newProjectName = func() string {
		return projectNamePrefix + strconv.FormatInt(time.Now().UnixMilli(), 10)
	}

	var preConfig = func(projectName string) string {
		ref, err := readProjectInfo(client, projectName)
		if err != nil {
			panic(err)
		}
		_, err = client.DeleteProject(ref.ID)
		if err != nil {
			panic(err)
		}
		return ref.ID
	}

	t.Run("shall indicate non empty plan if the project was deleted outside of terraform", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}`, projectName)
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
								"neon_project.this",
								"name", projectName,
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

	t.Run("shall destroy even if the project was deleted outside of terraform,", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}`, projectName)
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
								"neon_project.this",
								"name", projectName,
							),
						),
					},
					{
						PreConfig: func() {
							preConfig(projectName)
						},
						Config:  config,
						Destroy: true,
						Check: func(s *terraform.State) error {
							_, ok := s.RootModule().Resources["neon_project.this"]
							assert.False(t, ok, "resource neon_project.this should be destroyed")
							return nil
						},
					},
				},
			})
	})

	t.Run("shall recreate project upon update if it was deleted outside of terraform", func(t *testing.T) {
		projectName := newProjectName()
		var refProjectID string
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s-foo"}`, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_project.this",
								"name", fmt.Sprintf("%s-foo", projectName),
							),
						),
					},
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s-bar"}`, projectName),
						PreConfig: func() {
							refProjectID = preConfig(fmt.Sprintf("%s-foo", projectName))
						},
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_project.this",
								"name", fmt.Sprintf("%s-bar", projectName),
							),
							func(_ *terraform.State) error {
								got, err := readProjectInfo(client, fmt.Sprintf("%s-bar", projectName))
								if err != nil {
									return err
								}
								assert.NotEqualf(t, refProjectID, got.ID,
									"project ID should be different after recreation")
								return nil
							},
						),
					},
				},
			})
	})

	t.Run("shall fail to import project if it was deleted", func(t *testing.T) {
		projectName := newProjectName()
		config := fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}`, projectName)
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
							preConfig(projectName)
						},
						ImportState:  true,
						ResourceName: "neon_project.this",
						ExpectError:  regexp.MustCompile("404"),
					},
				},
			})
	})
}
