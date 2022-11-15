package stardog

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

type SecurityService service

type ListUsersResponse struct {
	Users []string `json:"users"`
}

// List all users in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/listUsers
func (s *SecurityService) ListUsers(ctx context.Context) (*ListUsersResponse, *Response, error) {
	u := "admin/users"
	req, err := s.client.NewRequest("GET", u, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}

	var listUsersResponse *ListUsersResponse
	resp, err := s.client.Do(ctx, req, &listUsersResponse)
	if err != nil {
		return nil, resp, err
	}
	return listUsersResponse, resp, err
}

type GetUserDetailsResponse struct {
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

type GetUserPermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// Get permissions explicitly assigned to user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getUserPermissions
func (s *SecurityService) GetUserPermissions(ctx context.Context, username string) (*GetUserPermissionsResponse, *Response, error) {
	u := fmt.Sprintf("admin/permissions/user/%s", username)

	request, err := s.client.NewRequest("GET", u, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}

	var getUserPermissionsResponse *GetUserPermissionsResponse
	resp, err := s.client.Do(ctx, request, &getUserPermissionsResponse)
	if err != nil {
		return nil, resp, err
	}
	return getUserPermissionsResponse, resp, nil
}

// Get all permissions assigned to a given user as well as those granted by assigned roles
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getEffectiveUserPermissions
func (s *SecurityService) GetUserEffectivePermissions(ctx context.Context, username string) (*GetUserPermissionsResponse, *Response, error) {
	u := fmt.Sprintf("%s/admin/permissions/effective/user/%s", s.client.BaseURL, username)

	request, err := s.client.NewRequest("GET", u, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}

	var getUserPermissionsResponse *GetUserPermissionsResponse
	resp, err := s.client.Do(ctx, request, &getUserPermissionsResponse)
	if err != nil {
		return nil, resp, err
	}

	return getUserPermissionsResponse, resp, nil
}

// Get user attributes, roles, and effective permissions
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUser
func (s *SecurityService) GetUserDetails(ctx context.Context, username string) (*GetUserDetailsResponse, *Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)

	request, err := s.client.NewRequest("GET", u, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}

	var getUserDetailsResponse *GetUserDetailsResponse
	resp, err := s.client.Do(ctx, request, &getUserDetailsResponse)
	if err != nil {
		return nil, resp, err
	}

	return getUserDetailsResponse, resp, nil
}

type IsSuperuserResponse struct {
	Superuser bool `json:"superuser"`
}

// Is specified user a superuser
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/isSuper
func (s *SecurityService) IsSuperuser(ctx context.Context, username string) (*IsSuperuserResponse, *Response, error) {
	u := fmt.Sprintf("admin/users/%s/superuser", username)

	request, err := s.client.NewRequest("GET", u, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}

	var isSuperuserResponse *IsSuperuserResponse
	resp, err := s.client.Do(ctx, request, &isSuperuserResponse)
	if err != nil {
		return nil, resp, err
	}

	return isSuperuserResponse, resp, nil
}

type IsEnabledResponse struct {
	Enabled bool `json:"enabled"`
}

// Is user enabled or disabled
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/userEnabled
func (s *SecurityService) IsEnabled(ctx context.Context, username string) (*IsEnabledResponse, *Response, error) {
	u := fmt.Sprintf("admin/users/%s/enabled", username)

	request, err := s.client.NewRequest("GET", u, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}

	var isEnabled *IsEnabledResponse
	resp, err := s.client.Do(ctx, request, &isEnabled)
	if err != nil {
		return nil, resp, err
	}
	return isEnabled, resp, nil
}

type credentials struct {
	Username string   `json:"username"`
	Password []string `json:"password"`
}

// Add a user to the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUser
func (s *SecurityService) CreateUser(ctx context.Context, username string, password string) (bool, *Response, error) {
	u := "admin/users"

	credentials := credentials{
		Username: username,
		Password: strings.Split(password, ""),
	}
	request, err := s.client.NewRequest("POST", u, "application/json", "application/json", credentials)
	fmt.Println(request.Header.Get("Content-Type"))
	if err != nil {
		return false, nil, err
	}

	resp, err := s.client.Do(ctx, request, nil)
	isCreated, err := parseBoolResponse(err)
	return isCreated, resp, nil
}

// Delete a user from the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/deleteUser
func (s *SecurityService) DeleteUser(ctx context.Context, username string) (bool, *Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)
	request, err := s.client.NewRequest("DELETE", u, "", "application/json", nil)
	if err != nil {
		return false, nil, err
	}

	resp, err := s.client.Do(ctx, request, nil)

	isDeleted, err := parseBoolResponse(err)
	return isDeleted, resp, nil
}

type changePasswordRequest struct {
	Password string `json:"password"`
}

// Change user's password
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/changePassword
func (s *SecurityService) ChangeUserPassword(ctx context.Context, username string, password string) (bool, *Response, error) {
	u := fmt.Sprintf("admin/users/%s/pwd", username)

	reqBody := changePasswordRequest{
		Password: password,
	}
	request, err := s.client.NewRequest("PUT", u, "application/json", "application/json", reqBody)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, request, nil)

	isPasswordChanged, err := parseBoolResponse(err)
	return isPasswordChanged, resp, err
}

type setEnabledRequest struct {
	Enabled string `json:"enabled"`
}

// Enable/disable user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *SecurityService) SetEnabled(ctx context.Context, username string, enabled bool) (bool, *Response, error) {
	url := fmt.Sprintf("admin/users/%s/enabled", username)

	reqBody := setEnabledRequest{
		Enabled: strconv.FormatBool(enabled),
	}

	req, err := s.client.NewRequest("PUT", url, "application/json", "application/json", reqBody)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)

	success, err := parseBoolResponse(err)
	return success, resp, nil
}

