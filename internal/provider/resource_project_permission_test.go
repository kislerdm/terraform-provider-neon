package provider

import (
	"context"
	"errors"
	"testing"

	neon "github.com/kislerdm/neon-sdk-go"
)

func Test_resourceProjectPermissionCreate(t *testing.T) {
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

		d := resourceProjectPermissionCreate(context.TODO(), definition, meta)
		if d.HasError() {
			t.Fatalf("unexpected errors: %v", d)
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

		d := resourceProjectPermissionCreate(context.TODO(), definition, meta)
		if !d.HasError() {
			t.Fatalf("error expected")
		}

		if definition.Id() != "" {
			t.Errorf("unexpected resource ID: want=%s, got=%s", "", definition.Id())
		}
	})
}

func Test_resourceProjectPermissionDelete(t *testing.T) {
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
		d := resourceProjectPermissionDelete(context.TODO(), definition, meta)
		if d.HasError() {
			t.Fatalf("unexpected errors: %v", d)
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

		d := resourceProjectPermissionCreate(context.TODO(), definition, meta)
		if !d.HasError() {
			t.Fatalf("error expected")
		}

		if definition.Id() != id {
			t.Errorf("unexpected resource ID: want=%s, got=%s", id, definition.Id())
		}
	})
}

func Test_resourceProjectPermissionRead(t *testing.T) {

}
