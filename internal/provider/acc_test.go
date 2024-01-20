package provider

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
)

func TestAcc(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	end2end(t, client)

	projectAllowedIPs(t, client)

	projectLogicalReplication(t, client)
}

func end2end(t *testing.T, client *neon.Client) {
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
				autoscalingCUMin        = "0.25"
				autoscalingCUMax        = "0.5"
				suspendTimeoutSec       = "10"

				quotaActiveTimeSeconds  int = 100
				quotaComputeTimeSeconds int = 100
				quotaWrittenDataBytes   int = 10000
				quotaDataTransferBytes  int = 20000
				quotaLogicalSizeBytes   int = 30000

				branchName     string = "br-foo"
				branchRoleName string = "role-foo"
				dbName         string = "db-foo"
				customRoleName string = "qux"
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
	name 	   = "%s"
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
				customRoleName,
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

									if float64(*ref.DefaultEndpointSettings.AutoscalingLimitMinCu) != mustParseFloat64(autoscalingCUMin) {
										return errors.New("AutoscalingLimitMinCu was not set")
									}

									if float64(*ref.DefaultEndpointSettings.AutoscalingLimitMaxCu) != mustParseFloat64(autoscalingCUMax) {
										return errors.New("AutoscalingLimitMaxCu was not set")
									}

									v, err := strconv.Atoi(suspendTimeoutSec)
									if err != nil {
										t.Fatal(err)
									}

									if int(*ref.DefaultEndpointSettings.SuspendTimeoutSeconds) != v {
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

											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "database_name", db.Name,
											)(state); err != nil {
												return err
											}

											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "branch.0.database_name", db.Name,
											)(state); err != nil {
												return err
											}
										}
									}

									return nil
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
									roles := resp.Roles
									if len(roles) != 2 {
										return errors.New("two roles are expected for the branch " + defaultBranchID)
									}

									for _, role := range roles {
										// validate password
										resp, err := client.GetProjectBranchRolePassword(projectID,
											defaultBranchID, role.Name)
										if err != nil {
											return err
										}

										switch role.Name {
										case customRoleName:
											if err := resource.TestCheckResourceAttr(
												"neon_role.this", "password", resp.Password,
											)(state); err != nil {
												return err
											}
										default:
											if err := resource.TestCheckResourceAttr(
												resourceNameProject, "database_password", resp.Password,
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

func projectAllowedIPs(t *testing.T, client *neon.Client) {
	wantAllowedIPs := []string{"192.168.1.0", "192.168.2.0/24"}
	ips := `["` + strings.Join(wantAllowedIPs, `", "`) + `"]`

	t.Run("shall provision a project with a custom list of allowed IPs", func(t *testing.T) {
		projectName := strconv.FormatInt(time.Now().UnixMilli(), 10)

		resourceDefinition := fmt.Sprintf(`resource "neon_project" "this" {
			name      				  = "%s"
			region_id 				  = "aws-us-west-2"
			pg_version				  = 14
			allowed_ips               = %s
		}`, projectName, ips)

		const resourceNameProject = "neon_project.this"
		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.3.0"), nil
					},
				},
				Steps: []resource.TestStep{
					{
						ResourceName: "project",
						Config:       resourceDefinition,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"name", projectName,
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"allowed_ips.#", fmt.Sprintf("%d", len(wantAllowedIPs)),
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"allowed_ips.0", wantAllowedIPs[0],
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"allowed_ips.1", wantAllowedIPs[1],
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

								var exceedingAllowedIPs []string

								missingIPs := map[string]struct{}{}
								for _, ip := range wantAllowedIPs {
									missingIPs[ip] = struct{}{}
								}

								for _, ip := range *ref.Settings.AllowedIps.Ips {
									if _, ok := missingIPs[ip]; ok {
										delete(missingIPs, ip)
										continue
									}

									exceedingAllowedIPs = append(exceedingAllowedIPs, ip)
								}

								if len(exceedingAllowedIPs) > 0 || len(missingIPs) > 0 {
									return fmt.Errorf("unexpected allowed ips. want=%v, got=%v",
										wantAllowedIPs, ref.Settings.AllowedIps.Ips)
								}

								if ref.Settings.AllowedIps.PrimaryBranchOnly {
									return errors.New("primary_branch_only is expected to be set to 'false'")
								}

								return nil
							},
						),
					},
				},
			},
		)
	})

	t.Run("shall provision a project with a custom list of allowed IPs set for default branch only", func(t *testing.T) {
		projectName := strconv.FormatInt(time.Now().UnixMilli(), 10)

		resourceDefinition := fmt.Sprintf(`resource "neon_project" "this" {
			name      				  = "%s"
			region_id 				  = "aws-us-west-2"
			pg_version				  = 14
			allowed_ips               = %s
			allowed_ips_primary_branch_only = true
		}`, projectName, ips)

		const resourceNameProject = "neon_project.this"
		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.3.0"), nil
					},
				},
				Steps: []resource.TestStep{
					{
						ResourceName: "project",
						Config:       resourceDefinition,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"name", projectName,
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"allowed_ips.#", fmt.Sprintf("%d", len(wantAllowedIPs)),
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"allowed_ips.0", wantAllowedIPs[0],
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"allowed_ips.1", wantAllowedIPs[1],
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

								var exceedingAllowedIPs []string

								missingIPs := map[string]struct{}{}
								for _, ip := range wantAllowedIPs {
									missingIPs[ip] = struct{}{}
								}

								for _, ip := range *ref.Settings.AllowedIps.Ips {
									if _, ok := missingIPs[ip]; ok {
										delete(missingIPs, ip)
										continue
									}

									exceedingAllowedIPs = append(exceedingAllowedIPs, ip)
								}

								if len(exceedingAllowedIPs) > 0 || len(missingIPs) > 0 {
									return fmt.Errorf("unexpected allowed ips. want=%v, got=%v",
										wantAllowedIPs, ref.Settings.AllowedIps.Ips)
								}

								if !ref.Settings.AllowedIps.PrimaryBranchOnly {
									return errors.New("primary_branch_only is expected to be set to 'true'")
								}

								return nil
							},
						),
					},
				},
			},
		)
	})
}

