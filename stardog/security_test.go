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

func TestPermissionAction_Valid(t *testing.T) {
	r := PermissionAction(100)
	if r.Valid() {
		t.Errorf("should be an invalid PermissionAction")
	}
	if r.String() != PermissionActionUnknown.String() {
		t.Errorf("PermissionAction string value should be unknown")
	}
}

func TestPermissionResourceType_Valid(t *testing.T) {
	r := PermissionResourceType(100)
	if r.Valid() {
		t.Errorf("should be an invalid PermissionResourceType")
	}
	if r.String() != PermissionResourceTypeUnknown.String() {
		t.Errorf("PermissionResourceType string value should be unknown")
	}
}

func TestPermissionAction_UnmarshalText(t *testing.T) {
	r := PermissionActionWrite
	r.UnmarshalText([]byte("write"))
	if r != PermissionActionWrite {
		t.Error("should still be PermissionActionWrite")
	}
	r.UnmarshalText([]byte("trite"))
	if r.Valid() {
		t.Error("should be an invalid PermissionAction")
	}
}

func TestPermissionResourceType_UnmarshalText(t *testing.T) {
	r := PermissionResourceTypeDatabaseAdmin
	r.UnmarshalText([]byte("admin"))
	if r != PermissionResourceTypeDatabaseAdmin {
		t.Error("should still be PermissionResourceTypeDatabaseAdmin")
	}
	r.UnmarshalText([]byte("trite"))
	if r.Valid() {
		t.Error("should be an invalid PermissionResourceType")
	}
}

