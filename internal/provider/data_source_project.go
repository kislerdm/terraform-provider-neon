package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	neon "github.com/kislerdm/neon-sdk-go"
)

func dataSourceProject() *schema.Resource {
	return &schema.Resource{
		Description:   `Fetch Project.`,
		SchemaVersion: 1,
		ReadContext:   dataSourceProjectRead,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:        schema.TypeString,
				Description: "Project ID.",
				Required:    true,
			},
			"name": {
				Type:        schema.TypeString,
				Description: "Project Name.",
				Computed:    true,
			},
			"default_branch_id": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default branch ID.",
			},
			"database_host": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default database host.",
			},
			"database_name": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default database name.",
			},
			"database_user": {
				Type:        schema.TypeString,
				Computed:    true,
				Description: "Default database role.",
			},
			"database_password": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Default database access password.",
			},
			"connection_uri": {
				Type:        schema.TypeString,
				Computed:    true,
				Sensitive:   true,
				Description: "Default connection uri. **Note** that it contains access credentials.",
			},
		},
	}
}

func dataSourceProjectRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	tflog.Trace(ctx, "get Project")

	client := meta.(*neon.Client)

	resp, err := client.GetProject(d.Get("id").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	project := resp.Project

	d.SetId(project.ID)
	if err := d.Set("name", project.Name); err != nil {
		return diag.FromErr(err)
	}

	branches, err := client.ListProjectBranches(project.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	var defaultBranch neon.Branch
	for _, v := range branches.Branches {
		if v.Primary {
			defaultBranch = v
			break
		}
	}

	if err := d.Set("default_branch_id", defaultBranch.ID); err != nil {
		return diag.FromErr(err)
	}

	endpoints, err := client.ListProjectBranchEndpoints(project.ID, defaultBranch.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	databases, err := client.ListProjectBranchDatabases(project.ID, defaultBranch.ID)
	if err != nil {
		return diag.FromErr(err)
	}

	info, err := newDbConnectionInfo(client, project.ID, defaultBranch.ID, endpoints.Endpoints,
		databases.Databases)
	if err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_host", info.host); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", info.dbName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_user", info.userName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_password", info.pass); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("connection_uri", info.connectionURI()); err != nil {
		return diag.FromErr(err)
	}

	return diag.FromErr(nil)
}
