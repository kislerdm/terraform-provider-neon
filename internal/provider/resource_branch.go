package provider

import (
	"context"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
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
		CreateContext: resourceBranchCreateRetry,
		ReadContext:   resourceBranchReadRetry,
		UpdateContext: resourceBranchUpdateRetry,
		DeleteContext: resourceBranchDeleteRetry,
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
				Type:          schema.TypeInt,
				Optional:      true,
				Computed:      true,
				ForceNew:      true,
				ValidateFunc:  intValidationNotNegative,
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
		},
	}
}

func updateStateBranch(d *schema.ResourceData, v neon.Branch) error {
	if err := d.Set("name", v.Name); err != nil {
		return err
	}
	if err := d.Set("parent_id", v.ParentID); err != nil {
		return err
	}
	if err := d.Set("parent_lsn", v.ParentLsn); err != nil {
		return err
	}
	if err := d.Set("parent_timestamp", int(v.ParentTimestamp.Unix())); err != nil {
		return err
	}
	if err := d.Set("logical_size", int(v.LogicalSize)); err != nil {
		return err
	}
	if err := d.Set("physical_size_size", int(v.PhysicalSize)); err != nil {
		return err
	}
	if err := d.Set("current_state", v.CurrentState); err != nil {
		return err
	}
	if err := d.Set("pending_state", v.PendingState); err != nil {
		return err
	}
	if err := d.Set("created_at", v.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", v.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	return nil
}

func resourceBranchCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceBranchCreate, ctx, d, meta)
}

func resourceBranchReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceBranchRead, ctx, d, meta)
}

func resourceBranchUpdateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceBranchUpdate, ctx, d, meta)
}

func resourceBranchDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceBranchDelete, ctx, d, meta)
}

func resourceBranchCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "created Branch")

	cfg := neon.BranchCreateRequest{
		Branch: neon.BranchCreateRequestBranch{
			ParentID:  d.Get("parent_id").(string),
			Name:      d.Get("name").(string),
			ParentLsn: d.Get("parent_lsn").(string),
		},
	}

	if v, ok := d.GetOk("parent_timestamp"); ok && v.(int) > 0 {
		t := time.Unix(int64(v.(int)), 0)
		cfg.Branch.ParentTimestamp = &t
	}

	resp, err := meta.(neon.Client).CreateProjectBranch(
		d.Get("project_id").(string),
		&cfg,
	)
	if err != nil {
		return err
	}

	d.SetId(resp.Branch.ID)
	return updateStateBranch(d, resp.Branch)
}

func resourceBranchUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "update Branch")

	v, ok := d.GetOk("name")
	if !ok || v.(string) == "" {
		return nil
	}

	cfg := neon.BranchUpdateRequest{
		Branch: neon.BranchUpdateRequestBranch{
			Name: v.(string),
		},
	}

	resp, err := meta.(neon.Client).UpdateProjectBranch(d.Get("project_id").(string), d.Id(), cfg)
	if err != nil {
		return err
	}

	return updateStateBranch(d, resp.Branch)
}

func resourceBranchRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "read Branch")

	resp, err := meta.(neon.Client).GetProjectBranch(d.Get("project_id").(string), d.Id())
	if err != nil {
		return err
	}

	return updateStateBranch(d, resp.Branch)
}

func resourceBranchDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Branch")

	if _, err := meta.(neon.Client).DeleteProjectBranch(d.Get("project_id").(string), d.Id()); err != nil {
		return err
	}

	d.SetId("")
	return updateStateBranch(d, neon.Branch{})
}

func resourceBranchImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	if err := resourceBranchRead(ctx, d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
