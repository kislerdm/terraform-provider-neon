package provider

import (
	"errors"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

func TestAccResourceProjectCreated(t *testing.T) {
	client, err := neon.NewClient()
	if err != nil {
		t.Fatal(err)
	}

	t.Run(
		"shall successfully provision a project", func(t *testing.T) {

			var (
				projectID       string
				defaultBranchID string
			)

			const resourceDefinition = `
resource "neon_project" "this" {
	name      				  = "foo"
	region_id 				  = "aws-us-west-2"
	history_retention_seconds = 30
	pg_version				  = 14
}`

			const resourceName = "neon_project.this"

			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: providerFactories,
					Steps: []resource.TestStep{
						{
							Config: resourceDefinition,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									resourceName,
									"name", "foo",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"region_id", "aws-us-west-2",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"history_retention_seconds", "30",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"pg_version", "14",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"store_password", "true",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"compute_provisioner", "k8s-pod",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"branch.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"quota.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"default_endpoint_settings.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"default_endpoint_settings.0.autoscaling_limit_max_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"default_endpoint_settings.0.autoscaling_limit_min_cu", "0.25",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"default_endpoint_settings.0.suspend_timeout_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"quota.0.active_time_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"quota.0.active_time_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"quota.0.compute_time_seconds", "0",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"quota.0.written_data_bytes", "0",
								),
								resource.TestCheckResourceAttr(
									resourceName,
									"quota.0.data_transfer_bytes", "0",
								),
								resource.TestCheckResourceAttr(
									resourceName,
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
									branches := resp.Branches
									if len(branches) != 1 {
										return errors.New("only a single branch is expected")
									}

									defaultBranchID = branches[0].ID
									if err := resource.TestCheckResourceAttr(
										resourceName, "branch.0.id", defaultBranchID,
									)(state); err != nil {
										return err
									}

									if err := resource.TestCheckResourceAttr(
										resourceName, "default_branch_id", defaultBranchID,
									)(state); err != nil {
										return err
									}

									if err := resource.TestCheckResourceAttr(
										resourceName, "branch.0.name", branches[0].Name,
									)(state); err != nil {
										return err
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
									if len(endpoints) != 1 {
										return errors.New("only a single endpoint is expected")
									}

									if err := resource.TestCheckResourceAttr(
										resourceName, "database_host", endpoints[0].Host,
									)(state); err != nil {
										return err
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
									if len(dbs) != 1 {
										return errors.New("only a single database is expected")
									}

									if err := resource.TestCheckResourceAttr(
										resourceName, "database_user", dbs[0].OwnerName,
									)(state); err != nil {
										return err
									}

									return resource.TestCheckResourceAttr(
										resourceName, "database_name", dbs[0].Name,
									)(state)
								},

								// check the roles
								func(state *terraform.State) error {
									// WHEN
									// list projects
									resp, err := client.ListProjectBranchRoles(projectID, defaultBranchID)
									if err != nil {
										return err
									}

									// THEN
									o := resp.Roles
									if len(o) != 1 {
										return errors.New("only a single role is expected")
									}

									return nil
								},
							),
							ExpectNonEmptyPlan: false,
						},
					},
				},
			)
		},
	)
}
