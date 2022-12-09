package stardog

import (
	"context"
	"fmt"
	"strings"
)

type SecurityService service

type UserDetails struct {
	Enabled     bool         `json:"enabled"`
	Superuser   bool         `json:"superuser"`
	Roles       []string     `json:"roles"`
	Permissions []Permission `json:"permissions"`
}

type usersResponse struct {
	Users []string `json:"users"`
}

type isSuperuserResponse struct {
	Superuser bool `json:"superuser"`
}

type isEnabledResponse struct {
	Enabled bool `json:"enabled"`
}

type userPermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

type credentials struct {
	Username string   `json:"username"`
	Password []string `json:"password"`
}

type changePasswordRequest struct {
	Password string `json:"password"`
}

type enableRequest struct {
	Enabled bool `json:"enabled"`
}

type rolesResponse struct {
	Roles []string `json:"roles"`
}

type createRoleRequest struct {
	Rolename string `json:"rolename"`
}

type rolePermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

type DeleteRoleOptions struct {
	Force bool `url:"force"`
}

type assignRoleRequest struct {
	Rolename string `json:"rolename"`
}

type Action string

const (
	Read       Action = "read"
	Write      Action = "write"
	Create     Action = "create"
	Delete     Action = "delete"
	Grant      Action = "grant"
	Revoke     Action = "revoke"
	Execute    Action = "execute"
	AllActions Action = "*"
)

type ResourceType string

const (
	Database          ResourceType = "db"
	Metadata          ResourceType = "metadata"
	User              ResourceType = "user"
	Role              ResourceType = "role"
	NamedGraph        ResourceType = "named-graph"
	VirtualGraph      ResourceType = "virtual-graph"
	DataSource        ResourceType = "data-source"
	ServerAdmin       ResourceType = "dbms-admin"
	DatabaseAdmin     ResourceType = "admin"
	SensitiveProperty ResourceType = "sensitive-properties"
	StoredQuery       ResourceType = "stored-query"
	AllResourceTypes  ResourceType = "*"
)

// Helper function to create a Permission struct
func NewPermission(action Action, resource_type ResourceType, resource []string) *Permission {
	permission := Permission{
		Action:       string(action),
		ResourceType: string(resource_type),
		Resource:     resource,
	}
	return &permission
}

// Represents a user/role permission
type Permission struct {
	Action       string   `json:"action"`
	ResourceType string   `json:"resource_type"`
	Resource     []string `json:"resource"`
	Explicit     *bool    `json:"explicit,omitempty"`
}

// Get names of all users
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/GetUsers/operation/listUsers
func (s *SecurityService) GetUsers(ctx context.Context) ([]string, *Response, error) {
	u := "admin/users"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var usersResponse *usersResponse
	resp, err := s.client.Do(ctx, req, &usersResponse)
	if err != nil {
		return nil, resp, err
	}
	return usersResponse.Users, resp, err
}

