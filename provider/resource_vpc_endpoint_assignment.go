package provider

import (
	"context"
	"errors"
	"strings"

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
				ForceNew:    true,
				Description: "The Neon organization ID.",
			},
			"region_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Neon region ID.",
			},
			"vpc_endpoint_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
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
	err := meta.(sdkVPCEndpoint).AssignOrganizationVPCEndpoint(
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
	resp, err := meta.(sdkVPCEndpoint).GetOrganizationVPCEndpointDetails(
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
	err := meta.(sdkVPCEndpoint).DeleteOrganizationVPCEndpoint(
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
	r, err := parseVPCEndpointAssignmentID(d.Id())
	if err != nil {
		return nil, err
	}

	setResourceAttrsFromVPCEndpointAssignmentID(d, r)
	if err := resourceVPCEndpointAssignmentRead(ctx, d, meta); err != nil {
		return nil, err
	}

	// Set ID to vpc_endpoint_id only for backwards compatibility with existing state
	d.SetId(r.VPCEndpointID)

	return []*schema.ResourceData{d}, nil
}

type sdkVPCEndpoint interface {
	AssignOrganizationVPCEndpoint(string, string, string, neon.VPCEndpointAssignment) error
	GetOrganizationVPCEndpointDetails(string, string, string) (neon.VPCEndpointDetails, error)
	DeleteOrganizationVPCEndpoint(string, string, string) error
}

type vpcEndpointAssignmentID struct {
	OrgID, RegionID, VPCEndpointID string
}

func setResourceAttrsFromVPCEndpointAssignmentID(d *schema.ResourceData, r vpcEndpointAssignmentID) {
	_ = d.Set("org_id", r.OrgID)
	_ = d.Set("region_id", r.RegionID)
	_ = d.Set("vpc_endpoint_id", r.VPCEndpointID)
}

func parseVPCEndpointAssignmentID(s string) (vpcEndpointAssignmentID, error) {
	spl := strings.Split(s, "/")
	if len(spl) != 3 {
		return vpcEndpointAssignmentID{}, errors.New(
			"ID of this resource type shall follow the template: {{.OrgID}}/{{.RegionID}}/{{.VPCEndpointID}}",
		)
	}
	return vpcEndpointAssignmentID{
		OrgID:          spl[0],
		RegionID:       spl[1],
		VPCEndpointID:  spl[2],
	}, nil
}

