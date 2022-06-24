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

type SecurityService service

type listUsersResponse struct {
	Users []string `json:"users"`
}

// List all users in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/listUsers
func (s *SecurityService) ListUsers(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/admin/users", s.client.BaseURL)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var listUsersResponse listUsersResponse
	if err := s.client.sendRequest(request, &listUsersResponse); err != nil {
		return nil, err
	}

	return listUsersResponse.Users, nil
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

type userPermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// Get permissions explicitly assigned to user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getUserPermissions
func (s *SecurityService) GetUserPermissions(ctx context.Context, username string) ([]Permission, error) {
	url := fmt.Sprintf("%s/admin/permissions/user/%s", s.client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var userPermissionsResponse userPermissionsResponse
	if err := s.client.sendRequest(request, &userPermissionsResponse); err != nil {
		return nil, err
	}

	return userPermissionsResponse.Permissions, nil
}

// Get all permissions assigned to a given user as well as those granted by assigned roles
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getEffectiveUserPermissions
func (s *SecurityService) GetUserEffectivePermissions(ctx context.Context, username string) ([]Permission, error) {
	url := fmt.Sprintf("%s/admin/permissions/effective/user/%s", s.client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var userPermissionsResponse userPermissionsResponse
	if err := s.client.sendRequest(request, &userPermissionsResponse); err != nil {
		return nil, err
	}

	return userPermissionsResponse.Permissions, nil
}

// Get user attributes, roles, and effective permissions
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUser
func (s *SecurityService) GetUserDetails(ctx context.Context, username string) (*UserDetails, error) {
	url := fmt.Sprintf("%s/admin/users/%s", s.client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var userDetails UserDetails
	if err := s.client.sendRequest(request, &userDetails); err != nil {
		return nil, err
	}

	return &userDetails, nil
}

type isSuperUserResponse struct {
	Superuser bool `json:"superuser"`
}

// Is specified user a superuser
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/isSuper
func (s *SecurityService) IsSuperuser(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/superuser", s.client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	var isSuperUserResponse isSuperUserResponse
	if err := s.client.sendRequest(request, &isSuperUserResponse); err != nil {
		return false, err
	}

	return isSuperUserResponse.Superuser, nil
}

type isEnabledResponse struct {
	Enabled bool `json:"enabled"`
}

// Is user enabled or disabled
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/userEnabled
func (s *SecurityService) IsEnabled(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/enabled", s.client.BaseURL, username)

	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return false, err
	}

	var isEnabledResponse isEnabledResponse
	if err := s.client.sendRequest(request, &isEnabledResponse); err != nil {
		return false, err
	}
	return isEnabledResponse.Enabled, nil
}

type credentials struct {
	Username string   `json:"username"`
	Password []string `json:"password"`
}

// Add a user to the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUser
func (s *SecurityService) CreateUser(ctx context.Context, username string, password string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users", s.client.BaseURL)

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
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Delete a user from the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/deleteUser
func (s *SecurityService) DeleteUser(ctx context.Context, username string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s", s.client.BaseURL, username)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Change user's password
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/changePassword
func (s *SecurityService) ChangeUserPassword(ctx context.Context, username string, password string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/pwd", s.client.BaseURL, username)

	reqBody, _ := json.Marshal(map[string]string{
		"password": password,
	})
	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuffer)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil

}

// Enable/disable user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *SecurityService) SetEnabled(ctx context.Context, username string, enabled bool) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/enabled", s.client.BaseURL, username)

	reqBody, _ := json.Marshal(map[string]string{
		"enabled": strconv.FormatBool(enabled),
	})

	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuffer)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Retrieve a list of all roles explicitly assigned to a user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUserRoles
func (s *SecurityService) ListRolesAssignedToUser(ctx context.Context, username string) ([]string, error) {
	url := fmt.Sprintf("%s/admin/users/%s/roles", s.client.BaseURL, username)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var listRolesResponse listRolesResponse
	if err := s.client.sendRequest(request, &listRolesResponse); err != nil {
		return nil, err
	}

	return listRolesResponse.Roles, nil

}

type listRolesResponse struct {
	Roles []string `json:"roles"`
}

// List the names of all roles in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/listRoles
func (s *SecurityService) ListRoles(ctx context.Context) ([]string, error) {
	url := fmt.Sprintf("%s/admin/roles", s.client.BaseURL)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var listRolesResponse listRolesResponse
	if err := s.client.sendRequest(request, &listRolesResponse); err != nil {
		return nil, err
	}
	return listRolesResponse.Roles, nil
}

// Create a role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/addRole
func (s *SecurityService) CreateRole(ctx context.Context, rolename string) (bool, error) {
	url := fmt.Sprintf("%s/admin/roles", s.client.BaseURL)

	reqBody, _ := json.Marshal(map[string]string{
		"rolename": rolename,
	})
	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuffer)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

type rolePermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// Get all permissions granted to a given role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getRolePermissions
func (s *SecurityService) GetRolePermissions(ctx context.Context, rolename string) ([]Permission, error) {
	url := fmt.Sprintf("%s/admin/permissions/role/%s", s.client.BaseURL, rolename)
	request, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var rolePermissionsResponse rolePermissionsResponse
	if err := s.client.sendRequest(request, &rolePermissionsResponse); err != nil {
		return nil, err
	}

	return rolePermissionsResponse.Permissions, nil
}

// Assign a permission to a specified role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *SecurityService) GrantRolePermission(ctx context.Context, rolename string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/role/%s/", s.client.BaseURL, rolename)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuf)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Revoke permission from specified role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteRolePermission
func (s *SecurityService) RevokeRolePermission(ctx context.Context, rolename string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/role/%s/delete", s.client.BaseURL, rolename)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuf)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Delete a role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/deleteRole
func (s *SecurityService) DeleteRole(ctx context.Context, rolename string, force bool) (bool, error) {
	url := fmt.Sprintf("%s/admin/roles/%s?force=%t", s.client.BaseURL, rolename, force)

	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return false, err
	}

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Returns a list of all users that have a specific role
//
//Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/getUsersWithRole
func (s *SecurityService) ListUsersAssignedToRole(ctx context.Context, rolename string) ([]string, error) {
	url := fmt.Sprintf("%s/admin/roles/%s/users", s.client.BaseURL, rolename)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	var listUsersResponse listUsersResponse
	if err := s.client.sendRequest(req, &listUsersResponse); err != nil {
		return nil, err
	}
	return listUsersResponse.Users, nil
}

// Represents a permission
type Permission struct {
	Action       string   `json:"action"`
	ResourceType string   `json:"resource_type"`
	Resource     []string `json:"resource"`
}

// Helper function to create a Permission struct
func NewPermission(action string, resource_type string, resource []string) *Permission {
	permission := Permission{
		Action:       action,
		ResourceType: resource_type,
		Resource:     resource,
	}
	return &permission
}

// Grant a permission directly to a specified user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *SecurityService) GrantUserPermission(ctx context.Context, username string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/user/%s/", s.client.BaseURL, username)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "PUT", url, payloadBuf)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Revoke a permission from a given user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteUserPermission
func (s *SecurityService) RevokeUserPermission(ctx context.Context, username string, permission Permission) (bool, error) {
	url := fmt.Sprintf("%s/admin/permissions/user/%s/delete", s.client.BaseURL, username)

	payloadBuf := new(bytes.Buffer)
	json.NewEncoder(payloadBuf).Encode(permission)
	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuf)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Assign a role to user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUserRole
func (s *SecurityService) AssignRoleToUser(ctx context.Context, username string, rolename string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/roles", s.client.BaseURL, username)

	reqBody, _ := json.Marshal(map[string]string{
		"rolename": rolename,
	})
	payloadBuffer := bytes.NewBuffer(reqBody)

	req, err := http.NewRequestWithContext(ctx, "POST", url, payloadBuffer)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}

// Remove role from a user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/removeUserRole
func (s *SecurityService) RemoveRoleFromUser(ctx context.Context, username string, rolename string) (bool, error) {
	url := fmt.Sprintf("%s/admin/users/%s/roles/%s", s.client.BaseURL, username, rolename)
	req, err := http.NewRequestWithContext(ctx, "DELETE", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Add("Content-type", "application/json")

	var res struct{}
	if err := s.client.sendRequest(req, &res); err != nil {
		return false, err
	}
	return true, nil
}
