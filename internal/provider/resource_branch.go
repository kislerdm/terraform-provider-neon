package provider

import (
	"context"
	"errors"
	"time"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceBranch() *schema.Resource {
	return &schema.Resource{
		Description:   "Project Branch. See details: https://neon.tech/docs/introduction/branching/",
		SchemaVersion: versionSchema,
		Importer: &schema.ResourceImporter{
			StateContext: resourceBranchImport,
		},
		CreateContext: resourceBranchCreate,
		ReadContext:   resourceBranchRead,
		UpdateContext: resourceBranchUpdate,
		DeleteContext: resourceBranchDelete,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				Description: "Project ID.",
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
			"parent_lsn": {
				Type:          schema.TypeString,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ConflictsWith: []string{"parent_timestamp"},
				Description: `Log Sequence Number (LSN) horizon for the data to be present in the new branch.
See details: https://neon.tech/docs/reference/glossary/#lsn`,
			},
			"parent_timestamp": {
				Type:     schema.TypeInt,
				Optional: true,
				Computed: true,
				ForceNew: true,
				ValidateFunc: func(i interface{}, s string) (warns []string, errs []error) {
					if i.(int) < 0 {
						errs = append(errs, errors.New("timestamp must be not negative"))
						return
					}
					return
				},
				ConflictsWith: []string{"parent_lsn"},
				Description: `Timestamp horizon for the data to be present in the new branch. 
**Note**: it's defined as Unix epoch.'`,
			},
			"physical_size_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Branch physical size in MB.",
			},
			"logical_size": {
				Type:        schema.TypeInt,
				Computed:    true,
				Description: "Branch logical size in MB.",
			},
			"current_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Branch state.",
			},
			"pending_state": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Branch pending state.",
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Branch creation timestamp.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Branch last update timestamp.",
			},
			"endpoints": {
				Type:        schema.TypeList,
				Optional:    true,
				MaxItems:    1,
				Description: "Endpoints for the branch.",
				Elem: map[string]*schema.Schema{
					"type": {
						Type:         schema.TypeString,
						Required:     true,
						InputDefault: "read_write",
						Description: `Endpoint type. 
Either "read_write" for read-write primary or "read_only" for read-only secondary.
**Note**: "read_only" endpoints are NOT yet implemented.`,
					},
					"autoscaling_limit_min_cu": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"autoscaling_limit_max_cu": {
						Type:     schema.TypeInt,
						Optional: true,
					},
					"host": {
						Type:        schema.TypeString,
						Computed:    true,
						Description: "Hostname to connect to.",
					},
					"disabled": {
						Type:        schema.TypeBool,
						Computed:    true,
						Description: "Restrict any connections to this endpoint.",
					},
					"pg_settings": {
						Type:     schema.TypeMap,
						Optional: true,
						Computed: true,
					},
				},
			},
		},
	}
}

func updateStateBranch(d *schema.ResourceData, r neon.Branch) {
	_ = d.Set("project_id", r.ProjectID)
	_ = d.Set("name", r.Name)
	_ = d.Set("parent_id", r.ParentID)
	_ = d.Set("parent_lsn", r.ParentLsn)
	_ = d.Set("parent_timestamp", int(r.ParentTimestamp.Unix()))
	_ = d.Set("logical_size", int(r.LogicalSize))
	_ = d.Set("physical_size_size", int(r.PhysicalSize))
	_ = d.Set("current_state", r.CurrentState)
	_ = d.Set("pending_state", r.PendingState)
	_ = d.Set("created_at", r.CreatedAt.Format(time.RFC3339))
	_ = d.Set("updated_at", r.CreatedAt.Format(time.RFC3339))
}

func updateStateBranchEndpoints(d *schema.ResourceData, r neon.Endpoint) {
	panic("todo")
}

func resourceBranchDelete(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	panic("todo")
}

func resourceBranchUpdate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	panic("todo")
}

func resourceBranchRead(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	panic("todo")
}

func resourceBranchCreate(ctx context.Context, data *schema.ResourceData, i interface{}) diag.Diagnostics {
	panic("todo")
}

func resourceBranchImport(ctx context.Context, data *schema.ResourceData, i interface{}) (
	[]*schema.ResourceData, error,
) {
	panic("todo")
}