func projectLogicalReplication(t *testing.T, client *neon.Client) {
	t.Run("shall create project without logical replication", func(t *testing.T) {
		projectName := strconv.FormatInt(time.Now().UnixMilli(), 10)
		resourceDefinition := fmt.Sprintf(`resource "neon_project" "this" {
			name      				   = "%s"
			region_id 				   = "aws-us-west-2"
			pg_version				   = 16
		}`, projectName)
		const resourceNameProject = "neon_project.this"
		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.3.0"), nil
					},
				},
				Steps: []resource.TestStep{
					{
						ResourceName: "project",
						Config:       resourceDefinition,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"name", projectName,
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"enable_logical_replication", "false",
							),
							func(state *terraform.State) error {
								resp, err := client.ListProjects(nil, nil)
								if err != nil {
									return errors.New("listing error: " + err.Error())
								}

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

								if ref.Settings.EnableLogicalReplication == nil || *ref.Settings.EnableLogicalReplication {
									return errors.New("unexpected enable_logical_replication value, shall be 'false'")
								}

								return nil
							},
						),
					},
				},
			})
	})

	t.Run("shall create project with logical replication", func(t *testing.T) {
		projectName := strconv.FormatInt(time.Now().UnixMilli(), 10)

		resourceDefinition := fmt.Sprintf(`resource "neon_project" "this" {
			name      				   = "%s"
			region_id 				   = "aws-us-west-2"
			pg_version				   = 16
			enable_logical_replication = true
		}`, projectName)

		const resourceNameProject = "neon_project.this"
		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.3.0"), nil
					},
				},
				Steps: []resource.TestStep{
					{
						ResourceName: "project",
						Config:       resourceDefinition,
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"name", projectName,
							),
							resource.TestCheckResourceAttr(
								resourceNameProject,
								"enable_logical_replication", "true",
							),
							func(state *terraform.State) error {
								resp, err := client.ListProjects(nil, nil)
								if err != nil {
									return errors.New("listing error: " + err.Error())
								}

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

								if ref.Settings.EnableLogicalReplication == nil || !*ref.Settings.EnableLogicalReplication {
									return errors.New("unexpected enable_logical_replication value, shall be 'true'")
								}

								return nil
							},
						),
					},
				},
			},
		)
	})
}

func mustParseFloat64(s string) float64 {
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		panic(err)
	}
	return v
}
