package provider

import (
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func TestResourcesImport(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	var (
		projectID      string
		customBranchID string
	)
	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_, _ = client.DeleteProject(projectID)
	})

	// GIVEN the project
	projectName := newProjectName()
	const (
		defaultDatabaseName = "db-foo"
		defaultRoleName     = "r-foo"
		defaultBranchName   = "br-foo"
	)

	respProject, err := client.CreateProject(neon.ProjectCreateRequest{
		Project: neon.ProjectCreateRequestProject{
			Branch: &neon.ProjectCreateRequestProjectBranch{
				DatabaseName: pointer(defaultDatabaseName),
				Name:         pointer(defaultBranchName),
				RoleName:     pointer(defaultRoleName),
			},
			Name: pointer(projectName),
		}},
	)
	if err != nil {
		t.Fatal(err)
	}
	projectID = respProject.Project.ID
	defaultBranchID := respProject.Branch.ID
	defaultEndpointID := respProject.Endpoints[0].ID
	defaultHost := respProject.Endpoints[0].Host
	defaultRolePassword := *respProject.Roles[0].Password

	sleepDuringRunningOperations(t, client, projectID)

	t.Run("shall successfully import the project", func(t *testing.T) {
		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.5.0"), nil
					},
				},
				Steps: []resource.TestStep{
					{
						ResourceName: "neon_project.this",
						Config: fmt.Sprintf(`resource "neon_project" "this" {
  name = "%s"
  branch {
    name          = "%s"
    database_name = "%s"
    role_name     = "%s"
  }
}`, projectName, defaultBranchName, defaultDatabaseName, defaultRoleName),
						// WHEN run terraform import
						ImportState:   true,
						ImportStateId: projectID,
						Check: resource.ComposeTestCheckFunc(
							// THEN
							resource.TestCheckResourceAttr("neon_project.this", "id", projectID),
							resource.TestCheckResourceAttr("neon_project.this", "default_branch_id", defaultBranchID),
							resource.TestCheckResourceAttr("neon_project.this", "default_endpoint_id", defaultEndpointID),
							resource.TestCheckResourceAttr("neon_project.this", "database_host", defaultHost),
							resource.TestCheckResourceAttr("neon_project.this", "database_user", defaultRoleName),
							resource.TestCheckResourceAttr("neon_project.this", "database_name", defaultDatabaseName),
							resource.TestCheckResourceAttr("neon_project.this", "database_password", defaultRolePassword),
						),
					},
				},
			},
		)
	})

	// See https://github.com/kislerdm/terraform-provider-neon/issues/88
	t.Run("shall successfully import the role", func(t *testing.T) {
		// GIVEN a custom role
		const customRoleName = "r-bar"
		respRole, err := client.CreateProjectBranchRole(projectID, defaultBranchID, neon.RoleCreateRequest{
			Role: neon.RoleCreateRequestRole{
				Name: customRoleName,
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		customRoleID := complexID{
			ProjectID: projectID,
			BranchID:  defaultBranchID,
			Name:      customRoleName,
		}
		customRolePassword := *respRole.Role.Password

		sleepDuringRunningOperations(t, client, projectID)

		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.5.0"), nil
					},
				},
				Steps: []resource.TestStep{
					{
						ResourceName: "neon_role.this",
						Config: fmt.Sprintf(`resource "neon_role" "this" {
  project_id = "%s"
  branch_id  = "%s"
  name       = "%s"
}`, projectID, defaultBranchID, customRoleName),
						// WHEN run terraform import
						ImportState:   true,
						ImportStateId: customRoleID.toString(),
						Check: resource.ComposeTestCheckFunc(
							// THEN
							resource.TestCheckResourceAttr("neon_role.this", "name", customRoleName),
							resource.TestCheckResourceAttr("neon_role.this", "password", customRolePassword),
						),
					},
				},
			},
		)
	})

	t.Run("shall successfully import the database", func(t *testing.T) {
		// GIVEN a custom database
		const customDatabaseName = "db-bar"
		_, err = client.CreateProjectBranchDatabase(projectID, defaultBranchID, neon.DatabaseCreateRequest{
			Database: neon.DatabaseCreateRequestDatabase{
				Name:      customDatabaseName,
				OwnerName: defaultRoleName,
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		customDatabaseID := complexID{
			ProjectID: projectID,
			BranchID:  defaultBranchID,
			Name:      customDatabaseName,
		}

		sleepDuringRunningOperations(t, client, projectID)

		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.5.0"), nil
					},
				},
				Steps: []resource.TestStep{
					// WHEN run terraform import
					{
						ResourceName: "neon_database.this",
						Config: fmt.Sprintf(`resource "neon_database" "this" {
  name       = "%s"
  project_id = "%s"
  branch_id  = "%s"
  owner_name = "%s"
}`, customDatabaseName, projectID, defaultBranchID, defaultRoleName),
						ImportState:   true,
						ImportStateId: customDatabaseID.toString(),
						Check: resource.ComposeTestCheckFunc(
							// THEN
							resource.TestCheckResourceAttr("neon_database.this", "name", customDatabaseName),
						),
					},
				},
			},
		)
	})

	t.Run("branch-endpoint", func(t *testing.T) {
		// GIVEN a custom branch
		const customBranchName = "br-bar"
		respBranch, err := client.CreateProjectBranch(projectID, &neon.BranchCreateRequest{
			Branch: &neon.BranchCreateRequestBranch{
				Name: pointer(customBranchName),
			},
		})
		if err != nil {
			t.Fatal(err)
		}

		customBranchID = respBranch.BranchResponse.Branch.ID

		sleepDuringRunningOperations(t, client, projectID)

		t.Run("shall successfully import the branch", func(t *testing.T) {
			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: map[string]func() (*schema.Provider, error){
						"neon": func() (*schema.Provider, error) {
							return New("0.5.0"), nil
						},
					},
					Steps: []resource.TestStep{
						// WHEN run terraform import
						{
							ResourceName: "neon_branch.this",
							Config: fmt.Sprintf(`resource "neon_branch" "this" {
  name       = "%s"
  project_id = "%s"
}`, customBranchName, projectID),
							ImportState:   true,
							ImportStateId: customBranchID,
							Check: resource.ComposeTestCheckFunc(
								// THEN
								resource.TestCheckResourceAttr("neon_branch.this", "name", customBranchName),
								resource.TestCheckResourceAttr("neon_branch.this", "parent_id", defaultBranchID),
							),
						},
					},
				},
			)
		})

		// Note that the only support endpoint type is "read_write",
		// and a single endpoint of that type is allowed per branch.
		// Hence, the endpoint is provisioned for a custom branch, which must be provisioned beforehand.
		// It's the reason why both tests are in the scope of the same function.
		t.Run("shall successfully import the endpoint", func(t *testing.T) {
			// GIVEN a custom endpoint
			respEp, err := client.CreateProjectEndpoint(projectID, neon.EndpointCreateRequest{
				Endpoint: neon.EndpointCreateRequestEndpoint{
					BranchID: customBranchID,
					Type:     endpointTypeRW,
				},
			})
			if err != nil {
				t.Fatal(err)
			}

			customEndpointID := respEp.EndpointResponse.Endpoint.ID

			sleepDuringRunningOperations(t, client, projectID)

			resource.UnitTest(
				t, resource.TestCase{
					ProviderFactories: map[string]func() (*schema.Provider, error){
						"neon": func() (*schema.Provider, error) {
							return New("0.5.0"), nil
						},
					},
					Steps: []resource.TestStep{
						{
							ResourceName: "neon_endpoint.this",
							Config: fmt.Sprintf(`resource "neon_endpoint" "this" {
	project_id = "%s"
	branch_id  = "%s"
}`, projectID, customBranchID),
							// WHEN run terraform import
							ImportState:   true,
							ImportStateId: customEndpointID,
							Check: resource.ComposeTestCheckFunc(
								// THEN
								// endpointTypeRW is the default type
								resource.TestCheckResourceAttr("neon_endpoint.this", "type", endpointTypeRW),
							),
						},
					},
				},
			)
		})
	})

	t.Run("shall successfully import the project permission", func(t *testing.T) {
		// GIVEN a custom permission project grant
		const granteeEmail = "foo@bar.baz"
		respGrant, err := client.GrantPermissionToProject(projectID, neon.GrantPermissionToProjectRequest{
			Email: granteeEmail,
		})
		if err != nil {
			t.Fatal(err)
		}
		grantID := joinedIDProjectPermission{
			projectID:    projectID,
			permissionID: respGrant.ID,
		}

		sleepDuringRunningOperations(t, client, projectID)

		resource.UnitTest(
			t, resource.TestCase{
				ProviderFactories: map[string]func() (*schema.Provider, error){
					"neon": func() (*schema.Provider, error) {
						return New("0.5.0"), nil
					},
				},
				Steps: []resource.TestStep{
					{
						ResourceName: "neon_project_permission.this",
						Config: fmt.Sprintf(`resource "neon_project_permission" "this" {
	project_id = "%s"
	grantee    = "%s"
}`, projectID, granteeEmail),
						// WHEN run terraform import
						ImportState:   true,
						ImportStateId: grantID.ToString(),
						// THEN
						Check: resource.ComposeTestCheckFunc(
							resource.TestCheckResourceAttr("neon_project_permission.this", "grantee", granteeEmail),
						),
					},
				},
			},
		)
	})
}

func sleepDuringRunningOperations(t *testing.T, client *neon.Client, projectID string) {
	t.Helper()

	respOps, err := client.ListProjectOperations(projectID, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	operations := respOps.OperationsResponse.Operations

	var isOperationsFinished bool
	for !isOperationsFinished {
		var isFinished = make([]bool, len(operations))
		for i, operation := range operations {
			isFinished[i] = operation.Status == "finished"
		}

		isOperationsFinished = !slices.Contains(isFinished, false)
		if !isOperationsFinished {
			time.Sleep(500 * time.Millisecond)
			respOps, err := client.ListProjectOperations(projectID, nil, nil)
			if err != nil {
				t.Fatal(err)
			}
			operations = respOps.OperationsResponse.Operations
		}
	}
}
