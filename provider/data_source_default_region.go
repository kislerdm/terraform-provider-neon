package provider

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func dataSourceDefaultRegion() *schema.Resource {
	return &schema.Resource{
		Description:   "Fetch the default Neon region for new projects.",
		SchemaVersion: 1,
		ReadContext:   dataSourceDefaultRegionRead,
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
	}
}

func dataSourceDefaultRegionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "read DefaultRegion")

	resp, err := meta.(*neon.Client).GetActiveRegions()
	if err != nil {
		return diag.FromErr(err)
	}

	for _, v := range resp.Regions {
		if v.Default {
			d.SetId(v.RegionID)

			if err := d.Set("id", v.RegionID); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("name", v.Name); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("geo_lat", v.GeoLat); err != nil {
				return diag.FromErr(err)
			}
			if err := d.Set("geo_long", v.GeoLong); err != nil {
				return diag.FromErr(err)
			}

			return nil
		}
	}

	return diag.FromErr(errors.New("no default region found"))
}
