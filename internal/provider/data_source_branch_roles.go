package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func dataSourceBranchRoles() *schema.Resource {
	return &schema.Resource{
		Description:   "Fetch Branch Roles.",
		SchemaVersion: 1,
		ReadContext:   dataSourceBranchRolesRead,
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
			"roles": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Role name.",
						},
						"protected": {
							Type:     schema.TypeBool,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBranchRolesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "read Roles")

	projectID := d.Get("project_id").(string)
	branchID := d.Get("branch_id").(string)

	d.SetId(fmt.Sprintf("%s/%s/roles", projectID, branchID))

	resp, err := meta.(*neon.Client).ListProjectBranchRoles(
		projectID,
		branchID,
	)

	if err != nil {
		diag.FromErr(err)
	}

	var roles []map[string]interface{}
	for _, v := range resp.Roles {
		protected := true
		if v.Protected != nil {
			protected = *v.Protected
		}

		roles = append(roles, map[string]interface{}{
			"name":      v.Name,
			"protected": protected,
		})
	}

	if err := d.Set("roles", roles); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(nil)
}
