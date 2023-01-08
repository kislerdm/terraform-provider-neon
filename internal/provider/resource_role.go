package provider

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceRole() *schema.Resource {
	return &schema.Resource{
		Description: `Project Role. **Note** that User and Role are synonymous terms in Neon. 
See details: https://neon.tech/docs/manage/users/

~> Password will be reset upon import of the Role resource.
`,
		SchemaVersion: versionSchema,
		Importer: &schema.ResourceImporter{
			StateContext: resourceRoleImport,
		},
		CreateContext: resourceRoleCreateRetry,
		ReadContext:   resourceRoleReadRetry,
		DeleteContext: resourceRoleDeleteRetry,
		Schema: map[string]*schema.Schema{
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID.",
			},
			"branch_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Branch ID.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Role name.",
			},
			"password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Database authentication password.",
			},
			"protected": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"created_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Role creation timestamp.",
			},
			"updated_at": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Role last update timestamp.",
			},
		},
	}
}

func updateStateRole(d *schema.ResourceData, v neon.Role) error {
	if err := d.Set("protected", v.Protected); err != nil {
		return err
	}
	if err := d.Set("created_at", v.CreatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	if err := d.Set("updated_at", v.UpdatedAt.Format(time.RFC3339)); err != nil {
		return err
	}
	return nil
}

type roleID struct {
	ProjectID, BranchID, Name string
}

func setResourceDataFrom(d *schema.ResourceData, r roleID) {
	_ = d.Set("project_id", r.ProjectID)
	_ = d.Set("branch_id", r.BranchID)
	_ = d.Set("name", r.Name)
}

func (v roleID) toString() string {
	return v.ProjectID + "/" + v.BranchID + "/" + v.Name
}

func parseRoleID(s string) (roleID, error) {
	spl := strings.Split(s, "/")
	if len(spl) != 3 {
		return roleID{}, errors.New("role ID must follow the format: {{.ProjectID}}/{{.BranchID}}/{{.RoleName}}")
	}
	return roleID{
		ProjectID: spl[0],
		BranchID:  spl[1],
		Name:      spl[2],
	}, nil
}

func resourceRoleCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceRoleCreate, ctx, d, meta)
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "created Role")

	r := roleID{
		ProjectID: d.Get("project_id").(string),
		BranchID:  d.Get("branch_id").(string),
		Name:      d.Get("name").(string),
	}
	resp, err := meta.(neon.Client).CreateProjectBranchRole(
		r.ProjectID, r.BranchID, neon.RoleCreateRequest{
			Role: neon.RoleCreateRequestRole{
				Name: r.Name,
			},
		},
	)
	if err != nil {
		return err
	}

	d.SetId(r.toString())

	if err := d.Set("password", resp.Role.Password); err != nil {
		return err
	}

	return updateStateRole(d, resp.Role)
}

func resourceRoleReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceRoleRead, ctx, d, meta)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "read Role")

	resp, err := meta.(neon.Client).GetProjectBranchRole(
		d.Get("project_id").(string), d.Get("branch_id").(string), d.Get("name").(string),
	)
	if err != nil {
		return err
	}

	return updateStateRole(d, resp.Role)
}

func resourceRoleDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceRoleDelete, ctx, d, meta)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Role")
	if _, err := meta.(neon.Client).DeleteProjectBranchRole(
		d.Get("project_id").(string),
		d.Get("branch_id").(string),
		d.Get("name").(string),
	); err != nil {
		return err
	}
	d.SetId("")
	return updateStateRole(d, neon.Role{})
}

func resourceRoleImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	tflog.Trace(ctx, "import Role")

	r, err := parseRoleID(d.Id())
	if err != nil {
		return nil, err
	}

	resp, err := meta.(neon.Client).ResetProjectBranchRolePassword(
		r.ProjectID,
		r.BranchID,
		r.Name,
	)
	if err != nil {
		return nil, err
	}

	if err := updateStateRole(d, resp.Role); err != nil {
		return nil, err
	}
	setResourceDataFrom(d, r)

	return []*schema.ResourceData{d}, nil
}