// Get permissions explicitly assigned to user. Permissions granted by a role the user may be assigned will
// not be contained in the response. Use UserEffectivePermissions for that.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getUserPermissions
func (s *SecurityService) GetUserPermissions(ctx context.Context, username string) (*[]Permission, *Response, error) {
	u := fmt.Sprintf("admin/permissions/user/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	request, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var getUserPermissionsResponse *userPermissionsResponse
	resp, err := s.client.Do(ctx, request, &getUserPermissionsResponse)
	if err != nil {
		return nil, resp, err
	}
	return &getUserPermissionsResponse.Permissions, resp, nil
}

// Get all permissions assigned to a given user as well as those granted by assigned roles
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getEffectiveUserPermissions
func (s *SecurityService) GetUserEffectivePermissions(ctx context.Context, username string) (*[]Permission, *Response, error) {
	u := fmt.Sprintf("admin/permissions/effective/user/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	request, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var getUserPermissionsResponse *userPermissionsResponse
	resp, err := s.client.Do(ctx, request, &getUserPermissionsResponse)
	if err != nil {
		return nil, resp, err
	}

	return &getUserPermissionsResponse.Permissions, resp, nil
}

// Get user attributes (enabled and superuser), roles, and effective permissions
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUser
func (s *SecurityService) GetUserDetails(ctx context.Context, username string) (*UserDetails, *Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}

	request, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var getUserDetailsResponse *UserDetails
	resp, err := s.client.Do(ctx, request, &getUserDetailsResponse)
	if err != nil {
		return nil, resp, err
	}

	return getUserDetailsResponse, resp, nil
}

// Is user a superuser
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/isSuper
func (s *SecurityService) IsSuperuser(ctx context.Context, username string) (*bool, *Response, error) {
	u := fmt.Sprintf("admin/users/%s/superuser", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	request, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var isSuperuserResponse *isSuperuserResponse
	resp, err := s.client.Do(ctx, request, &isSuperuserResponse)
	if err != nil {
		return nil, resp, err
	}

	return &isSuperuserResponse.Superuser, resp, nil
}

// Is user enabled or disabled
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/userEnabled
func (s *SecurityService) IsEnabled(ctx context.Context, username string) (*bool, *Response, error) {
	u := fmt.Sprintf("admin/users/%s/enabled", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	request, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var isEnabledResponse *isEnabledResponse
	resp, err := s.client.Do(ctx, request, &isEnabledResponse)
	if err != nil {
		return nil, resp, err
	}
	return &isEnabledResponse.Enabled, resp, nil
}

// Add a user to the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUser
func (s *SecurityService) CreateUser(ctx context.Context, username string, password string) (*Response, error) {
	u := "admin/users"

	credentials := credentials{
		Username: username,
		Password: strings.Split(password, ""),
	}
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	request, err := s.client.NewRequest("POST", u, &headerOpts, credentials)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, request, nil)
}

// Delete a user from the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/deleteUser
func (s *SecurityService) DeleteUser(ctx context.Context, username string) (*Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)
	request, err := s.client.NewRequest("DELETE", u, nil, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, request, nil)

}

// Change user's password
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/changePassword
func (s *SecurityService) ChangeUserPassword(ctx context.Context, username string, password string) (*Response, error) {
	u := fmt.Sprintf("admin/users/%s/pwd", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}

	reqBody := changePasswordRequest{
		Password: password,
	}
	request, err := s.client.NewRequest("PUT", u, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, request, nil)
}

// Enable/disable user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *SecurityService) EnableUser(ctx context.Context, username string, enabled bool) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/enabled", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	reqBody := enableRequest{
		Enabled: enabled,
	}

	req, err := s.client.NewRequest("PUT", url, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Grant a permission directly to a specified user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *SecurityService) GrantUserPermission(ctx context.Context, username string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/user/%s", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("PUT", url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Revoke a permission from a given user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteUserPermission
func (s *SecurityService) RevokeUserPermission(ctx context.Context, username string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/user/%s/delete", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("POST", url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Returns a list of all users that have a specific role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/getUsersWithRole
func (s *SecurityService) GetUsersAssignedRole(ctx context.Context, rolename string) ([]string, *Response, error) {
	u := fmt.Sprintf("admin/roles/%s/users", rolename)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var listUsersResponse *usersResponse
	resp, err := s.client.Do(ctx, req, &listUsersResponse)
	if err != nil {
		return nil, resp, err
	}
	return listUsersResponse.Users, resp, err
}

// Assign a role to user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUserRole
func (s *SecurityService) AssignRole(ctx context.Context, username string, rolename string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	reqBody := assignRoleRequest{
		Rolename: rolename,
	}
	req, err := s.client.NewRequest("POST", url, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Unassign role from a user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/removeUserRole
func (s *SecurityService) UnassignRole(ctx context.Context, username string, rolename string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles/%s", username, rolename)
	req, err := s.client.NewRequest("DELETE", url, nil, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Retrieve a list of all roles explicitly assigned to a user
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUserRoles
func (s *SecurityService) GetRolesAssignedToUser(ctx context.Context, username string) ([]string, *Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("GET", url, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var listRolesResponse *rolesResponse
	resp, err := s.client.Do(ctx, req, &listRolesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listRolesResponse.Roles, resp, err
}

// List the names of all roles in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/GetRoles/operation/listRoles
func (s *SecurityService) GetRoles(ctx context.Context) ([]string, *Response, error) {
	u := "admin/roles"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("GET", u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var listRolesResponse *rolesResponse
	resp, err := s.client.Do(ctx, req, &listRolesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listRolesResponse.Roles, resp, nil
}

// Create a role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/addRole
func (s *SecurityService) CreateRole(ctx context.Context, rolename string) (*Response, error) {
	u := "admin/roles"
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	reqBody := createRoleRequest{
		Rolename: rolename,
	}
	req, err := s.client.NewRequest("POST", u, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Get all permissions granted to a given role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getRolePermissions
func (s *SecurityService) GetRolePermissions(ctx context.Context, rolename string) (*[]Permission, *Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s", rolename)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("GET", url, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var rolePermissionsResponse *rolePermissionsResponse
	resp, err := s.client.Do(ctx, req, &rolePermissionsResponse)
	if err != nil {
		return nil, resp, err
	}
	return &rolePermissionsResponse.Permissions, resp, nil
}

// Grant a permission to role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *SecurityService) GrantRolePermission(ctx context.Context, rolename string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s", rolename)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("PUT", url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Revoke permission from specified role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteRolePermission
func (s *SecurityService) RevokeRolePermission(ctx context.Context, rolename string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s/delete", rolename)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("POST", url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// Delete a role
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/deleteRole
func (s *SecurityService) DeleteRole(ctx context.Context, rolename string, opts *DeleteRoleOptions) (*Response, error) {
	u := fmt.Sprintf("admin/roles/%s", rolename)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJson,
	}
	req, err := s.client.NewRequest("DELETE", urlWithOptions, &headerOpts, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
