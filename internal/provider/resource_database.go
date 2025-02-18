package provider

import (
	"context"
	"errors"

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
	client := meta.(*neon.Client)
	resp, err := client.CreateProjectBranchDatabase(
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
	waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)

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

	client := meta.(*neon.Client)
	resp, err := client.UpdateProjectBranchDatabase(
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
	waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)
	r.Name = resp.DatabaseResponse.Database.Name
	d.SetId(r.toString())
	return updateStateDatabase(d, resp.Database)
}

func resourceDatabaseDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceDatabaseDelete, ctx, d, meta)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	tflog.Trace(ctx, "delete Database")
	client := meta.(*neon.Client)
	resp, err := client.DeleteProjectBranchDatabase(
		d.Get("project_id").(string),
		d.Get("branch_id").(string),
		d.Get("name").(string),
	)
	if err != nil {
		return err
	}
	waitUnfinishedOperations(ctx, client, resp.OperationsResponse.Operations)
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

	setResourceAttrsFromComplexID(d, r)
	if diags := resourceDatabaseReadRetry(ctx, d, meta); diags.HasError() {
		return nil, errors.New(diags[0].Summary)
	}

	return []*schema.ResourceData{d}, nil
}
