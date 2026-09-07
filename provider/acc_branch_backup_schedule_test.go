package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	neon "github.com/kislerdm/neon-sdk-go"
	"github.com/stretchr/testify/assert"
)

func TestBranchBackupSchedule(t *testing.T) {
	if os.Getenv("TF_ACC") != "1" {
		t.Skip("TF_ACC must be set to 1")
	}

	client, err := neon.NewClient(neon.Config{Key: os.Getenv("NEON_API_KEY")})
	if err != nil {
		t.Fatal(err)
	}

	projectNamePrefix := "branchBackupSchedule"

	t.Cleanup(func() {
		resp, _ := client.ListProjects(nil, nil, &projectNamePrefix, nil, nil, nil)
		for _, project := range resp.Projects {
			_, _ = client.DeleteProject(project.ID)
		}
	})

	projectName := newProjectName(projectNamePrefix)
	resource.Test(
		t, resource.TestCase{
			ProtoV6ProviderFactories: map[string]func() (tfprotov6.ProviderServer, error){
				"neon": func() (tfprotov6.ProviderServer, error) {
					return newAccTestFramework(), nil
				},
			},
			Steps: []resource.TestStep{
				{
					Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}

resource "neon_branch_backup_schedule" "this" {
  project_id = neon_project.this.id
  branch_id  = neon_project.this.default_branch_id
  schedule = [
    {
      frequency         = "weekly"
      day               = 1
      hour              = 4
      retention_seconds = 86400 * 7
    }
  ]
}
`, projectName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.#", "1",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.frequency", "weekly",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.day", "1",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.hour", "4",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.retention_seconds", "604800",
						),
						func(_ *terraform.State) error {
							pr, err := readProjectInfo(client, projectName)
							if err != nil {
								return err
							}

							br, err := client.ListProjectBranches(pr.ID, nil, nil, nil, nil,
								nil, nil)
							if err != nil {
								return err
							}
							for _, el := range br.Branches {
								schedule, err := client.GetSnapshotSchedule(pr.ID, el.ID)
								if err != nil {
									return err
								}
								assert.Len(t, schedule.Schedule, 1)
								assert.Equal(t, "weekly", schedule.Schedule[0].Frequency)
								assert.Equal(t, uint8(1), *schedule.Schedule[0].Day)
								assert.Equal(t, uint8(4), *schedule.Schedule[0].Hour)
								assert.Equal(t, uint32(604800), *schedule.Schedule[0].RetentionSeconds)
							}

							return nil
						},
					),
				},
				// update: mutate existing monthly schedule
				{
					Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}

resource "neon_branch_backup_schedule" "this" {
  project_id = neon_project.this.id
  branch_id  = neon_project.this.default_branch_id
  schedule = [
    {
      frequency         = "monthly"
      day               = 1
      hour              = 4
      retention_seconds = 86400 * 30
    }
  ]
}
`, projectName),
					Check: resource.ComposeTestCheckFunc(
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.#", "1",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.frequency", "monthly",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.day", "1",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.hour", "4",
						),
						resource.TestCheckResourceAttr(
							"neon_branch_backup_schedule.this",
							"schedule.0.retention_seconds", "2592000",
						),
						func(_ *terraform.State) error {
							pr, err := readProjectInfo(client, projectName)
							if err != nil {
								return err
							}

							br, err := client.ListProjectBranches(pr.ID, nil, nil, nil, nil,
								nil, nil)
							if err != nil {
								return err
							}
							for _, el := range br.Branches {
								schedule, err := client.GetSnapshotSchedule(pr.ID, el.ID)
								if err != nil {
									return err
								}
								assert.Len(t, schedule.Schedule, 1)
								assert.Equal(t, "monthly", schedule.Schedule[0].Frequency)
								assert.Equal(t, uint8(1), *schedule.Schedule[0].Day)
								assert.Equal(t, uint8(4), *schedule.Schedule[0].Hour)
								assert.Equal(t, uint32(2592000), *schedule.Schedule[0].RetentionSeconds)
							}

							return nil
						},
					),
				},
				// delete
				{
					Config: fmt.Sprintf(`resource "neon_project" "this" {name = "%s"}`, projectName),
					Check: resource.ComposeTestCheckFunc(
						func(_ *terraform.State) error {
							pr, err := readProjectInfo(client, projectName)
							if err != nil {
								return err
							}

							br, err := client.ListProjectBranches(pr.ID, nil, nil, nil, nil,
								nil, nil)
							if err != nil {
								return err
							}
							for _, el := range br.Branches {
								schedule, err := client.GetSnapshotSchedule(pr.ID, el.ID)
								if err != nil {
									return err
								}
								// the API-set schedule remains set in Neon server
								assert.Len(t, schedule.Schedule, 1)
							}

							return nil
						},
					),
				},
			},
		})
}
