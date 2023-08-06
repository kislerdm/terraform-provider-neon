//go:build acceptance
// +build acceptance

package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

var providerFactories = map[string]func() (*schema.Provider, error){
	"neon": func() (*schema.Provider, error) {
		return New("0.3.0"), nil
	},
}

func TestProvider(t *testing.T) {
	if err := New("dev").InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestAccEndToEnd(t *testing.T) {
	client, err := neon.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	var (
		projectID       string
		defaultBranchID string
		defaultUser     string
	)

	t.Run(
		"shall successfully provision a project, a branch, an endpoint", func(t *testing.T) {

			const resourceDefinition = `
resource "neon_project" "this" {
	name      				  = "foo"
	region_id 				  = "aws-us-west-2"
	history_retention_seconds = 30
	pg_version				  = 14
}

resource "neon_branch" "this" {
	name 	   = "bar"
	project_id = neon_project.this.id
	parent_id  = neon_project.this.default_branch_id
}

resource "neon_endpoint" "this" {
	project_id 				= neon_project.this.id
	branch_id  				= neon_branch.this.id
}

resource "neon_role" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	name 	   = "qux"
}

resource "neon_database" "this" {
	project_id = neon_project.this.id
	branch_id  = neon_project.this.default_branch_id
	name 	   = "quxx"
	owner_name = neon_role.this.name
}
`

			const resourceNameProject = "neon_project.this"

			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						{
							ResourceName: "resource",
							Config:       resourceDefinition,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"name", "foo",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"region_id", "aws-us-west-2",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"history_retention_seconds", "30",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"pg_version", "14",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"store_password", "true",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"compute_provisioner", "k8s-pod",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"branch.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.0.autoscaling_limit_max_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.0.autoscaling_limit_min_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.0.suspend_timeout_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.active_time_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.active_time_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.compute_time_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.written_data_bytes", "0",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.data_transfer_bytes", "0",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.logical_size_bytes", "0",
								),

								// check the number of created projects
								func(state *terraform.State) error {
									// WHEN
									// list projects
									resp, err := client.ListProjects(nil, nil)
									if err != nil {
										return err
									}

									// THEN
									projects := resp.ProjectsResponse.Projects
									if len(projects) != 1 {
										return errors.New("only a single project is expected")
									}

									projectID = projects[0].ID
									defaultUser = projects[0].OwnerID

									return nil
								},

								// check the branches
								func(state *terraform.State) error {
									// WHEN
									// list projects
									resp, err := client.ListProjectBranches(projectID)
									if err != nil {
										return err
									}

									// THEN
									if len(resp.Branches) != 2 {
										return errors.New("only two branches are expected")
									}

									for _, branch := range resp.Branches {
										if branch.Primary {
											defaultBranchID = branch.ID
											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "branch.0.id", defaultBranchID,
											)(state); err != nil {
												return err
											}

											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "default_branch_id", defaultBranchID,
											)(state); err != nil {
												return err
											}

											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "branch.0.name", branch.Name,
											)(state); err != nil {
												return err
											}
											break
										}
									}

									return nil
								},

								// check the endpoints
								func(state *terraform.State) error {
									// WHEN
									// list projects
									resp, err := client.ListProjectEndpoints(projectID)
									if err != nil {
										return err
									}

									// THEN
									endpoints := resp.Endpoints
									if len(endpoints) != 2 {
										return errors.New("only two endpoints are expected")
									}

									for _, endpoint := range endpoints {
										if endpoint.BranchID == defaultBranchID {
											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "database_host", endpoint.Host,
											)(state); err != nil {
												return err
											}
										}
									}

									return nil
								},

								// check the databases
								func(state *terraform.State) error {
									// WHEN
									// list projects
									resp, err := client.ListProjectBranchDatabases(projectID, defaultBranchID)
									if err != nil {
										return err
									}

									// THEN
									dbs := resp.Databases
									if len(dbs) != 2 {
										return errors.New("only two databases is expected")
									}

									for _, db := range dbs {
										if db.OwnerName == defaultUser {
											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "database_user", db.OwnerName,
											)(state); err != nil {
												return err
											}

											return resource.TestCheckResourceAttr(
												resourceNameProject, "database_name", db.Name,
											)(state)
										}
									}

									return nil
								},
							),
						},
						{
							ResourceName: "branch",
							RefreshState: true,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_branch.this", "name", "bar",
								),
							),
						},
						{
							ResourceName: "endpoint",
							RefreshState: true,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "autoscaling_limit_max_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "autoscaling_limit_min_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "disabled", "false",
								),
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "suspend_timeout_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "compute_provisioner", "k8s-pod",
								),
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "type", "read_write",
								),
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "region_id", "aws-us-west-2",
								),

								func(state *terraform.State) error {
									// WHEN
									resp, err := client.ListProjectEndpoints(projectID)
									if err != nil {
										return err
									}

									// THEN
									for _, endpoint := range resp.Endpoints {
										if endpoint.BranchID != defaultBranchID {
											if err := resource.TestCheckResourceAttr(
												"neon_endpoint.this", "host", endpoint.Host,
											)(state); err != nil {
												return err
											}
										}
									}

									return nil
								},
							),
						},
						{
							ResourceName: "role",
							RefreshState: true,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr("neon_role.this", "name", "qux"),
								resource.TestCheckResourceAttr("neon_role.this", "protected", "false"),
							),
						},
						{
							ResourceName: "database",
							RefreshState: true,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr("neon_database.this", "name", "quxx"),
							),
						},
					},
				},
			)
		},
	)
}
