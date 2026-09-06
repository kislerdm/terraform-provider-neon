package provider

import (
	"fmt"
	"os"
	"regexp"
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

	projectNamePrefix := "orgTest-"
	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, &orgID, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	projectName := newProjectName(projectNamePrefix)

	resourceDefinition := fmt.Sprintf(`resource "neon_project" "this" {
    org_id              = "%s"
	name                = "%s"
}`, orgID, projectName)

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
					Check: func(state *terraform.State) error {
						var (
							e    error
							resp neon.ListProjectsRespObj
						)
						resp, e = client.ListProjects(nil, nil, &projectName, &orgID, nil, nil)
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

// issue 203: https://github.com/kislerdm/terraform-provider-neon/issues/203
func TestAccHIPAA(t *testing.T) {
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

	projectNamePrefix := "hipaa-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, &orgID, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	var newResourceDefinition = func(projectName string, withHipaa string) string {
		return fmt.Sprintf(`resource "neon_project" "this" {
    org_id = "%s"
	name   = "%s"
	hipaa  = "%s"
}`, orgID, projectName, withHipaa)
	}

	t.Run("shall fail to disable initially enabled HIPAA", func(t *testing.T) {
		projectName := newProjectName(projectNamePrefix)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: newResourceDefinition(projectName, "yes"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_project.this",
								"hipaa", "yes",
							),
							func(state *terraform.State) error {
								resp, err := client.ListProjects(nil, nil, &projectName, &orgID, nil, nil)
								if err != nil {
									return err
								}

								if len(resp.Projects) != 1 {
									return fmt.Errorf(
										"project %s should have been creted in the org %s", projectName, orgID,
									)
								}

								settings := resp.Projects[0].Settings
								if settings == nil || settings.Hipaa == nil || !*settings.Hipaa {
									return fmt.Errorf("project %s does not have hipaa enabled", projectName)
								}

								return nil
							},
						),
					},
					{
						Config:      newResourceDefinition(projectName, "no"),
						ExpectError: regexp.MustCompile("disabling HIPAA is not allowed"),
					},
				},
			},
		)
	})

	t.Run("shall enable initially implicitly disabled HIPAA", func(t *testing.T) {
		projectName := newProjectName(projectNamePrefix)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {
    org_id = "%s"
	name   = "%s"
}`, orgID, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckNoResourceAttr("neon_project.this", "hipaa"),
							func(state *terraform.State) error {
								resp, err := client.ListProjects(nil, nil, &projectName, &orgID, nil, nil)
								if err != nil {
									return err
								}

								if len(resp.Projects) != 1 {
									return fmt.Errorf(
										"project %s should have been creted in the org %s", projectName, orgID,
									)
								}

								settings := resp.Projects[0].Settings
								if settings == nil || settings.Hipaa == nil || *settings.Hipaa {
									return fmt.Errorf("project %s does have hipaa enabled", projectName)
								}

								return nil
							},
						),
					},
					{
						Config: newResourceDefinition(projectName, "yes"),
						Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(
							"neon_project.this",
							"hipaa", "yes",
						),
							func(state *terraform.State) error {
								resp, err := client.ListProjects(nil, nil, &projectName, &orgID, nil, nil)
								if err != nil {
									return err
								}

								if len(resp.Projects) != 1 {
									return fmt.Errorf(
										"project %s should have been creted in the org %s", projectName, orgID,
									)
								}

								settings := resp.Projects[0].Settings
								if settings == nil || settings.Hipaa == nil || !*settings.Hipaa {
									return fmt.Errorf("project %s does not have hipaa enabled", projectName)
								}

								return nil
							}),
					},
				},
			},
		)
	})

	t.Run("shall enable initially explicitly disabled HIPAA", func(t *testing.T) {
		projectName := newProjectName(projectNamePrefix)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: newResourceDefinition(projectName, "no"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_project.this",
								"hipaa", "no",
							),
							func(state *terraform.State) error {
								resp, err := client.ListProjects(nil, nil, &projectName, &orgID, nil, nil)
								if err != nil {
									return err
								}

								if len(resp.Projects) != 1 {
									return fmt.Errorf(
										"project %s should have been creted in the org %s", projectName, orgID,
									)
								}

								settings := resp.Projects[0].Settings
								if settings == nil || settings.Hipaa == nil || *settings.Hipaa {
									return fmt.Errorf("project %s does have hipaa enabled", projectName)
								}

								return nil
							},
						),
					},
					{
						Config: newResourceDefinition(projectName, "yes"),
						Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(
							"neon_project.this",
							"hipaa", "yes",
						),
							func(state *terraform.State) error {
								resp, err := client.ListProjects(nil, nil, &projectName, &orgID, nil, nil)
								if err != nil {
									return err
								}

								if len(resp.Projects) != 1 {
									return fmt.Errorf(
										"project %s should have been creted in the org %s", projectName, orgID,
									)
								}

								settings := resp.Projects[0].Settings
								if settings == nil || settings.Hipaa == nil || !*settings.Hipaa {
									return fmt.Errorf("project %s does not have hipaa enabled", projectName)
								}

								return nil
							}),
					},
				},
			},
		)
	})

	t.Run("shall import project with enabled HIPAA", func(t *testing.T) {
		projectName := newProjectName(projectNamePrefix)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: newResourceDefinition(projectName, "yes"),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								"neon_project.this",
								"hipaa", "yes",
							),
						),
					},
					{
						ImportState:  true,
						Config:       newResourceDefinition(projectName, "yes"),
						ResourceName: "neon_project.this",
						Check: resource.ComposeTestCheckFunc(resource.TestCheckResourceAttr(
							"neon_project.this",
							"hipaa", "yes",
						)),
					},
				},
			},
		)
	})

	t.Run("shall have empty plan after for implicitly disabled HIPAA if no config changed", func(t *testing.T) {
		projectName := newProjectName(projectNamePrefix)
		resource.Test(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return newAccTest(), nil
					},
				},
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {
    org_id = "%s"
	name   = "%s"
}`, orgID, projectName),
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckNoResourceAttr("neon_project.this", "hipaa"),
						),
					},
					{
						Config: fmt.Sprintf(`resource "neon_project" "this" {
    org_id = "%s"
	name   = "%s"
}`, orgID, projectName),
						PlanOnly:           true,
						ExpectNonEmptyPlan: false,
					},
				},
			},
		)
	})
}
