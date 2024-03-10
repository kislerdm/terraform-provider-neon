package provider

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	neon "github.com/kislerdm/neon-sdk-go"
)

func Test_resourceProjectPermissionCreate(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	t.Parallel()

	t.Run("shall create permission for foo@bar.baz", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()

		const (
			email     = "foo@bar.baz"
			projectID = "myproject"
		)

		if err := definition.Set("grantee", email); err != nil {
			t.Fatal(err)
		}
		if err := definition.Set("project_id", projectID); err != nil {
			t.Fatal(err)
		}

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				ProjectPermissions: neon.ProjectPermissions{
					ProjectPermissions: []neon.ProjectPermission{},
				},
			},
		}

		if err := resourceProjectPermissionCreate(context.TODO(), definition, meta); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gotEmail := meta.ProjectPermissions.ProjectPermissions[0].GrantedToEmail
		if gotEmail != email {
			t.Fatal("email is not set")
		}

		gotPermissionID := meta.ProjectPermissions.ProjectPermissions[0].ID

		wantID := projectID + "/" + gotPermissionID
		if definition.Id() != wantID {
			t.Errorf("unexpected resource ID: want=%s, got=%s", wantID, definition.Id())
		}
	})

	t.Run("unhappy path", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()

		const (
			email     = "foo@bar.baz"
			projectID = "myproject"
		)

		if err := definition.Set("grantee", email); err != nil {
			t.Fatal(err)
		}
		if err := definition.Set("project_id", projectID); err != nil {
			t.Fatal(err)
		}

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				err: errors.New("foobar"),
			},
		}

		if err := resourceProjectPermissionCreate(context.TODO(), definition, meta); err == nil {
			t.Fatalf("error expected")
		}

		if definition.Id() != "" {
			t.Errorf("unexpected resource ID: want=%s, got=%s", "", definition.Id())
		}
	})
}

func Test_resourceProjectPermissionDelete(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	t.Parallel()

	t.Run("shall revoke existing permission", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()

		const (
			email        = "foo@bar.baz"
			projectID    = "myproject"
			permissionID = "mypermission"
		)

		if err := definition.Set("grantee", email); err != nil {
			t.Fatal(err)
		}
		if err := definition.Set("project_id", projectID); err != nil {
			t.Fatal(err)
		}

		id := projectID + "/" + permissionID
		definition.SetId(id)

		meta := &sdkClientStub{}
		if err := resourceProjectPermissionDelete(context.TODO(), definition, meta); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if definition.Id() != "" {
			t.Errorf("unexpected resource ID: want=%s, got=%s", "", definition.Id())
		}
	})

	t.Run("unhappy path", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()

		const (
			email        = "foo@bar.baz"
			projectID    = "myproject"
			permissionID = "mypermission"
		)

		if err := definition.Set("grantee", email); err != nil {
			t.Fatal(err)
		}
		if err := definition.Set("project_id", projectID); err != nil {
			t.Fatal(err)
		}
		id := projectID + "/" + permissionID
		definition.SetId(id)

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				err: errors.New("foobar"),
			},
		}

		if err := resourceProjectPermissionCreate(context.TODO(), definition, meta); err == nil {
			t.Fatalf("error expected")
		}

		if definition.Id() != id {
			t.Errorf("unexpected resource ID: want=%s, got=%s", id, definition.Id())
		}
	})
}

