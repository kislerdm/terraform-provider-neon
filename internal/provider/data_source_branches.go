package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func dataSourceBranches() *schema.Resource {
	return &schema.Resource{
		Description:   "Fetch Project Branches.",
		SchemaVersion: 1,
		ReadContext:   dataSourceBranchesRead,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project ID.",
			},
			"branches": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Branch ID.",
						},
						"name": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "Branch name.",
						},
						"parent_id": {
							Type:        schema.TypeString,
							Optional:    true,
							Computed:    true,
							Description: "ID of the branch to checkout.",
						},
						"logical_size": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: "Branch logical size in MB.",
						},
						"primary": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Primary branch flag.",
						},
					},
				},
			},
		},
	}
}

func dataSourceBranchesRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "read Branches")

	projectID := d.Get("project_id").(string)

	d.SetId(fmt.Sprintf("%s/branches", projectID))

	resp, err := meta.(*neon.Client).ListProjectBranches(projectID)
	if err != nil {
		diag.FromErr(err)
	}

	var branches []map[string]interface{}
	for _, v := range resp.Branches {
		parentID := ""
		if v.ParentID != nil {
			parentID = *v.ParentID
		}
		logicalSize := int64(0)
		if v.LogicalSize != nil {
			logicalSize = *v.LogicalSize
		}

		branches = append(branches, map[string]interface{}{
			"id":           v.ID,
			"name":         v.Name,
			"parent_id":    parentID,
			"logical_size": logicalSize,
			"primary":      v.Primary,
		})
	}

	if err := d.Set("branches", branches); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(nil)
}
