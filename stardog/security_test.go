package stardog

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func newTrue() *bool {
	b := true
	return &b
}

func newFalse() *bool {
	b := false
	return &b
}

func Test_GetUsers(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var usersJSON = []byte(`{
  "users": ["admin", "charlie"] 
  }`)
	var wantUsers = []string{"admin", "charlie"}

	mux.HandleFunc("/admin/users", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(usersJSON)
	})

	ctx := context.Background()
	got, _, err := client.Security.GetUsers(ctx)
	if err != nil {
		t.Errorf("Security.GetUsers returned error: %v", err)
	}
	if want := wantUsers; !cmp.Equal(got, want) {
		t.Errorf("Security.GetUsers = %+v, want %+v", got, want)
	}

	const methodName = "GetUsers"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetUsers(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetUserPermissions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var userPermissionsJSON = `{
    "permissions": [
      {"action":"READ","resource_type":"named-graph","resource":["db1"]}
      ]
    }`
	var wantUserPermissions = &[]Permission{
		{Action: "READ", ResourceType: "named-graph", Resource: []string{"db1"}}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/user/%s", "bob"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userPermissionsJSON))
	})

	ctx := context.Background()
	got, _, err := client.Security.GetUserPermissions(ctx, "bob")
	if err != nil {
		t.Errorf("Security.GetUserPermissions returned error: %v", err)
	}
	if want := wantUserPermissions; !cmp.Equal(got, want) {
		t.Errorf("Security.UserPermissions = %+v, want %+v", got, want)
	}

	const methodName = "GetUserPermissions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetUserPermissions(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetUserEffectivePermissions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var userEffectivePermissionsJSON = `{
    "permissions": [
      {"action":"DELETE","resource_type":"named-graph","resource":["db1"], "explicit": false}
      ]
    }`
	var wantUserEffectivePermissions = &[]Permission{
		{Action: "DELETE", ResourceType: "named-graph", Resource: []string{"db1"}, Explicit: newFalse()}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/effective/user/%s", "bob"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userEffectivePermissionsJSON))
	})

	ctx := context.Background()
	got, _, err := client.Security.GetUserEffectivePermissions(ctx, "bob")
	if err != nil {
		t.Errorf("Security.GetUserEffectivePermissions returned error: %v", err)
	}
	if want := wantUserEffectivePermissions; !cmp.Equal(got, want) {
		t.Errorf("Security.GetUserEffectivePermissions = %+v, want %+v", got, want)
	}

	const methodName = "GetUserEffectivePermissions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetUserEffectivePermissions(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetUserDetails(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var userDetailsJSON = `{
    "enabled": true,
    "superuser": false,
    "roles": [],
    "permissions": [
      {
        "action": "READ",
        "resource_type": "db",
        "resource": [
          "myDatabase"
        ],
        "explicit": true
      },
      {
        "action": "READ",
        "resource_type": "user",
        "resource": [
          "frodo"
        ],
        "explicit": true
      },
      {
        "action": "WRITE",
        "resource_type": "user",
        "resource": [
          "frodo"
        ],
        "explicit": true
      }
    ]
  }
  `
	var wantUserDetails = &UserDetails{
		Enabled:   true,
		Superuser: false,
		Roles:     []string{},
		Permissions: []Permission{
			{Action: "READ", ResourceType: "db", Resource: []string{"myDatabase"}, Explicit: newTrue()},
			{Action: "READ", ResourceType: "user", Resource: []string{"frodo"}, Explicit: newTrue()},
			{Action: "WRITE", ResourceType: "user", Resource: []string{"frodo"}, Explicit: newTrue()},
		},
	}
	mux.HandleFunc(fmt.Sprintf("/admin/users/%s", "bob"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(userDetailsJSON))
	})

	ctx := context.Background()
	got, _, err := client.Security.GetUserDetails(ctx, "bob")
	if err != nil {
		t.Errorf("Security.GetUserDetails returned error: %v", err)
	}
	if want := wantUserDetails; !cmp.Equal(got, want) {
		t.Errorf("Security.GetUserDetails = %+v, want %+v", got, want)
	}

	const methodName = "GetUserDetails"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetUserDetails(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_IsSuperuser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var isSuperuserJson = `{"superuser": false}`
	var isSuperuser = newFalse()

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/superuser", "bob"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(isSuperuserJson))
	})

	ctx := context.Background()
	got, _, err := client.Security.IsSuperuser(ctx, "bob")
	if err != nil {
		t.Errorf("Security.IsSuperuser returned error: %v", err)
	}
	if want := isSuperuser; !cmp.Equal(got, want) {
		t.Errorf("Security.IsSuperuser = %+v, want %+v", *got, *want)
	}

	const methodName = "IsSuperuser"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.IsSuperuser(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, *got)
		}
		return resp, err
	})
}

func Test_IsEnabled(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var isEnabledJson = `{"enabled": false}`
	var isEnabled = newFalse()

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/enabled", "bob"), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(isEnabledJson))
	})

	ctx := context.Background()
	got, _, err := client.Security.IsEnabled(ctx, "bob")
	if err != nil {
		t.Errorf("Security.IsEnabled returned error: %v", err)
	}
	if want := isEnabled; !cmp.Equal(got, want) {
		t.Errorf("Security.IsEnabled = %+v, want %+v", *got, *want)
	}

	const methodName = "IsEnabled"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.IsEnabled(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, *got)
		}
		return resp, err
	})
}

func Test_CreateUser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	const username = "frodo"
	var password = strings.Split("gandalf", "")

	mux.HandleFunc("/admin/users", func(w http.ResponseWriter, r *http.Request) {
		v := new(credentials)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", "application/json")

		want := &credentials{Username: username, Password: password}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusCreated)
	})

	ctx := context.Background()
	_, err := client.Security.CreateUser(ctx, username, strings.Join(password, ""))
	if err != nil {
		t.Errorf("Security.CreateUser returned error: %v", err)
	}

	const methodName = "CreateUser"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.CreateUser(nil, "someone", "password")
	})
}

func Test_DeleteUser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	username := "bob"
	mux.HandleFunc(fmt.Sprintf("/admin/users/%s", username), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")

		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.Security.DeleteUser(ctx, username)
	if err != nil {
		t.Errorf("Security.DeleteUser returned error: %v", err)
	}
	const methodName = "DeleteUser"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.DeleteUser(nil, "someone")
	})
}

func Test_ChangeUserPassword(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"
	var password = "somePassword"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/pwd", username), func(w http.ResponseWriter, r *http.Request) {
		v := new(changePasswordRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", "application/json")

		want := &changePasswordRequest{Password: password}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.Security.ChangeUserPassword(ctx, username, password)
	if err != nil {
		t.Errorf("Security.ChangeUserPassword returned error: %v", err)
	}
	const methodName = "ChangeUserPassword"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.ChangeUserPassword(nil, "someone", "password")
	})
}

func Test_EnableUser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/enabled", username), func(w http.ResponseWriter, r *http.Request) {
		v := new(enableRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", "application/json")

		want := &enableRequest{Enabled: false}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.Security.EnableUser(ctx, username, false)
	if err != nil {
		t.Errorf("Security.SetEnabled returned error: %v", err)
	}

	const methodName = "SetEnabled"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.EnableUser(nil, "someone", false)
	})
}

func Test_GetRoles(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var usersJSON = []byte(`{
  "roles": ["reader", "writer"] 
  }`)
	var wantRoles = []string{"reader", "writer"}

	mux.HandleFunc("/admin/roles", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(usersJSON)
	})

	ctx := context.Background()
	got, _, err := client.Security.GetRoles(ctx)
	if err != nil {
		t.Errorf("Security.GetRoles returned error: %v", err)
	}
	if want := wantRoles; !cmp.Equal(got, want) {
		t.Errorf("Security.GetRoles = %+v, want %+v", got, want)
	}

	const methodName = "GetRoles"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetRoles(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetRolesAssignedToUser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"
	var getRolesJSON = []byte(`{
  "roles": ["reader", "writer"] 
  }`)
	var wantGetRoles = []string{"reader", "writer"}

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/roles", username), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(getRolesJSON)
	})

	ctx := context.Background()
	got, _, err := client.Security.GetRolesAssignedToUser(ctx, username)
	if err != nil {
		t.Errorf("Security.GetRolesAssignedToUser returned error: %v", err)
	}
	if want := wantGetRoles; !cmp.Equal(got, want) {
		t.Errorf("Security.GetRolesAssignedToUser = %+v, want %+v", got, want)
	}

	const methodName = "GetRolesAssignedToUser"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetRolesAssignedToUser(nil, username)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_CreateRole(t *testing.T) {
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
	_, err := client.Security.CreateRole(ctx, rolename)
	if err != nil {
		t.Errorf("Security.CreateRole returned error: %v", err)
	}
	const methodName = "CreateRole"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.CreateRole(nil, rolename)
	})
}

func Test_GetRolePermissions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolename = "reader"

	var rolePermissionsJSON = `{
    "permissions": [
      {"action":"READ","resource_type":"named-graph","resource":["db1"]}
      ]
    }`
	var wantRolePermissions = &[]Permission{
		{Action: "READ", ResourceType: "named-graph", Resource: []string{"db1"}}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/role/%s", rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(rolePermissionsJSON))
	})

	ctx := context.Background()
	got, _, err := client.Security.GetRolePermissions(ctx, rolename)
	if err != nil {
		t.Errorf("Security.GetRolePermissions returned error: %v", err)
	}
	if want := wantRolePermissions; !cmp.Equal(got, want) {
		t.Errorf("Security.GetRolePermissions = %+v, want %+v", got, want)
	}

	const methodName = "GetRolePermissions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetRolePermissions(nil, "somerole")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GetRolePermissions_writeResultsToBuffer(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolename = "reader"

	var rolePermissionsJSON = `{
    "permissions": [
      {"action":"READ","resource_type":"named-graph","resource":["db1"]}
      ]
    }`
	var wantRolePermissions = &[]Permission{
		{Action: "READ", ResourceType: "named-graph", Resource: []string{"db1"}}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/role/%s", rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(rolePermissionsJSON))
	})

	ctx := context.Background()
	got, _, err := client.Security.GetRolePermissions(ctx, rolename)
	if err != nil {
		t.Errorf("Security.GetRolePermissions returned error: %v", err)
	}
	if want := wantRolePermissions; !cmp.Equal(got, want) {
		t.Errorf("Security.GetRolePermissions = %+v, want %+v", got, want)
	}

	const methodName = "GetRolePermissions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetRolePermissions(nil, "somerole")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GrantRolePermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolename = "reader"
	var permission = &Permission{Action: "read", ResourceType: "db", Resource: []string{"*"}}

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
	_, err := client.Security.GrantRolePermission(ctx, rolename, *permission)
	if err != nil {
		t.Errorf("Security.GrantRolePermission returned error: %v", err)
	}

	const methodName = "GrantRolePermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.GrantRolePermission(nil, rolename, *permission)
	})
}

func Test_RevokeRolePermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var rolename = "reader"
	var permission = &Permission{Action: "read", ResourceType: "db", Resource: []string{"*"}}

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
	_, err := client.Security.RevokeRolePermission(ctx, rolename, *permission)
	if err != nil {
		t.Errorf("Security.RevokeRolePermission returned error: %v", err)
	}

	const methodName = "RevokeRolePermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.RevokeRolePermission(nil, rolename, *permission)
	})
}

func Test_DeleteRole(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	rolename := "reader"
	opt := &DeleteRoleOptions{Force: true}

	mux.HandleFunc(fmt.Sprintf("/admin/roles/%s", rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")

		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.Security.DeleteRole(ctx, rolename, opt)
	if err != nil {
		t.Errorf("Security.DeleteRole returned error: %v", err)
	}
	const methodName = "DeleteRole"
	testBadOptions(t, methodName, func() (err error) {
		_, err = client.Security.DeleteRole(ctx, "\n", &DeleteRoleOptions{})
		return err
	})

	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.DeleteRole(nil, "writer", opt)
	})
}

func Test_GetUsersAssignedRole(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	rolename := "reader"
	usersJSON := []byte(`{
  "users": ["admin", "charlie"] 
  }`)
	wantUsers := []string{"admin", "charlie"}

	mux.HandleFunc(fmt.Sprintf("/admin/roles/%s/users", rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(usersJSON)
	})

	ctx := context.Background()
	got, _, err := client.Security.GetUsersAssignedRole(ctx, rolename)
	if err != nil {
		t.Errorf("Security.GetUsersAssignedToRole returned error: %v", err)
	}
	if want := wantUsers; !cmp.Equal(got, want) {
		t.Errorf("Security.GetUsersAssignedToRole = %+v, want %+v", got, want)
	}

	const methodName = "GetUsersAssignedToRole"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetUsersAssignedRole(nil, rolename)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func Test_GrantUserPermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"
	var permission = &Permission{Action: "read", ResourceType: "db", Resource: []string{"*"}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/user/%s", username), func(w http.ResponseWriter, r *http.Request) {
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
	_, err := client.Security.GrantUserPermission(ctx, username, *permission)
	if err != nil {
		t.Errorf("Security.GrantUserPermission returned error: %v", err)
	}

	const methodName = "GrantUserPermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.GrantUserPermission(nil, username, *permission)
	})
}

func Test_RevokeUserPermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"
	var permission = &Permission{Action: "read", ResourceType: "db", Resource: []string{"*"}}

	mux.HandleFunc(fmt.Sprintf("/admin/permissions/user/%s/delete", username), func(w http.ResponseWriter, r *http.Request) {
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
	_, err := client.Security.RevokeUserPermission(ctx, username, *permission)
	if err != nil {
		t.Errorf("Security.RevokeUserPermission returned error: %v", err)
	}

	const methodName = "RevokeUserPermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.RevokeUserPermission(nil, username, *permission)
	})
}

func Test_AssignRole(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	rolename := "reader"
	username := "frodo"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/roles", username), func(w http.ResponseWriter, r *http.Request) {
		v := new(assignRoleRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", "application/json")

		want := &assignRoleRequest{Rolename: rolename}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.Security.AssignRole(ctx, username, rolename)
	if err != nil {
		t.Errorf("Security.AssignRole returned error: %v", err)
	}

	const methodName = "AssignRole"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.AssignRole(nil, username, rolename)
	})
}

func Test_UnassignRole(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	rolename := "reader"
	username := "frodo"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/roles/%s", username, rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.Security.UnassignRole(ctx, username, rolename)
	if err != nil {
		t.Errorf("Security.UnassignRole returned error: %v", err)
	}
	const methodName = "UnassignRole"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.UnassignRole(nil, username, rolename)
	})
}

func Test_NewPermission(t *testing.T) {
	newPermission := NewPermission(Read, Database, []string{"*"})

	want := &Permission{
		Action:       string(Read),
		ResourceType: string(Database),
		Resource:     []string{"*"},
	}

	if !cmp.Equal(newPermission, want) {
		t.Errorf("NewPermission returned %+v, want %+v", newPermission, want)
	}
}
