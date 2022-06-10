package stardog

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

func (client *Client) Alive(ctx context.Context) (bool, error) {
	url := fmt.Sprintf("%s/admin/alive", client.BaseURL)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	var b struct{}
	if err := client.sendRequest(request, &b); err != nil {
		return false, err
	}

	return true, nil
}

type UserList struct {
	Users []string `json:"users"`
}

// Get all users in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/listUsers
func (client *Client) GetUsers(ctx context.Context) (*UserList, error) {
	url := fmt.Sprintf("%s/admin/users", client.BaseURL)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var userList UserList
	if err := client.sendRequest(request, &userList); err != nil {
		return nil, err
	}

	return &userList, nil
}

type credentials struct {
	Username string   `json:"username"`
	Password []string `json:"password"`
}

// Add a user to the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUser
func (client *Client) AddUser(ctx context.Context, username string, password string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users", client.BaseURL)

	credentials := credentials{
		Username: username,
		Password: strings.Split(password, ""),
	}

	reqBody, _ := json.Marshal(credentials)
	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuffer)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Delete a user from the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/deleteUser
func (client *Client) DeleteUser(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s", client.BaseURL, username)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Change user's password
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/changePassword
func (client *Client) ChangeUserPassword(ctx context.Context, username string, password string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/pwd", client.BaseURL, username)

	reqBody, _ := json.Marshal(map[string]string{
		"password": password,
	})
	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuffer)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil

}

// Represents a permission
type Permission struct {
	Action       string   `json:"action"`
	ResourceType string   `json:"resource_type"`
	Resource     []string `json:"resource"`
}

// Create an instance of
func NewPermission(action string, resource_type string, resource []string) *Permission {
	permission := Permission{
		Action:       action,
		ResourceType: resource_type,
		Resource:     resource,
	}
	return &permission
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (client *Client) AddUserPermission(ctx context.Context, username string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/user/%s/", client.BaseURL, username)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuf)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteUserPermission
func (client *Client) DeleteUserPermission(ctx context.Context, username string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/user/%s/delete", client.BaseURL, username)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuf)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

type UserDetails struct {
	Enabled     bool     `json:"enabled"`
	Superuser   bool     `json:"superuser"`
	Roles       []string `json:"roles"`
	Permissions []struct {
		Action       string   `json:"action"`
		ResourceType string   `json:"resource_type"`
		Resource     []string `json:"resource"`
		Explicit     bool     `json:"explicit"`
	} `json:"permissions"`
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUser
func (client *Client) GetUserDetails(ctx context.Context, username string) (*UserDetails, error) {
	url := fmt.Sprintf("%s/admin/users/%s", client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var userDetails UserDetails
	if err := client.sendRequest(request, &userDetails); err != nil {
		return nil, err
	}

	return &userDetails, nil
}

type RoleList struct {
	Roles []string `json:"roles"`
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/listRoles
func (client *Client) GetRoles(ctx context.Context) (*RoleList, error) {
	url := fmt.Sprintf("%s/admin/roles", client.BaseURL)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var roleList RoleList
	if err := client.sendRequest(request, &roleList); err != nil {
		return nil, err
	}

	return &roleList, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/addRole
func (client *Client) CreateRole(ctx context.Context, rolename string) (bool, error) {
	url := fmt.Sprintf("%s/admin/roles", client.BaseURL)

	reqBody, _ := json.Marshal(map[string]string{
		"rolename": rolename,
	})
	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuffer)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/deleteRole
func (client *Client) DeleteRole(ctx context.Context, rolename string, force bool) (bool, error) {
	url := fmt.Sprintf("%s/admin/roles/%s?force=%t", client.BaseURL, rolename, force)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUserRoles
func (client *Client) GetUsersAssignedRoles(ctx context.Context, username string) (*RoleList, error) {
	url := fmt.Sprintf("%s/admin/users/%s/roles", client.BaseURL, username)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var roleList RoleList
	if err := client.sendRequest(request, &roleList); err != nil {
		return nil, err
	}

	return &roleList, nil

}

type RolePermissions struct {
	Permissions []struct {
		Action       string   `json:"action"`
		ResourceType string   `json:"resource_type"`
		Resource     []string `json:"resource"`
	} `json:"permissions"`
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getRolePermissions
func (client *Client) RolePermissions(ctx context.Context, rolename string) (*RolePermissions, error) {
	url := fmt.Sprintf("%s/admin/permissions/role/%s", client.BaseURL, rolename)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var rolePermissionsResponse RolePermissions
	if err := client.sendRequest(request, &rolePermissionsResponse); err != nil {
		return nil, err
	}

	return &rolePermissionsResponse, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (client *Client) CreateRolePermission(ctx context.Context, rolename string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/role/%s/", client.BaseURL, rolename)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuf)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteRolePermission
func (client *Client) DeleteRolePermission(ctx context.Context, rolename string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/role/%s/delete", client.BaseURL, rolename)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuf)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

//Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/getUsersWithRole
func (client *Client) GetUsersAssignedToRole(ctx context.Context, rolename string) (*UserList, error) {
	url := fmt.Sprintf("%s/admin/roles/%s/users", client.BaseURL, rolename)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var userList UserList
	if err := client.sendRequest(req, &userList); err != nil {
		return nil, err
	}
	return &userList, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUserRole
func (client *Client) AssignRoleToUser(ctx context.Context, username string, rolename string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/roles", client.BaseURL, username)

	reqBody, _ := json.Marshal(map[string]string{
		"rolename": rolename,
	})
	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuffer)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/removeUserRole
func (client *Client) RemoveRoleFromUser(ctx context.Context, username string, rolename string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/roles/%s", client.BaseURL, username, rolename)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Stardog API:  https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (client *Client) SetIsEnabled(ctx context.Context, username string, enabled bool) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/enabled", client.BaseURL, username)

	reqBody, _ := json.Marshal(map[string]string{
		"enabled": strconv.FormatBool(enabled),
	})

	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuffer)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

type isEnabledResponse struct {
	Enabled bool `json:"enabled"`
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/userEnabled
func (client *Client) IsEnabled(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/enabled", client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	var isEnabledResponse isEnabledResponse
	if err := client.sendRequest(request, &isEnabledResponse); err != nil {
		return false, err
	}

	return isEnabledResponse.Enabled, nil

}

type isSuperUserResponse struct {
	Superuser bool `json:"superuser"`
}

// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/isSuper
func (client *Client) IsSuperuser(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/superuser", client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	var isSuperUserResponse isSuperUserResponse
	if err := client.sendRequest(request, &isSuperUserResponse); err != nil {
		return false, err
	}

	return isSuperUserResponse.Superuser, nil

}
