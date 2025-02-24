package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func dataSourceBranchRolePassword() *schema.Resource {
	return &schema.Resource{
		Description:   "Fetch Role Password.",
		SchemaVersion: 1,
		ReadContext:   dataSourceBranchRolePasswordRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project ID.",
			},
			"branch_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Branch ID.",
			},
			"role_name": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Role name.",
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Password.",
			},
		},
	}
}

func dataSourceBranchRolePasswordRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "read Roles")

	projectID := d.Get("project_id").(string)
	branchID := d.Get("branch_id").(string)
	roleName := d.Get("role_name").(string)

	d.SetId(fmt.Sprintf("%s/%s/%s/password", projectID, branchID, roleName))

	resp, err := meta.(*neon.Client).GetProjectBranchRolePassword(
		projectID,
		branchID,
		roleName,
	)

	if err != nil {
		diag.FromErr(err)
	}

	if err := d.Set("password", resp.Password); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(nil)
}
