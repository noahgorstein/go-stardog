package stardog

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// SecurityService handles communication with the security related methods of the Stardog API.
type SecurityService service


// UserDetails represents all details about a Stardog user
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

// DeleteRoleOptions specifies the optional parameters to the SecurityService.DeleteRole method.
type DeleteRoleOptions struct {
  // useful if you want to remove the role and it is currently assigned to users
	Force bool `url:"force"`
}

type assignRoleRequest struct {
	Rolename string `json:"rolename"`
}

// Action represents the action in a permission definition.
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

// ResourceType represents the resource type in a permission definition.
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

// NewPermission represents a new permission to be granted/revoked 
func NewPermission(action Action, resourceType ResourceType, resource []string) *Permission {
	permission := Permission{
		Action:       string(action),
		ResourceType: string(resourceType),
		Resource:     resource,
	}
	return &permission
}

// Permission represents a user/role permission. 
//
// Some read-only methods will return an optional 'explicit' field indicating if the permission 
// was explicitly granted to the user or if it is implicitly granted via a role. When granting/revoking 
// a permission you should not provide a value for 'explicit'. The NewPermission function is provided
// for when you need construct a permission to be granted/revoked.
type Permission struct {
	Action       string   `json:"action"`
	ResourceType string   `json:"resource_type"`
	Resource     []string `json:"resource"`
	Explicit     *bool    `json:"explicit,omitempty"`
}

// GetUsers returns the name of all users in the server 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/GetUsers/operation/listUsers
func (s *SecurityService) GetUsers(ctx context.Context) ([]string, *Response, error) {
	u := "admin/users"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// GetUserPermissions returns the permissions explicitly assigned to user. Permissions granted by a 
// role the user may be assigned will not be contained in the response. Use UserEffectivePermissions for that.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getUserPermissions
func (s *SecurityService) GetUserPermissions(ctx context.Context, username string) (*[]Permission, *Response, error) {
	u := fmt.Sprintf("admin/permissions/user/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// GetUserEffectivePermissions returns permissions assigned to a given user as well as those granted by assigned roles.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getEffectiveUserPermissions
func (s *SecurityService) GetUserEffectivePermissions(ctx context.Context, username string) (*[]Permission, *Response, error) {
	u := fmt.Sprintf("admin/permissions/effective/user/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// GetUserDetails returns user attributes (enabled and superuser), roles assigned to the user, and the user's
// effective permissions
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUser
func (s *SecurityService) GetUserDetails(ctx context.Context, username string) (*UserDetails, *Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// IsSuperuser returns whether a the user is a superuser or not
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/isSuper
func (s *SecurityService) IsSuperuser(ctx context.Context, username string) (*bool, *Response, error) {
	u := fmt.Sprintf("admin/users/%s/superuser", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// IsEnabled returns whether the user is enabled or not.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/userEnabled
func (s *SecurityService) IsEnabled(ctx context.Context, username string) (*bool, *Response, error) {
	u := fmt.Sprintf("admin/users/%s/enabled", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// CreateUser adds a user to the system. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUser
func (s *SecurityService) CreateUser(ctx context.Context, username string, password string) (*Response, error) {
	u := "admin/users"

	credentials := credentials{
		Username: username,
		Password: strings.Split(password, ""),
	}
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodPost, u, &headerOpts, credentials)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, request, nil)
}

// DeleteUser deletes a user from the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/deleteUser
func (s *SecurityService) DeleteUser(ctx context.Context, username string) (*Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)
	request, err := s.client.NewRequest(http.MethodDelete, u, nil, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, request, nil)
}

// ChangeUserPassword changes a user's password.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/changePassword
func (s *SecurityService) ChangeUserPassword(ctx context.Context, username string, password string) (*Response, error) {
	u := fmt.Sprintf("admin/users/%s/pwd", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}

	reqBody := changePasswordRequest{
		Password: password,
	}
	request, err := s.client.NewRequest(http.MethodPut, u, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, request, nil)
}

// EnableUser enables/disables a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *SecurityService) EnableUser(ctx context.Context, username string, enabled bool) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/enabled", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := enableRequest{
		Enabled: enabled,
	}

	req, err := s.client.NewRequest(http.MethodPut, url, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// GrantUserPermission grants a permission directly to a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *SecurityService) GrantUserPermission(ctx context.Context, username string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/user/%s", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodPut, url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// RevokeUserPermission revokes a permission from a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteUserPermission
func (s *SecurityService) RevokeUserPermission(ctx context.Context, username string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/user/%s/delete", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodPost, url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// GetUsersAssignedRole returns all the names of users assigned a given role.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/getUsersWithRole
func (s *SecurityService) GetUsersAssignedRole(ctx context.Context, rolename string) ([]string, *Response, error) {
	u := fmt.Sprintf("admin/roles/%s/users", rolename)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// AssignRole assigns a role to a user. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUserRole
func (s *SecurityService) AssignRole(ctx context.Context, username string, rolename string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := assignRoleRequest{
		Rolename: rolename,
	}
	req, err := s.client.NewRequest(http.MethodPost, url, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// UnassignRole unassigns a role from a user. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/removeUserRole
func (s *SecurityService) UnassignRole(ctx context.Context, username string, rolename string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles/%s", username, rolename)
	req, err := s.client.NewRequest(http.MethodDelete, url, nil, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// GetRolesAssignedToUser returns the names of all roles assigned to a user. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUserRoles
func (s *SecurityService) GetRolesAssignedToUser(ctx context.Context, username string) ([]string, *Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, url, &headerOpts, nil)
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

// GetRoles returns the names of all roles in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/GetRoles/operation/listRoles
func (s *SecurityService) GetRoles(ctx context.Context) ([]string, *Response, error) {
	u := "admin/roles"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
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

// CreateRole adds a role to the system. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/addRole
func (s *SecurityService) CreateRole(ctx context.Context, rolename string) (*Response, error) {
	u := "admin/roles"
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := createRoleRequest{
		Rolename: rolename,
	}
	req, err := s.client.NewRequest(http.MethodPost, u, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// GetRolePermissions returns the permissions assigned to a role. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getRolePermissions
func (s *SecurityService) GetRolePermissions(ctx context.Context, rolename string) (*[]Permission, *Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s", rolename)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, url, &headerOpts, nil)
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

// GrantRolePermission grants a permission to a role. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *SecurityService) GrantRolePermission(ctx context.Context, rolename string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s", rolename)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodPut, url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// RevokeRolePermission revokes a permission from a role. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteRolePermission
func (s *SecurityService) RevokeRolePermission(ctx context.Context, rolename string, permission Permission) (*Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s/delete", rolename)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodPost, url, &headerOpts, permission)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// DeleteRole deletes the role from the system. 
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/deleteRole
func (s *SecurityService) DeleteRole(ctx context.Context, rolename string, opts *DeleteRoleOptions) (*Response, error) {
	u := fmt.Sprintf("admin/roles/%s", rolename)
	urlWithOptions, err := addOptions(u, opts)
	if err != nil {
		return nil, err
	}
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodDelete, urlWithOptions, &headerOpts, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}