func Test_resourceProjectPermissionRead(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	t.Parallel()
	const (
		projectID    = "myproject"
		permissionID = "mypermission"
		id           = projectID + "/" + permissionID
	)

	t.Run("shall find the permission given the resource id", func(t *testing.T) {
		const email = "foo@bar.baz"
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()
		definition.SetId(id)

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				ProjectPermissions: neon.ProjectPermissions{
					ProjectPermissions: []neon.ProjectPermission{
						{
							GrantedAt:      time.Now().UTC(),
							GrantedToEmail: email,
							ID:             permissionID,
						},
					},
				},
			},
		}

		if err := resourceProjectPermissionRead(context.TODO(), definition, meta); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gotEmail := definition.Get("grantee").(string)
		if gotEmail != email {
			t.Fatalf("unexpected grantee email found: want=%s, got=%s", email, gotEmail)
		}

		gotProjectID := definition.Get("project_id").(string)
		if gotProjectID != projectID {
			t.Fatalf("unexpected projectID found: want=%s, got=%s", projectID, gotProjectID)
		}
	})

	t.Run("shall find no permission by its id", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()
		definition.SetId(id)

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				ProjectPermissions: neon.ProjectPermissions{},
			},
		}

		if err := resourceProjectPermissionRead(context.TODO(), definition, meta); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		gotEmail := definition.Get("grantee").(string)
		if gotEmail != "" {
			t.Fatalf("unexpected grantee email found: want=%s, got=%s", "", gotEmail)
		}

		gotProjectID := definition.Get("project_id").(string)
		if gotProjectID != "" {
			t.Fatalf("unexpected projectID found: want=%s, got=%s", "", gotProjectID)
		}
	})

	t.Run("shall fail when listing project's permissions", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()
		definition.SetId(id)

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				err: errors.New("foobar"),
			},
		}

		if err := resourceProjectPermissionRead(context.TODO(), definition, meta); err == nil {
			t.Fatal("error expected")
		}
	})
}

func Test_resourceProjectPermissionImport(t *testing.T) {
	if os.Getenv("TF_ACC") == "1" {
		t.Skip("acceptance tests are running")
	}

	t.Parallel()

	const (
		projectID    = "foo"
		permissionID = "bar"
		id           = projectID + "/" + permissionID
		email        = "foo@bar.baz"
	)

	t.Run("shall import the permission given its id", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()
		definition.SetId(id)

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				ProjectPermissions: neon.ProjectPermissions{
					ProjectPermissions: []neon.ProjectPermission{
						{
							GrantedAt:      time.Now().UTC(),
							GrantedToEmail: email,
							ID:             permissionID,
						},
					},
				},
			},
		}

		resources, err := resourceProjectPermissionImport(context.TODO(), definition, meta)
		if err != nil {
			t.Fatalf("unexpected errors: %v", err)
		}

		d := resources[0]

		gotEmail := d.Get("grantee").(string)
		if gotEmail != email {
			t.Fatalf("unexpected grantee email found: want=%s, got=%s", email, gotEmail)
		}

		gotProjectID := d.Get("project_id").(string)
		if gotProjectID != projectID {
			t.Fatalf("unexpected projectID found: want=%s, got=%s", projectID, gotProjectID)
		}
	})

	t.Run("shall fail because no permission was found by its id", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()
		definition.SetId(id)

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				ProjectPermissions: neon.ProjectPermissions{},
			},
		}

		_, err := resourceProjectPermissionImport(context.TODO(), definition, meta)
		const wantErrStr = "no permission found"
		if err.Error() != wantErrStr {
			t.Fatalf("'%s' error expected", wantErrStr)
		}
	})

	t.Run("shall fail because provided id is not correct", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()
		definition.SetId("qux")

		meta := &sdkClientStub{}
		_, err := resourceProjectPermissionImport(context.TODO(), definition, meta)

		const wantErrStr = "not recognized format of the project permission resource's ID"
		if err.Error() != wantErrStr {
			t.Fatalf("'%s' error expected, got: %s", wantErrStr, err.Error())
		}
	})

	t.Run("shall fail when listing project's permissions", func(t *testing.T) {
		resource := resourceProjectPermission()
		definition := resource.TestResourceData()
		definition.SetId(id)

		meta := &sdkClientStub{
			stubProjectPermission: stubProjectPermission{
				err: errors.New("foobar"),
			},
		}

		_, err := resourceProjectPermissionImport(context.TODO(), definition, meta)
		if err == nil {
			t.Fatal("error expected")
		}
	})
}
