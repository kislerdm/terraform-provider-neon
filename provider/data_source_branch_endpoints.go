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
						"host_pooling": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Endpoint URI for connection pooling.",
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
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Compute name.",
						},
						"autoscaling_limit_min_cu": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Minimal value of the compute autoscaling limit.",
						},
						"autoscaling_limit_max_cu": {
							Type:        schema.TypeFloat,
							Computed:    true,
							Description: "Maximal value of the compute autoscaling limit.",
						},
						"suspend_timeout_seconds": {
							Type:        schema.TypeInt,
							Computed:    true,
							Description: `Duration of inactivity in seconds after which the compute endpoint is automatically suspended.`,
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
		el := map[string]interface{}{
			"id":                       v.ID,
			"host":                     v.Host,
			"type":                     v.Type.String(),
			"region_id":                v.RegionID,
			"proxy_host":               v.ProxyHost,
			"host_pooling":             newPooledHost(v.Host),
			"suspend_timeout_seconds":  int(v.SuspendTimeoutSeconds),
			"autoscaling_limit_min_cu": float64(v.AutoscalingLimitMinCu),
			"autoscaling_limit_max_cu": float64(v.AutoscalingLimitMinCu),
		}
		if v.Name != nil {
			el["name"] = *v.Name
		}
		endpoints = append(endpoints, el)
	}

	if err := d.Set("endpoints", endpoints); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(nil)
}
