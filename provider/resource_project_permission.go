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

func resourceProjectPermission() *schema.Resource {
	return &schema.Resource{
		SchemaVersion: 1,
		Description:   `Project's access permission.`,
		Importer: &schema.ResourceImporter{
			StateContext: resourceProjectPermissionImport,
		},
		CreateContext: resourceProjectPermissionCreateRetry,
		ReadContext:   resourceProjectPermissionReadRetry,
		DeleteContext: resourceProjectPermissionDeleteRetry,
		Schema: map[string]*schema.Schema{
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"project_id": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Project ID.",
			},
			"grantee": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "Email of the user whom to grant project permission.",
			},
		},
	}
}

func resourceProjectPermissionCreateRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceProjectPermissionCreate, ctx, d, meta)
}

func resourceProjectPermissionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	projectID := d.Get("project_id").(string)
	email := d.Get("grantee").(string)

	tflog.Trace(ctx, "grant project permission", map[string]interface{}{"projectID": projectID, "email": email})

	resp, err := meta.(sdkProject).GrantPermissionToProject(projectID, neon.GrantPermissionToProjectRequest{Email: email})
	if err != nil {
		return err
	}

	id := joinedIDProjectPermission{
		projectID:    projectID,
		permissionID: resp.ID,
	}
	d.SetId(id.ToString())

	return nil
}

func resourceProjectPermissionDeleteRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceProjectPermissionDelete, ctx, d, meta)
}

func resourceProjectPermissionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	joinedID, _ := parseJoinedIDProjectPermission(d.Id())

	tflog.Trace(ctx, "revoke project permission", map[string]interface{}{
		"projectID":    joinedID.projectID,
		"permissionID": joinedID.permissionID,
	})

	if _, err := meta.(sdkProject).RevokePermissionFromProject(joinedID.projectID, joinedID.permissionID); err != nil {
		return err
	}

	d.SetId("")
	return nil
}

func resourceProjectPermissionImport(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	tflog.Trace(ctx, "import project permission", map[string]interface{}{"id": d.Id()})

	var found bool
	diags := projectReadiness.Retry(
		func(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
			return func() error {
				var err error
				err, found = readProjectPermission(ctx, d, meta)
				return err
			}()
		},
		ctx, d, meta,
	)
	if diags.HasError() {
		var errs = make([]error, len(diags))
		for i, di := range diags {
			errs[i] = errors.New(di.Summary)
		}
		return nil, errors.Join(errs...)
	}

	if !found {
		return nil, errors.New("no permission found")
	}

	return []*schema.ResourceData{d}, nil
}

func resourceProjectPermissionReadRetry(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	return projectReadiness.Retry(resourceProjectPermissionRead, ctx, d, meta)
}

func resourceProjectPermissionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) error {
	err, found := readProjectPermission(ctx, d, meta)
	if err != nil {
		return err
	}

	if !found {
		tflog.Trace(ctx, "no project permission found")
	}

	return nil
}

func readProjectPermission(ctx context.Context, d *schema.ResourceData, meta interface{}) (error, bool) {
	tflog.Trace(ctx, "parse project permission found", map[string]interface{}{"id": d.Id()})

	joinedID, err := parseJoinedIDProjectPermission(d.Id())
	if err != nil {
		return err, false
	}

	projectID := joinedID.projectID
	tflog.Trace(ctx, "list project permissions", map[string]interface{}{"projectID": projectID})

	resp, err := meta.(sdkProject).ListProjectPermissions(projectID)
	if err != nil {
		return err, false
	}

	tflog.Trace(ctx, "search project permission", map[string]interface{}{
		"projectID":    projectID,
		"permissionID": joinedID.permissionID,
	})

	for _, permission := range resp.ProjectPermissions {
		if permission.ID == joinedID.permissionID {
			if err := d.Set("project_id", projectID); err != nil {
				return err, false
			}
			if err := d.Set("grantee", permission.GrantedToEmail); err != nil {
				return err, false
			}
			return nil, true
		}
	}
	return nil, false
}

type joinedIDProjectPermission struct {
	projectID, permissionID string
}

const joinedIDProjectPermissionSeparator = "/"

func (v joinedIDProjectPermission) ToString() string {
	return v.projectID + joinedIDProjectPermissionSeparator + v.permissionID
}

func parseJoinedIDProjectPermission(s string) (*joinedIDProjectPermission, error) {
	els := strings.SplitN(s, joinedIDProjectPermissionSeparator, 2)

	if len(els) != 2 {
		return nil, errors.New("not recognized format of the project permission resource's ID")
	}

	return &joinedIDProjectPermission{
		projectID:    els[0],
		permissionID: els[1],
	}, nil
}
