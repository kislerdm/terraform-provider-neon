package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func dataSourceRegions() *schema.Resource {
	return &schema.Resource{
		Description:   "Fetch available Neon regions.",
		SchemaVersion: 1,
		ReadContext:   dataSourceRegionsRead,
		Schema: map[string]*schema.Schema{
			"regions": {
				Type:        schema.TypeList,
				Computed:    true,
				Description: "List of available Neon regions.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Region ID used in API endpoints (e.g., aws-us-east-1).",
						},
						"name": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Human-readable region description.",
						},
						"default": {
							Type:        schema.TypeBool,
							Computed:    true,
							Description: "Whether this region is the default for new projects.",
						},
						"geo_lat": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Approximate geographical latitude.",
						},
						"geo_long": {
							Type:        schema.TypeString,
							Computed:    true,
							Description: "Approximate geographical longitude.",
						},
					},
				},
			},
		},
	}
}

func dataSourceRegionsRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "read Regions")

	resp, err := meta.(*neon.Client).GetActiveRegions()
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("regions")

	var regions []map[string]interface{}
	for _, v := range resp.Regions {
		regions = append(regions, map[string]interface{}{
			"id":       v.RegionID,
			"name":     v.Name,
			"default":  v.Default,
			"geo_lat":  v.GeoLat,
			"geo_long": v.GeoLong,
		})
	}

	if err := d.Set("regions", regions); err != nil {
		return diag.FromErr(err)
	}

	return nil
}
