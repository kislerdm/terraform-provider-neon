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

	projectNamePrefix := "projectRecreation-"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

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
		projectName := newProjectName(projectNamePrefix)
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
		projectName := newProjectName(projectNamePrefix)
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
		projectName := newProjectName(projectNamePrefix)
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
		projectName := newProjectName(projectNamePrefix)
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

func TestPrimaryCompute(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix := "primaryCompute"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	t.Run("shall create a new project with different compute configs and the default compute's configs",
		func(t *testing.T) {
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
		name = "%s"
		autoscaling_limit_min_cu = 0.25
		autoscaling_limit_max_cu = 1
		suspend_timeout_seconds  = 300
		
		primary_compute {
			autoscaling_limit_min_cu = 0.5
			autoscaling_limit_max_cu = 2
			suspend_timeout_seconds  = -1
		}
}
`, projectName),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"name", projectName,
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_min_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_max_cu", "1",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"suspend_timeout_seconds", "300",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_min_cu", "0.5",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_max_cu", "2",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.suspend_timeout_seconds", "-1",
								),
								func(_ *terraform.State) error {
									got, err := readProjectInfo(client, projectName)
									if err != nil {
										return err
									}
									assert.Equalf(t, int64(300),
										int64(*got.DefaultEndpointSettings.SuspendTimeoutSeconds),
										"expected compute defaults: suspend_timeout_seconds")
									assert.Equalf(t, float64(0.25),
										float64(*got.DefaultEndpointSettings.AutoscalingLimitMinCu),
										"expected compute defaults: autoscaling_limit_min_cu")
									assert.Equalf(t, float64(1),
										float64(*got.DefaultEndpointSettings.AutoscalingLimitMaxCu),
										"expected compute defaults: autoscaling_limit_max_cu")

									e, err := client.ListProjectEndpoints(got.ID)
									if err != nil {
										return err
									}
									defaultEndpoint := e.Endpoints[0]
									assert.Equalf(t, int64(-1),
										int64(defaultEndpoint.SuspendTimeoutSeconds),
										"expected default compute's suspend_timeout_seconds")
									assert.Equalf(t, float64(0.5),
										float64(defaultEndpoint.AutoscalingLimitMinCu),
										"expected default compute's autoscaling_limit_min_cu")
									assert.Equalf(t, float64(2),
										float64(defaultEndpoint.AutoscalingLimitMaxCu),
										"expected default compute's autoscaling_limit_max_cu")

									return nil
								},
							),
						},
					},
				})
		})

	t.Run("shall update default compute's configs w/o affecting project's compute configs",
		func(t *testing.T) {
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
		name = "%s"
		autoscaling_limit_min_cu = 0.25
		autoscaling_limit_max_cu = 1
		suspend_timeout_seconds  = 300
		
		primary_compute {
			autoscaling_limit_min_cu = 0.5
			autoscaling_limit_max_cu = 2
			suspend_timeout_seconds  = -1
		}
}
`, projectName),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"name", projectName,
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_min_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_max_cu", "1",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"suspend_timeout_seconds", "300",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_min_cu", "0.5",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_max_cu", "2",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.suspend_timeout_seconds", "-1",
								),
							),
						},
						{
							Config: fmt.Sprintf(`resource "neon_project" "this" {
		name = "%s"
		autoscaling_limit_min_cu = 0.25
		autoscaling_limit_max_cu = 1
		suspend_timeout_seconds  = 300
		
		primary_compute {
			autoscaling_limit_min_cu = 1
			autoscaling_limit_max_cu = 4
			suspend_timeout_seconds  = 1200
		}
}
`, projectName),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"name", projectName,
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_min_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_max_cu", "1",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"suspend_timeout_seconds", "300",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_min_cu", "1",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_max_cu", "4",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.suspend_timeout_seconds", "1200",
								),
								func(_ *terraform.State) error {
									got, err := readProjectInfo(client, projectName)
									if err != nil {
										return err
									}
									assert.Equalf(t, int64(300),
										int64(*got.DefaultEndpointSettings.SuspendTimeoutSeconds),
										"expected compute defaults: suspend_timeout_seconds")
									assert.Equalf(t, float64(0.25),
										float64(*got.DefaultEndpointSettings.AutoscalingLimitMinCu),
										"expected compute defaults: autoscaling_limit_min_cu")
									assert.Equalf(t, float64(1),
										float64(*got.DefaultEndpointSettings.AutoscalingLimitMaxCu),
										"expected compute defaults: autoscaling_limit_max_cu")

									e, err := client.ListProjectEndpoints(got.ID)
									if err != nil {
										return err
									}
									defaultEndpoint := e.Endpoints[0]
									assert.Equalf(t, int64(1200),
										int64(defaultEndpoint.SuspendTimeoutSeconds),
										"expected default compute's suspend_timeout_seconds")
									assert.Equalf(t, float64(1),
										float64(defaultEndpoint.AutoscalingLimitMinCu),
										"expected default compute's autoscaling_limit_min_cu")
									assert.Equalf(t, float64(4),
										float64(defaultEndpoint.AutoscalingLimitMaxCu),
										"expected default compute's autoscaling_limit_max_cu")

									return nil
								},
							),
						},
					},
				})
		})

	t.Run("shall update project's compute configs w/o affecting default compute's configs",
		func(t *testing.T) {
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
		name = "%s"
		autoscaling_limit_min_cu = 0.25
		autoscaling_limit_max_cu = 1
		suspend_timeout_seconds  = 300
		
		primary_compute {
			autoscaling_limit_min_cu = 0.5
			autoscaling_limit_max_cu = 2
			suspend_timeout_seconds  = -1
		}
}
`, projectName),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"name", projectName,
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_min_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_max_cu", "1",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"suspend_timeout_seconds", "300",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_min_cu", "0.5",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_max_cu", "2",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.suspend_timeout_seconds", "-1",
								),
							),
						},
						{
							Config: fmt.Sprintf(`resource "neon_project" "this" {
		name = "%s"
		autoscaling_limit_min_cu = 0.5
		autoscaling_limit_max_cu = 2
		suspend_timeout_seconds  = 600
		
		primary_compute {
			autoscaling_limit_min_cu = 0.5
			autoscaling_limit_max_cu = 2
			suspend_timeout_seconds  = -1
		}
}
`, projectName),
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"name", projectName,
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_min_cu", "0.5",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"autoscaling_limit_max_cu", "2",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"suspend_timeout_seconds", "600",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_min_cu", "0.5",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.autoscaling_limit_max_cu", "2",
								),
								resource.TestCheckResourceAttr(
									"neon_project.this",
									"primary_compute.0.suspend_timeout_seconds", "-1",
								),
								func(_ *terraform.State) error {
									got, err := readProjectInfo(client, projectName)
									if err != nil {
										return err
									}
									assert.Equalf(t, int64(600),
										int64(*got.DefaultEndpointSettings.SuspendTimeoutSeconds),
										"expected compute defaults: suspend_timeout_seconds")
									assert.Equalf(t, float64(0.5),
										float64(*got.DefaultEndpointSettings.AutoscalingLimitMinCu),
										"expected compute defaults: autoscaling_limit_min_cu")
									assert.Equalf(t, float64(2),
										float64(*got.DefaultEndpointSettings.AutoscalingLimitMaxCu),
										"expected compute defaults: autoscaling_limit_max_cu")

									e, err := client.ListProjectEndpoints(got.ID)
									if err != nil {
										return err
									}
									defaultEndpoint := e.Endpoints[0]
									assert.Equalf(t, int64(-1),
										int64(defaultEndpoint.SuspendTimeoutSeconds),
										"expected default compute's suspend_timeout_seconds")
									assert.Equalf(t, float64(0.5),
										float64(defaultEndpoint.AutoscalingLimitMinCu),
										"expected default compute's autoscaling_limit_min_cu")
									assert.Equalf(t, float64(2),
										float64(defaultEndpoint.AutoscalingLimitMaxCu),
										"expected default compute's autoscaling_limit_max_cu")

									return nil
								},
							),
						},
					},
				})
		})
}