// Retrieve a list of all roles explicitly assigned to a user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUserRoles
func (s *SecurityService) ListRolesAssignedToUser(ctx context.Context, username string) (*ListRolesResponse, *Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles", username)
	req, err := s.client.NewRequest("GET", url, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}
	var listRolesResponse *ListRolesResponse
	resp, err := s.client.Do(ctx, req, &listRolesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listRolesResponse, resp, err
}

type ListRolesResponse struct {
	Roles []string `json:"roles"`
}

// List the names of all roles in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/listRoles
func (s *SecurityService) ListRoles(ctx context.Context) (*ListRolesResponse, *Response, error) {
	u := "admin/roles"
	req, err := s.client.NewRequest("GET", u, "", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}
	var listRolesResponse *ListRolesResponse
	resp, err := s.client.Do(ctx, req, &listRolesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listRolesResponse, resp, nil
}

type createRoleRequest struct {
	Rolename string `json:"rolename"`
}

// Create a role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/addRole
func (s *SecurityService) CreateRole(ctx context.Context, rolename string) (bool, *Response, error) {
	u := "admin/roles"
	reqBody := createRoleRequest{
		Rolename: rolename,
	}
	req, err := s.client.NewRequest("POST", u, "application/json", "", reqBody)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	created, err := parseBoolResponse(err)
	return created, resp, err
}

type RolePermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// Get all permissions granted to a given role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getRolePermissions
func (s *SecurityService) GetRolePermissions(ctx context.Context, rolename string) (*RolePermissionsResponse, *Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s", rolename)
	req, err := s.client.NewRequest("GET", url, "application/json", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}
	var rolePermissionsResponse *RolePermissionsResponse
	resp, err := s.client.Do(ctx, req, &rolePermissionsResponse)
	if err != nil {
		return nil, resp, err
	}
	return rolePermissionsResponse, resp, nil
}

// Assign a permission to a specified role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *SecurityService) GrantRolePermission(ctx context.Context, rolename string, permission Permission) (bool, *Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s/", rolename)
	req, err := s.client.NewRequest("PUT", url, "application/json", "application/json", permission)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	isGranted, err := parseBoolResponse(err)
	return isGranted, resp, err
}

// Revoke permission from specified role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteRolePermission
func (s *SecurityService) RevokeRolePermission(ctx context.Context, rolename string, permission Permission) (bool, *Response, error) {
	url := fmt.Sprintf("%s/admin/permissions/role/%s/delete", s.client.BaseURL, rolename)
	req, err := s.client.NewRequest("POST", url, "application/json", "application/json", permission)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	isRevoked, err := parseBoolResponse(err)
	return isRevoked, resp, err
}

type DeleteRoleOptions struct {
	Force bool `url:"force"`
}

// Delete a role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/deleteRole
func (s *SecurityService) DeleteRole(ctx context.Context, rolename string, opts *DeleteRoleOptions) (bool, *Response, error) {
	u := fmt.Sprintf("admin/roles/%s", rolename)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return false, nil, err
	}
	fmt.Println(urlWithOptions)
	req, err := s.client.NewRequest("DELETE", urlWithOptions, "application/json", "", nil)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	isDeleted, err := parseBoolResponse(err)
	return isDeleted, resp, err
}

// Returns a list of all users that have a specific role
//
//Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/getUsersWithRole
func (s *SecurityService) ListUsersAssignedToRole(ctx context.Context, rolename string) (*ListUsersResponse, *Response, error) {
	u := fmt.Sprintf("admin/roles/%s/users", rolename)
	req, err := s.client.NewRequest("GET", u, "application/json", "application/json", nil)
	if err != nil {
		return nil, nil, err
	}

	var listUsersResponse *ListUsersResponse
	resp, err := s.client.Do(ctx, req, &listUsersResponse)
	if err != nil {
		return nil, resp, err
	}
	return listUsersResponse, resp, err
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
func (s *SecurityService) GrantUserPermission(ctx context.Context, username string, permission Permission) (bool, *Response, error) {
	url := fmt.Sprintf("admin/permissions/user/%s", username)
	req, err := s.client.NewRequest("PUT", url, "application/json", "application/json", permission)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	isGranted, err := parseBoolResponse(err)
	return isGranted, resp, err
}

// Revoke a permission from a given user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteUserPermission
func (s *SecurityService) RevokeUserPermission(ctx context.Context, username string, permission Permission) (bool, *Response, error) {
	url := fmt.Sprintf("admin/permissions/user/%s/delete", username)

	req, err := s.client.NewRequest("POST", url, "application/json", "application/json", permission)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	isRevoked, err := parseBoolResponse(err)
	return isRevoked, resp, err
}

type assignRoleToUserRequest struct {
	Rolename string `json:"rolename"`
}

// Assign a role to user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUserRole
func (s *SecurityService) AssignRoleToUser(ctx context.Context, username string, rolename string) (bool, *Response, error) {
	url := fmt.Sprintf("%s/admin/users/%s/roles", username)
	reqBody := assignRoleToUserRequest{
		Rolename: rolename,
	}
	req, err := s.client.NewRequest("POST", url, "application/json", "application/json", reqBody)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	isAssigned, err := parseBoolResponse(err)
	return isAssigned, resp, err
}

// Remove role from a user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/removeUserRole
func (s *SecurityService) RemoveRoleFromUser(ctx context.Context, username string, rolename string) (bool, *Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles/%s", username, rolename)
  req, err := s.client.NewRequest("DELETE", url, "application/json", "application/json", nil)
	if err != nil {
		return false, nil, err
	}
	resp, err := s.client.Do(ctx, req, nil)
	isDeleted, err := parseBoolResponse(err)
	return isDeleted, resp, err
}
