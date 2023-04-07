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

func TestUserService_WhoAmI(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()
	responseString := "frodo"
	want := newString(responseString)

	mux.HandleFunc("/admin/status/whoami", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "GET")
		testHeader(t, r, "Accept", mediaTypePlainText)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(responseString))
	})

	ctx := context.Background()
	got, _, err := client.User.WhoAmI(ctx)
	if err != nil {
		t.Errorf("User.WhoAmI returned error: %v", err)
	}
	if !cmp.Equal(got, want) {
		t.Errorf("User.WhoAmI = %+v, want %+v", got, want)
	}

	const methodName = "WhoAmI"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.WhoAmI(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestUserService_ListNames(t *testing.T) {
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
	got, _, err := client.User.ListNames(ctx)
	if err != nil {
		t.Errorf("User.ListNames returned error: %v", err)
	}
	if want := wantUsers; !cmp.Equal(got, want) {
		t.Errorf("User.ListNames = %+v, want %+v", got, want)
	}

	const methodName = "ListNames"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.ListNames(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestUserService_List(t *testing.T) {
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

	wantUsers := &listUsersResponse{
		Users: []User{
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
	got, _, err := client.User.List(ctx)
	if err != nil {
		t.Errorf("User.List returned error: %v", err)
	}
	if want := wantUsers.Users; !cmp.Equal(got, want) {
		t.Errorf("User.List = %+v, want %+v", got, want)
	}

	const methodName = "List"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.List(nil)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestUserService_Permissions(t *testing.T) {
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
	got, _, err := client.User.Permissions(ctx, "bob")
	if err != nil {
		t.Errorf("User.Permissions returned error: %v", err)
	}
	if want := wantUserPermissions; !cmp.Equal(got, want) {
		t.Errorf("User.UserPermissions = %+v, want %+v", got, want)
	}

	const methodName = "Permissions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.Permissions(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestUserService_Get(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	var userJSON = `{
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
	var wantUser = &User{
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
		w.Write([]byte(userJSON))
	})

	ctx := context.Background()
	got, _, err := client.User.Get(ctx, "bob")
	if err != nil {
		t.Errorf("User.Get returned error: %v", err)
	}
	if want := wantUser; !cmp.Equal(got, want) {
		t.Errorf("User.Get = %+v, want %+v", got, want)
	}

	const methodName = "Get"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.Get(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestUserService_UserEffectivePermissions(t *testing.T) {
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
	got, _, err := client.User.EffectivePermissions(ctx, "bob")
	if err != nil {
		t.Errorf("User.EffectivePermissions returned error: %v", err)
	}
	if want := wantUserEffectivePermissions; !cmp.Equal(got, want) {
		t.Errorf("User.EffectivePermissions = %+v, want %+v", got, want)
	}

	const methodName = "EffectivePermissions"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.EffectivePermissions(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestUserService_IsSuperuser(t *testing.T) {
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
	got, _, err := client.User.IsSuperuser(ctx, "bob")
	if err != nil {
		t.Errorf("User.IsSuperuser returned error: %v", err)
	}
	if want := isSuperuser; !cmp.Equal(got, want) {
		t.Errorf("User.IsSuperuser = %+v, want %+v", *got, *want)
	}

	const methodName = "IsSuperuser"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.IsSuperuser(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, *got)
		}
		return resp, err
	})
}

func TestUserService_IsEnabled(t *testing.T) {
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
	got, _, err := client.User.IsEnabled(ctx, "bob")
	if err != nil {
		t.Errorf("User.IsEnabled returned error: %v", err)
	}
	if want := isEnabled; !cmp.Equal(got, want) {
		t.Errorf("User.IsEnabled = %+v, want %+v", *got, *want)
	}

	const methodName = "IsEnabled"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.IsEnabled(nil, "someone")
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, *got)
		}
		return resp, err
	})
}

func TestUserService_Create(t *testing.T) {
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
	_, err := client.User.Create(ctx, username, strings.Join(password, ""))
	if err != nil {
		t.Errorf("User.Create returned error: %v", err)
	}

	const methodName = "Create"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.Create(nil, "someone", "password")
	})
}

