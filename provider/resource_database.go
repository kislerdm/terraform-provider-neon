package provider

import (
	"context"
	"errors"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func resourceDatabase() *schema.Resource {
	return &schema.Resource{
		Description:   `Project Database. See details: https://neon.tech/docs/manage/databases/`,
		SchemaVersion: 7,
		Importer: &schema.ResourceImporter{
			StateContext: resourceDatabaseImport,
		},
		CreateContext: resourceDatabaseCreateRetry,
		ReadContext:   resourceDatabaseReadRetry,
		UpdateContext: resourceDatabaseUpdateRetry,
		DeleteContext: resourceDatabaseDeleteRetry,
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
				Description: "Database name.",
			},
			"owner_name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Role name of the database owner.",
			},
		},
	}
}

func updateStateDatabase(d *schema.ResourceData, v neon.Database) error {
	if err := d.Set("owner_name", v.OwnerName); err != nil {
		return err
	}
	return nil
}

func resourceDatabaseCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceDatabaseCreate, ctx, d, meta)
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "created Database")

	r := complexID{
		ProjectID: d.Get("project_id").(string),
		BranchID:  d.Get("branch_id").(string),
		Name:      d.Get("name").(string),
	}
	resp, err := meta.(*neon.Client).CreateProjectBranchDatabase(
		r.ProjectID, r.BranchID, neon.DatabaseCreateRequest{
			Database: neon.DatabaseCreateRequestDatabase{
				Name:      r.Name,
				OwnerName: d.Get("owner_name").(string),
			},
		},
	)
	if err != nil {
		return err
	}

	d.SetId(r.toString())

	return updateStateDatabase(d, resp.DatabaseResponse.Database)
}

func resourceDatabaseReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceDatabaseRead, ctx, d, meta)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "read Database")

	resp, err := meta.(*neon.Client).GetProjectBranchDatabase(
		d.Get("project_id").(string), d.Get("branch_id").(string), d.Get("name").(string),
	)
	if err != nil {
		return err
	}

	return updateStateDatabase(d, resp.Database)
}

func resourceDatabaseUpdateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceDatabaseUpdate, ctx, d, meta)
}

func resourceDatabaseUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "update Database")

	r, err := parseComplexID(d.Id())
	if err != nil {
		panic(err)
	}

	resp, err := meta.(*neon.Client).UpdateProjectBranchDatabase(
		r.ProjectID, r.BranchID, r.Name,
		neon.DatabaseUpdateRequest{
			Database: neon.DatabaseUpdateRequestDatabase{
				Name:      pointer(d.Get("name").(string)),
				OwnerName: pointer(d.Get("owner_name").(string)),
			},
		},
	)
	if err != nil {
		return err
	}

	r.Name = resp.DatabaseResponse.Database.Name
	d.SetId(r.toString())
	return updateStateDatabase(d, resp.Database)
}

func resourceDatabaseDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceDatabaseDelete, ctx, d, meta)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Database")
	if _, err := meta.(*neon.Client).DeleteProjectBranchDatabase(
		d.Get("project_id").(string),
		d.Get("branch_id").(string),
		d.Get("name").(string),
	); err != nil {
		return err
	}
	d.SetId("")
	return updateStateDatabase(d, neon.Database{})
}

func resourceDatabaseImport(ctx context.Context, d *schema.ResourceData, meta interface{}) (
	[]*schema.ResourceData, error,
) {
	tflog.Trace(ctx, "import Database")

	r, err := parseComplexID(d.Id())
	if err != nil {
		return nil, err
	}

	setResourceDataFromComplexID(d, r)
	if err := resourceDatabaseRead(ctx, d, meta); err != nil {
		return nil, err
	}

	return []*schema.ResourceData{d}, nil
}

type complexID struct {
	ProjectID, BranchID, Name string
}

func setResourceDataFromComplexID(d *schema.ResourceData, r complexID) {
	_ = d.Set("project_id", r.ProjectID)
	_ = d.Set("branch_id", r.BranchID)
	_ = d.Set("name", r.Name)
}

func (v complexID) toString() string {
	return v.ProjectID + "/" + v.BranchID + "/" + v.Name
}

func parseComplexID(s string) (complexID, error) {
	spl := strings.Split(s, "/")
	if len(spl) != 3 {
		return complexID{}, errors.New(
			"ID of this resource type shall follow the template: {{.ProjectID}}/{{.BranchID}}/{{.Name}}",
		)
	}
	return complexID{
		ProjectID: spl[0],
		BranchID:  spl[1],
		Name:      spl[2],
	}, nil
}
