package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceVPCEndpointRestriction() *schema.Resource {
	return &schema.Resource{
		Description: `Sets or updates a VPC endpoint restriction for a Neon project.
When a VPC endpoint restriction is set, the project only accepts connections
from the specified VPC.
A VPC endpoint can be set as a restriction only after it is assigned to the
parent organization of the Neon project.`,
		Importer: &schema.ResourceImporter{
			StateContext: resourceVPCEndpointRestrictionImport,
		},
		CreateContext: resourceVPCEndpointRestrictionCreateRetry,
		UpdateContext: resourceVPCEndpointRestrictionCreateRetry,
		ReadContext:   resourceVPCEndpointRestrictionReadRetry,
		DeleteContext: resourceVPCEndpointRestrictionDeleteRetry,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The Neon project ID.",
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

type vpcEndpointRestrictionID struct {
	ProjectID     string
	VpcEndpointID string
}

func (v vpcEndpointRestrictionID) ID() string {
	return fmt.Sprintf("%s/%s", v.VpcEndpointID, v.ProjectID)
}

func newVpcEndpointRestrictionID(s string) (o vpcEndpointRestrictionID, err error) {
	els := strings.SplitN(s, "/", 2)
	switch len(els) {
	case 2:
		o.VpcEndpointID = els[0]
		o.ProjectID = els[1]
	default:
		err = fmt.Errorf("invalid VPC endpoint restriction ID format: %s", s)
	}
	return o, err
}

func resourceVPCEndpointRestrictionCreate(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	id := vpcEndpointRestrictionID{
		ProjectID:     d.Get("project_id").(string),
		VpcEndpointID: d.Get("vpc_endpoint_id").(string),
	}
	err := meta.(*neon.Client).AssignProjectVPCEndpoint(id.ProjectID, id.VpcEndpointID,
		neon.VPCEndpointAssignment{
			Label: d.Get("label").(string),
		},
	)
	if err == nil {
		d.SetId(id.ID())
	}
	return err
}

func resourceVPCEndpointRestrictionRead(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	id, err := newVpcEndpointRestrictionID(d.Id())
	var resp neon.VPCEndpointsResponse
	if err == nil {
		resp, err = meta.(*neon.Client).ListProjectVPCEndpoints(id.ProjectID)
	}
	if err == nil {
		for _, el := range resp.Endpoints {
			if id.VpcEndpointID == el.VpcEndpointID {
				err = d.Set("label", el.Label)
				break
			}
		}
	}
	return err
}

func resourceVPCEndpointRestrictionDelete(_ context.Context, d *schema.ResourceData, meta interface{}) error {
	err := meta.(*neon.Client).DeleteProjectVPCEndpoint(
		d.Get("project_id").(string), d.Get("vpc_endpoint_id").(string),
	)
	if err == nil {
		d.SetId("")
	}
	return err
}

func resourceVPCEndpointRestrictionCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceVPCEndpointRestrictionCreate, ctx, d, meta)
}

func resourceVPCEndpointRestrictionReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceVPCEndpointRestrictionRead, ctx, d, meta)
}

func resourceVPCEndpointRestrictionDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceVPCEndpointRestrictionDelete, ctx, d, meta)
}

func resourceVPCEndpointRestrictionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceVPCEndpointRestrictionRead(ctx, d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
