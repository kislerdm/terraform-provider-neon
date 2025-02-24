package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceVPCEndpointAssignment() *schema.Resource {
	return &schema.Resource{
		Description: `Assigns, or updates existing assignment of a VPC endpoint to a Neon organization.
See details: https://neon.tech/docs/guides/neon-private-networking#enable-private-dns
`,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCEndpointAssignmentImport,
		},
		CreateContext: resourceVPCEndpointAssignmentCreateRetry,
		UpdateContext: resourceVPCEndpointAssignmentCreateRetry,
		ReadContext:   resourceVPCEndpointAssignmentReadRetry,
		DeleteContext: resourceVPCEndpointAssignmentDeleteRetry,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Neon organization ID.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The Neon region ID.",
				ValidateFunc: func(v interface{}, s string) (warn []string, errs []error) {
					switch vv := v.(string); vv {
					case "us-east-1",
						"us-east-2",
						"eu-central-1",
						"us-west-2",
						"ap-southeast-1",
						"ap-southeast-2":
					default:
						errs = append(errs, fmt.Errorf("%v is not supported value for %s", v, s))
					}
					return warn, errs
				},
			},
			"vpc_endpoint_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "The VPC endpoint ID.",
			},
			"label": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "A descriptive label for the VPC endpoint.",
			},
		},
	}
}

func resourceVPCEndpointAssignmentCreate(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	err := meta.(*neon.Client).AssignOrganizationVPCEndpoint(
		d.Get("org_id").(string),
		d.Get("region_id").(string),
		d.Get("vpc_endpoint_id").(string),
		neon.VPCEndpointAssignment{
			Label: d.Get("label").(string),
		},
	)
	if err == nil {
		d.SetId(d.Get("vpc_endpoint_id").(string))
	}
	return err
}

func resourceVPCEndpointAssignmentRead(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	resp, err := meta.(*neon.Client).GetOrganizationVPCEndpointDetails(
		d.Get("org_id").(string),
		d.Get("region_id").(string),
		d.Get("vpc_endpoint_id").(string),
	)
	if err == nil {
		err = d.Set("label", resp.Label)
	}
	return err
}

func resourceVPCEndpointAssignmentDelete(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	err := meta.(*neon.Client).DeleteOrganizationVPCEndpoint(
		d.Get("org_id").(string),
		d.Get("region_id").(string),
		d.Get("vpc_endpoint_id").(string),
	)
	if err == nil {
		d.SetId("")
	}
	return err
}

func resourceVPCEndpointAssignmentCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceVPCEndpointAssignmentCreate, ctx, d, meta)
}

func resourceVPCEndpointAssignmentReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceVPCEndpointAssignmentRead, ctx, d, meta)
}

func resourceVPCEndpointAssignmentDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceVPCEndpointAssignmentDelete, ctx, d, meta)
}

func resourceVPCEndpointAssignmentImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceVPCEndpointAssignmentRead(ctx, d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
