//go:build acceptance
// +build acceptance

package provider

import (
	"errors"
	"fmt"
	"strconv"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

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

			projectName := strconv.FormatInt(time.Now().UnixMilli(), 10)

			const (
				historyRetentionSeconds = "100"
				autoscalingCUMin        = "0.5"
				autoscalingCUMax        = "2"
				suspendTimeoutSec       = "10"

				quotaActiveTimeSeconds  int = 100
				quotaComputeTimeSeconds int = 100
				quotaWrittenDataBytes   int = 10000
				quotaDataTransferBytes  int = 20000
				quotaLogicalSizeBytes   int = 30000

				branchName     string = "br-foo"
				branchRoleName string = "role-foo"
				dbName         string = "db-foo"
			)

			resourceDefinition := fmt.Sprintf(
				`
resource "neon_project" "this" {
	name      				  = "%s"
	region_id 				  = "aws-us-west-2"
	pg_version				  = 14
	
	history_retention_seconds = %s

	default_endpoint_settings {
    	autoscaling_limit_min_cu = %s
   	 	autoscaling_limit_max_cu = %s
    	suspend_timeout_seconds  = %s
  	}

	quota {
		active_time_seconds  = %d
		compute_time_seconds = %d
		written_data_bytes 	 = %d
		data_transfer_bytes  = %d
		logical_size_bytes 	 = %d
	}

	branch {
		name 	  	  = "%s"
		role_name 	  = "%s"
		database_name = "%s"
	}
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
`,
				projectName,
				historyRetentionSeconds,
				autoscalingCUMin,
				autoscalingCUMax,
				suspendTimeoutSec,
				quotaActiveTimeSeconds,
				quotaComputeTimeSeconds,
				quotaWrittenDataBytes,
				quotaDataTransferBytes,
				quotaLogicalSizeBytes,
				branchName,
				branchRoleName,
				dbName,
			)

			const resourceNameProject = "neon_project.this"

			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: map[string]func() (*schema.Provider, error){
						"neon": func() (*schema.Provider, error) {
							return New("0.2.2"), nil
						},
					},
					Steps: []resource.TestStep{
						{
							ResourceName: "resource",
							Config:       resourceDefinition,
							Check: resource.ComposeTestCheckFunc(
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"name", projectName,
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"region_id", "aws-us-west-2",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"history_retention_seconds", historyRetentionSeconds,
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
									"branch.0.name", branchName,
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"branch.0.role_name", branchRoleName,
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"branch.0.database_name", dbName,
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.0.autoscaling_limit_max_cu",
									autoscalingCUMax,
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.0.autoscaling_limit_min_cu",
									autoscalingCUMin,
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"default_endpoint_settings.0.suspend_timeout_seconds",
									suspendTimeoutSec,
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.#", "1",
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.active_time_seconds", strconv.Itoa(quotaActiveTimeSeconds),
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.compute_time_seconds", strconv.Itoa(quotaComputeTimeSeconds),
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.written_data_bytes", strconv.Itoa(quotaWrittenDataBytes),
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.data_transfer_bytes", strconv.Itoa(quotaDataTransferBytes),
								),
								resource.TestCheckResourceAttr(
									resourceNameProject,
									"quota.0.logical_size_bytes", strconv.Itoa(quotaLogicalSizeBytes),
								),

								// check the project and its settings
								func(state *terraform.State) error {
									// WHEN
									// list projects
									resp, err := client.ListProjects(nil, nil)
									if err != nil {
										return errors.New("listing error: " + err.Error())
									}

									// THEN
									var ref neon.ProjectListItem
									for _, project := range resp.ProjectsResponse.Projects {
										if project.Name == projectName {
											ref = project
										}
										break
									}

									if ref.ID == "" {
										return errors.New("project " + projectName + " shall be created")
									}

									if float64(ref.DefaultEndpointSettings.AutoscalingLimitMinCu) != mustParseFloat64(autoscalingCUMin) {
										return errors.New("AutoscalingLimitMinCu was not set")
									}

									if float64(ref.DefaultEndpointSettings.AutoscalingLimitMaxCu) != mustParseFloat64(autoscalingCUMax) {
										return errors.New("AutoscalingLimitMaxCu was not set")
									}

									v, err := strconv.Atoi(suspendTimeoutSec)
									if err != nil {
										t.Fatal(err)
									}

									if int(ref.DefaultEndpointSettings.SuspendTimeoutSeconds) != v {
										return errors.New("SuspendTimeoutSeconds was not set")
									}

									projectID = ref.ID
									defaultUser = ref.OwnerID

									return nil
								},

								// check data retention
								func(state *terraform.State) error {
									resp, err := client.GetProject(projectID)
									if err != nil {
										return err
									}

									v, err := strconv.Atoi(historyRetentionSeconds)
									if err != nil {
										t.Fatal(err)
									}

									if resp.Project.HistoryRetentionSeconds != int64(v) {
										return errors.New("HistoryRetentionSeconds was not set")
									}

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

											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "branch.0.role_name", db.OwnerName,
											)(state); err != nil {
												return err
											}

											return resource.TestCheckResourceAttr(
												resourceNameProject, "database_name", db.Name,
											)(state)

											return resource.TestCheckResourceAttr(
												resourceNameProject, "branch.0.database_name", db.Name,
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
									"neon_endpoint.this", "autoscaling_limit_max_cu", autoscalingCUMax,
								),
								resource.TestCheckResourceAttr(
									"neon_endpoint.this", "autoscaling_limit_min_cu", autoscalingCUMin,
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

func mustParseFloat64(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return v
}
