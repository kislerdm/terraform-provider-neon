package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func dataSourceBranchEndpoints() *schema.Resource {
	return &schema.Resource{
		Description:   "Fetch Branch Endpoints",
		SchemaVersion: 1,
		ReadContext:   dataSourceBranchEndpointsRead,
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
			"endpoints": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Endpoint ID.",
						},
						"host": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Endpoint URI.",
						},
						"type": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: `Access type.`,
						},
						"region_id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Deployment region: https://neon.tech/docs/introduction/regions",
						},
						"proxy_host": {
							Type:     schema.TypeString,
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func dataSourceBranchEndpointsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "read Endpoints")

	projectID, _ := d.Get("project_id").(string)
	branchID, _ := d.Get("branch_id").(string)

	d.SetId(projectID + "/" + branchID)

	resp, err := meta.(*neon.Client).ListProjectBranchEndpoints(
		projectID,
		branchID,
	)
	if err != nil {
		diag.FromErr(err)
	}

	var endpoints []map[string]interface{}
	for _, v := range resp.Endpoints {
		endpoints = append(endpoints, map[string]interface{}{
			"id":         v.ID,
			"host":       v.Host,
			"type":       string(v.Type),
			"region_id":  v.RegionID,
			"proxy_host": v.ProxyHost,
		})
	}

	if err := d.Set("endpoints", endpoints); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(nil)
}