func TestSecurityService_GetUsers(t *testing.T) {
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

func TestSecurityService_GetUsersWithDetails(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	usersJSON := []byte(`
{
  "users": [
    {
      "username": "admin",
      "enabled": true,
      "superuser": true,
      "roles": [],
      "permissions": []
    },
    {
      "username": "frodo",
      "enabled": true,
      "superuser": false,
      "roles": ["reader", "writer", "creator"],
      "permissions": [
        {
          "action": "READ",
          "resource_type": "db",
          "resource": [
            "myDatabase"
          ],
          "explicit": true
        }
      ]
    }
  ]
}
  `)

	wantUsers := &getUsersWithDetailsResponse{
		Users: []UserDetails{
			{
				Username:             newString("admin"),
				Roles:                []string{},
				Enabled:              true,
				Superuser:            true,
				EffectivePermissions: []EffectivePermission{},
			},
			{
				Username:  newString("frodo"),
				Roles:     []string{"reader", "writer", "creator"},
				Enabled:   true,
				Superuser: false,
				EffectivePermissions: []EffectivePermission{
					{
						Explicit: true,
						Permission: Permission{
							Action:       PermissionActionRead,
							ResourceType: PermissionResourceTypeDatabase,
							Resource:     []string{"myDatabase"},
						}},
				},
			},
		},
	}

	mux.HandleFunc("/admin/users/list", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(usersJSON)
	})

	ctx := context.Background()
	got, _, err := client.Security.GetUsersWithDetails(ctx)
	if err != nil {
		t.Errorf("Security.GetUsersWithDetails returned error: %v", err)
	}
	if want := wantUsers.Users; !cmp.Equal(got, want) {
		t.Errorf("Security.GetUsersWithDetails = %+v, want %+v", got, want)
	}

	const methodName = "GetUsersWithDetails"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetUsersWithDetails(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestSecurityService_GetUserPermissions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var userPermissionsJSON = `{
    "permissions": [
      {"action":"READ","resource_type":"named-graph","resource":["db1"]}
      ]
    }`
	var wantUserPermissions = []Permission{
		{
			Action:       PermissionActionRead,
			ResourceType: PermissionResourceTypeNamedGraph,
			Resource:     []string{"db1"}},
	}

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

func TestSecurityService_GetUserEffectivePermissions(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var userEffectivePermissionsJSON = `{
    "permissions": [
      {"action":"DELETE","resource_type":"named-graph","resource":["db1"], "explicit": false}
      ]
    }`
	var wantUserEffectivePermissions = []EffectivePermission{
		{
			Permission: Permission{
				Action:       PermissionActionDelete,
				ResourceType: PermissionResourceTypeNamedGraph,
				Resource:     []string{"db1"},
			},
			Explicit: false},
	}

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

func TestSecurityService_GetUserDetails(t *testing.T) {
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
		EffectivePermissions: []EffectivePermission{
			{
				Permission: Permission{
					Action:       PermissionActionRead,
					ResourceType: PermissionResourceTypeDatabase,
					Resource:     []string{"myDatabase"},
				},
				Explicit: true,
			},
			{
				Permission: Permission{
					Action:       PermissionActionRead,
					ResourceType: PermissionResourceTypeUser,
					Resource:     []string{"frodo"},
				},
				Explicit: true,
			},
			{
				Permission: Permission{
					Action:       PermissionActionWrite,
					ResourceType: PermissionResourceTypeUser,
					Resource:     []string{"frodo"},
				},
				Explicit: true,
			},
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

func TestSecurityService_IsSuperuser(t *testing.T) {
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

func TestSecurityService_IsEnabled(t *testing.T) {
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

func TestSecurityService_CreateUser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	const username = "frodo"
	var password = strings.Split("gandalf", "")

	mux.HandleFunc("/admin/users", func(w http.ResponseWriter, r *http.Request) {
		v := new(createUserRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "POST")
		testHeader(t, r, "Content-Type", "application/json")

		want := &createUserRequest{Username: username, Password: password}
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

func TestSecurityService_DeleteUser(t *testing.T) {
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

func TestSecurityService_ChangeUserPassword(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"
	var password = "somePassword"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/pwd", username), func(w http.ResponseWriter, r *http.Request) {
		v := new(changeUserPasswordRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", "application/json")

		want := &changeUserPasswordRequest{Password: password}
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

func TestSecurityService_EnableUser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/enabled", username), func(w http.ResponseWriter, r *http.Request) {
		v := new(enableUserRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", "application/json")

		want := &enableUserRequest{Enabled: true}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.Security.EnableUser(ctx, username)
	if err != nil {
		t.Errorf("Security.EnableUser returned error: %v", err)
	}

	const methodName = "EnableUser"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.EnableUser(nil, username)
	})
}

func TestSecurityService_DisableUser(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/enabled", username), func(w http.ResponseWriter, r *http.Request) {
		v := new(enableUserRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", "application/json")

		want := &enableUserRequest{Enabled: false}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusOK)
	})

	ctx := context.Background()
	_, err := client.Security.DisableUser(ctx, username)
	if err != nil {
		t.Errorf("Security.DisableUser returned error: %v", err)
	}

	const methodName = "DisableUser"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.DisableUser(nil, username)
	})
}

func TestSecurityService_GetRoles(t *testing.T) {
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

func TestSecurityService_GetRolesWithDetails(t *testing.T) {
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
	wantRoles := &getRolesWithDetailsResponse{
		Roles: []RoleDetails{
			{
				Rolename: "reader",
				Permissions: []Permission{
					{
						Action:       PermissionActionRead,
						ResourceType: PermissionResourceTypeAll,
						Resource:     []string{"*"},
					},
				},
			},
			{
				Rolename: "writer",
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
	got, _, err := client.Security.GetRolesWithDetails(ctx)
	if err != nil {
		t.Errorf("Security.GetRolesWithDetails returned error: %v", err)
	}
	if want := wantRoles.Roles; !cmp.Equal(got, want) {
		t.Errorf("Security.GetRolesWithDetails = %+v, want %+v", got, want)
	}

	const methodName = "GetRolesWithDetails"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.Security.GetRolesWithDetails(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestSecurityService_GetRolesAssignedToUser(t *testing.T) {
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

func TestSecurityService_CreateRole(t *testing.T) {
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

func TestSecurityService_GetRolePermissions(t *testing.T) {
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

func TestSecurityService_GrantRolePermission(t *testing.T) {
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
	_, err := client.Security.GrantRolePermission(ctx, rolename, *permission)
	if err != nil {
		t.Errorf("Security.GrantRolePermission returned error: %v", err)
	}

	const methodName = "GrantRolePermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.GrantRolePermission(nil, rolename, *permission)
	})
}

func TestSecurityService_RevokeRolePermission(t *testing.T) {
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
	_, err := client.Security.RevokeRolePermission(ctx, rolename, *permission)
	if err != nil {
		t.Errorf("Security.RevokeRolePermission returned error: %v", err)
	}

	const methodName = "RevokeRolePermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.RevokeRolePermission(nil, rolename, *permission)
	})
}

func TestSecurityService_DeleteRole(t *testing.T) {
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

func TestSecurityService_GetUsersAssignedRole(t *testing.T) {
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

func TestSecurityService_GrantUserPermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"
	var permission = &Permission{
		Action:       PermissionActionRead,
		ResourceType: PermissionResourceTypeDatabase,
		Resource:     []string{"*"},
	}

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

func TestSecurityService_RevokeUserPermission(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var username = "frodo"
	var permission = &Permission{
		Action:       PermissionActionRead,
		ResourceType: PermissionResourceTypeDatabase,
		Resource:     []string{"*"},
	}

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

func TestSecurityService_AssignRole(t *testing.T) {
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

func TestSecurityService_OverwriteRoles(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	roles := []string{"reader", "writer", "creator"}
	username := "frodo"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/roles", username), func(w http.ResponseWriter, r *http.Request) {
		v := new(overwriteRolesRequest)
		json.NewDecoder(r.Body).Decode(v)
		testMethod(t, r, "PUT")
		testHeader(t, r, "Content-Type", "application/json")

		want := &overwriteRolesRequest{Roles: []string{"reader", "writer", "creator"}}
		if !cmp.Equal(v, want) {
			t.Errorf("Request body = %+v, want %+v", v, want)
		}

		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.Security.OverwriteRoles(ctx, username, roles)
	if err != nil {
		t.Errorf("Security.OverwriteRoles returned error: %v", err)
	}

	const methodName = "OverwriteRoles"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.Security.OverwriteRoles(nil, username, roles)
	})
}

func TestSecurityService_UnassignRole(t *testing.T) {
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
