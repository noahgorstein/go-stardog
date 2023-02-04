package stardog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestRoleService_ListNames(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolesJSON = []byte(`{
  "roles": ["reader", "writer"] 
  }`)
	var wantRoles = []string{"reader", "writer"}

	mux.HandleFunc("/admin/roles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(rolesJSON)
	})

	ctx := context.Background()
	got, _, err := client.Role.ListNames(ctx)
	if err != nil {
		t.Errorf("Role.ListNames returned error: %v", err)
	}
	if want := wantRoles; !cmp.Equal(got, want) {
		t.Errorf("Role.ListNames = %+v, want %+v", got, want)
	}

	const methodName = "ListNames"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Role.ListNames(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestRoleService_List(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	rolesJSON := []byte(
		`
    {
      "roles": [
        {
          "rolename": "reader",
          "permissions": [
            {
              "action": "READ",
              "resource_type": "*",
              "resource": [
                "*"
              ]
            }
          ]
        },
        {
          "rolename": "writer",
          "permissions": [
            {
              "action": "WRITE",
              "resource_type": "*",
              "resource": [
                "*"
              ]
            }
          ]
        }
      ]
  }`)
	wantRoles := &listRolesResponse{
		Roles: []Role{
			{
				Name: "reader",
				Permissions: []Permission{
					{
						Action:       PermissionActionRead,
						ResourceType: PermissionResourceTypeAll,
						Resource:     []string{"*"},
					},
				},
			},
			{
				Name: "writer",
				Permissions: []Permission{
					{
						Action:       PermissionActionWrite,
						ResourceType: PermissionResourceTypeAll,
						Resource:     []string{"*"},
					},
				},
			},
		},
	}

	mux.HandleFunc("/admin/roles/list", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(rolesJSON)
	})

	ctx := context.Background()
	got, _, err := client.Role.List(ctx)
	if err != nil {
		t.Errorf("Role.List returned error: %v", err)
	}
	if want := wantRoles.Roles; !cmp.Equal(got, want) {
		t.Errorf("Role.List = %+v, want %+v", got, want)
	}

	const methodName = "List"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Role.List(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestRoleService_Create(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	const rolename = "reader"

	mux.HandleFunc("/admin/roles", func(w http.ResponseWriter, r *http.Request) {
		v := new(createRoleRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", mediaTypeApplicationJSON)

		want := &createRoleRequest{Rolename: rolename}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusCreated)
	})

	ctx := context.Background()
	_, err := client.Role.Create(ctx, rolename)
	if err != nil {
		t.Errorf("Role.Create returned error: %v", err)
	}
	const methodName = "Create"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Role.Create(nil, rolename)
	})
}

func TestRoleService_Permissions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolename = "reader"

	var rolePermissionsJSON = `{
    "permissions": [
      {"action":"READ","resource_type":"named-graph","resource":["db1"]}
      ]
    }`
	var wantRolePermissions = []Permission{
		{
			Action:       PermissionActionRead,
			ResourceType: PermissionResourceTypeNamedGraph,
			Resource:     []string{"db1"},
		}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/role/%s", rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(rolePermissionsJSON))
	})

	ctx := context.Background()
	got, _, err := client.Role.Permissions(ctx, rolename)
	if err != nil {
		t.Errorf("Role.Permissions returned error: %v", err)
	}
	if want := wantRolePermissions; !cmp.Equal(got, want) {
		t.Errorf("Role.Permissions = %+v, want %+v", got, want)
	}

	const methodName = "Permissions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Role.Permissions(nil, "somerole")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestRoleService_GrantPermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolename = "reader"
	var permission = &Permission{
		Action:       PermissionActionRead,
		ResourceType: PermissionResourceTypeDatabase,
		Resource:     []string{"*"},
	}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/role/%s", rolename), func(w http.ResponseWriter, r *http.Request) {
		v := new(Permission)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", "application/json")

		want := permission
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusCreated)
	})

	ctx := context.Background()
	_, err := client.Role.GrantPermission(ctx, rolename, *permission)
	if err != nil {
		t.Errorf("Role.GrantPermission returned error: %v", err)
	}

	const methodName = "GrantPermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Role.GrantPermission(nil, rolename, *permission)
	})
}

func TestRoleService_RevokePermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolename = "reader"
	var permission = &Permission{
		Action:       PermissionActionRead,
		ResourceType: PermissionResourceTypeDatabase,
		Resource:     []string{"*"}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/role/%s/delete", rolename), func(w http.ResponseWriter, r *http.Request) {
		v := new(Permission)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", "application/json")

		want := permission
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusCreated)
	})

	ctx := context.Background()
	_, err := client.Role.RevokePermission(ctx, rolename, *permission)
	if err != nil {
		t.Errorf("Role.RevokePermission returned error: %v", err)
	}

	const methodName = "RevokePermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Role.RevokePermission(nil, rolename, *permission)
	})
}

func TestRoleService_Delete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	rolename := "reader"
	opt := &DeleteRoleOptions{Force: true}

	mux.HandleFunc(fmt.Sprintf("/admin/roles/%s", rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")

		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.Role.Delete(ctx, rolename, opt)
	if err != nil {
		t.Errorf("Role.Delete returned error: %v", err)
	}
	const methodName = "Delete"
	testBadOptions(t, methodName, func() (err error) {
		_, err = client.Role.Delete(ctx, "\n", &DeleteRoleOptions{})
		return err
	})

	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Role.Delete(nil, "writer", opt)
	})
}