func TestUserService_Delete(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	username := "bob"
	mux.HandleFunc(fmt.Sprintf("/admin/users/%s", username), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")

		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.User.Delete(ctx, username)
	if err != nil {
		t.Errorf("User.Delete returned error: %v", err)
	}
	const methodName = "Delete"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.Delete(nil, "someone")
	})
}

func TestUserService_ChangePassword(t *testing.T) {
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
	_, err := client.User.ChangePassword(ctx, username, password)
	if err != nil {
		t.Errorf("User.ChangePassword returned error: %v", err)
	}
	const methodName = "ChangePassword"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.ChangePassword(nil, "someone", "password")
	})
}

func TestUserService_Enable(t *testing.T) {
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
	_, err := client.User.Enable(ctx, username)
	if err != nil {
		t.Errorf("User.Enable returned error: %v", err)
	}

	const methodName = "Enable"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.Enable(nil, username)
	})
}

func TestUserService_Disable(t *testing.T) {
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
	_, err := client.User.Disable(ctx, username)
	if err != nil {
		t.Errorf("User.Disable returned error: %v", err)
	}

	const methodName = "Disable"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.Disable(nil, username)
	})
}

func TestUserService_GrantPermission(t *testing.T) {
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
	_, err := client.User.GrantPermission(ctx, username, *permission)
	if err != nil {
		t.Errorf("User.GrantPermission returned error: %v", err)
	}

	const methodName = "GrantPermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.GrantPermission(nil, username, *permission)
	})
}

func TestUserService_RevokePermission(t *testing.T) {
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
	_, err := client.User.RevokePermission(ctx, username, *permission)
	if err != nil {
		t.Errorf("User.RevokePermission returned error: %v", err)
	}

	const methodName = "RevokePermission"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.RevokePermission(nil, username, *permission)
	})
}

func TestUserService_ListNamesAssignedRole(t *testing.T) {
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
	got, _, err := client.User.ListNamesAssignedRole(ctx, rolename)
	if err != nil {
		t.Errorf("User.ListNamesAssignedRole returned error: %v", err)
	}
	if want := wantUsers; !cmp.Equal(got, want) {
		t.Errorf("User.ListNamesAssignedRole = %+v, want %+v", got, want)
	}

	const methodName = "ListNamesAssignedRole"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.ListNamesAssignedRole(nil, rolename)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}

func TestUserService_AssignRole(t *testing.T) {
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
	_, err := client.User.AssignRole(ctx, username, rolename)
	if err != nil {
		t.Errorf("User.AssignRole returned error: %v", err)
	}

	const methodName = "AssignRole"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.AssignRole(nil, username, rolename)
	})
}

func TestUserService_UnassignRole(t *testing.T) {
	client, mux, _, teardown := setup()
	defer teardown()

	rolename := "reader"
	username := "frodo"

	mux.HandleFunc(fmt.Sprintf("/admin/users/%s/roles/%s", username, rolename), func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, "DELETE")
		w.WriteHeader(http.StatusNoContent)
	})

	ctx := context.Background()
	_, err := client.User.UnassignRole(ctx, username, rolename)
	if err != nil {
		t.Errorf("User.UnassignRole returned error: %v", err)
	}
	const methodName = "UnassignRole"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.UnassignRole(nil, username, rolename)
	})
}

func TestUserService_OverwriteRoles(t *testing.T) {
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
	_, err := client.User.OverwriteRoles(ctx, username, roles)
	if err != nil {
		t.Errorf("User.OverwriteRoles returned error: %v", err)
	}

	const methodName = "OverwriteRoles"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		return client.User.OverwriteRoles(nil, username, roles)
	})
}

func TestUserService_Roles(t *testing.T) {
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
	got, _, err := client.User.Roles(ctx, username)
	if err != nil {
		t.Errorf("User.Roles returned error: %v", err)
	}
	if want := wantGetRoles; !cmp.Equal(got, want) {
		t.Errorf("User.Roles = %+v, want %+v", got, want)
	}

	const methodName = "Roles"
	testNewRequestAndDoFailure(t, methodName, client, func() (*Response, error) {
		got, resp, err := client.User.Roles(nil, username)
		if got != nil {
			t.Errorf("testNewRequestAndDoFailure %v = %#v, want nil", methodName, got)
		}
		return resp, err
	})
}
