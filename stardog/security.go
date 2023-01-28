package stardog

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

// SecurityService handles communication with the security related methods of the Stardog API.
type SecurityService service

// UserDetails contains all details about a Stardog user
type UserDetails struct {
	Username             *string               `json:"username,omitempty"`
	Enabled              bool                  `json:"enabled"`
	Superuser            bool                  `json:"superuser"`
	Roles                []string              `json:"roles"`
	EffectivePermissions []EffectivePermission `json:"permissions"`
}

type getUsersResponse struct {
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

type getUsersEffectivePermissionsResponse struct {
	EffectivePermissions []EffectivePermission `json:"permissions"`
}

type createUserRequest struct {
	Username string   `json:"username"`
	Password []string `json:"password"`
}

type changeUserPasswordRequest struct {
	Password string `json:"password"`
}

type enableUserRequest struct {
	Enabled bool `json:"enabled"`
}

type rolesResponse struct {
	Roles []string `json:"roles"`
}

type getRolesWithDetailsResponse struct {
	Roles []RoleDetails `json:"roles"`
}

// RoleDetails contains details about a role
type RoleDetails struct {
	Rolename    string       `json:"rolename"`
	Permissions []Permission `json:"permissions"`
}

type createRoleRequest struct {
	Rolename string `json:"rolename"`
}

type rolePermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// DeleteRoleOptions specifies the optional parameters to the [SecurityService.DeleteRole] method.
type DeleteRoleOptions struct {
	// useful if you want to remove the role and it is currently assigned to users
	Force bool `url:"force"`
}

type assignRoleRequest struct {
	Rolename string `json:"rolename"`
}

type overwriteRolesRequest struct {
	Roles []string `json:"roles"`
}

type getUsersWithDetailsResponse struct {
	Users []UserDetails `json:"users"`
}

// Permission represents a user/role permission.
//
// Stardog [security model] states that a user/role can be perform an action (e.g. read)
// over a resource (e.g. db:myDatabase).
//
// [security model]: https://docs.stardog.com/operating-stardog/security/security-model#permissions
type Permission struct {
	// the access level (e.g. PermissionActionRead)
	Action PermissionAction `json:"action"`
	// the type of resource (e.g. PermissionResourceTypeDatabase)
	ResourceType PermissionResourceType `json:"resource_type"`
	// the resource identifier (e.g. myDatabase)
	Resource []string `json:"resource"`
}

// EffectivePermission represents a user
type EffectivePermission struct {
	Permission
	Explicit bool `json:"explicit"`
}

// PermissionAction represents the [action] in a Stardog permission.
// The zero value for a PermissionAction is [PermissionActionUnknown]
//
// [action]: https://docs.stardog.com/operating-stardog/security/security-model#actions
type PermissionAction int

// All available actions for a permission
const (
	PermissionActionUnknown PermissionAction = iota
	PermissionActionRead
	PermissionActionWrite
	PermissionActionCreate
	PermissionActionDelete
	PermissionActionGrant
	PermissionActionRevoke
	PermissionActionExecute
	PermissionActionAll
)

// permissionActionValues returns an array mapping each
// PermissionAction to its string value
//
//revive:disable:add-constant
func permissionActionValues() [9]string {
	return [9]string{
		PermissionActionUnknown: "",
		PermissionActionRead:    "read",
		PermissionActionWrite:   "write",
		PermissionActionCreate:  "create",
		PermissionActionDelete:  "delete",
		PermissionActionGrant:   "grant",
		PermissionActionRevoke:  "revoke",
		PermissionActionExecute: "execute",
		PermissionActionAll:     "all",
	}
}

//revive:enable:add-constant

// Valid returns if a given PermissionAction is known (valid) or not.
func (p PermissionAction) Valid() bool {
	return !(p <= PermissionActionUnknown || int(p) >= len(permissionActionValues()))
}

// String will return the string representation of the PermissionAction
func (p PermissionAction) String() string {
	if !p.Valid() {
		return permissionActionValues()[PermissionActionUnknown]
	}
	return permissionActionValues()[p]
}

// MarshalText implements TextMarshaler and is invoked when encoding the PermissionAction to JSON.
func (p PermissionAction) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements TextUnmarshaler and is invoked when decoding JSON to PermissionAction.
func (p *PermissionAction) UnmarshalText(text []byte) error {
	valsArr := permissionActionValues()
	valsSlice := valsArr[:]
	index := indexOf(valsSlice, strings.ToLower(string(text)))
	*p = PermissionAction(index)
	return nil
}

// PermissionResourceType represents the [resource type] in a Stardog permission.
// The zero value for a PermissionResourceType is [PermissionResourceTypeUnknown]
//
// [resource type]: https://docs.stardog.com/operating-stardog/security/security-model#resources
type PermissionResourceType int

// All available resource types for a permission
const (
	PermissionResourceTypeUnknown PermissionResourceType = iota
	PermissionResourceTypeDatabase
	PermissionResourceTypeMetadata
	PermissionResourceTypeUser
	PermissionResourceTypeRole
	PermissionResourceTypeNamedGraph
	PermissionResourceTypeVirtualGraph
	PermissionResourceTypeDataSource
	PermissionResourceTypeServeradmin
	PermissionResourceTypeDatabaseAdmin
	PermissionResourceTypeSensitiveProperty
	PermissionResourceTypeStoredQuery
	PermissionResourceTypeAll
)

// permissionResourceTypeValues returns an array mapping each
// PermissionResourceTypeAction to its string value
//
//revive:disable:add-constant
func permissionResourceTypeValues() [13]string {
	return [13]string{
		PermissionResourceTypeUnknown:           "UNKNOWN",
		PermissionResourceTypeDatabase:          "db",
		PermissionResourceTypeMetadata:          "metadata",
		PermissionResourceTypeUser:              "user",
		PermissionResourceTypeRole:              "role",
		PermissionResourceTypeNamedGraph:        "named-graph",
		PermissionResourceTypeVirtualGraph:      "virtual-graph",
		PermissionResourceTypeDataSource:        "data-source",
		PermissionResourceTypeServeradmin:       "dbms-admin",
		PermissionResourceTypeDatabaseAdmin:     "admin",
		PermissionResourceTypeSensitiveProperty: "sensitive-property",
		PermissionResourceTypeStoredQuery:       "stored-query",
		PermissionResourceTypeAll:               "*",
	}
}

//revive:enable:add-constant

// Valid returns if a given PermissionResourceType is known (valid) or not.
func (p PermissionResourceType) Valid() bool {
	return !(p <= PermissionResourceTypeUnknown || int(p) >= len(permissionResourceTypeValues()))
}

// String will return the string representation of the PermissionResourceType
func (p PermissionResourceType) String() string {
	if !p.Valid() {
		return permissionResourceTypeValues()[PermissionResourceTypeUnknown]
	}
	return permissionResourceTypeValues()[p]
}

// MarshalText implements TextMarshaler and is invoked when encoding the PermissionResourceType to JSON.
func (p PermissionResourceType) MarshalText() ([]byte, error) {
	return []byte(p.String()), nil
}

// UnmarshalText implements TextUnmarshaler and is invoked when decoding JSON to PermissionResourceType.
func (p *PermissionResourceType) UnmarshalText(text []byte) error {
	valsArr := permissionResourceTypeValues()
	valsSlice := valsArr[:]
	index := indexOf(valsSlice, strings.ToLower(string(text)))
	*p = PermissionResourceType(index)
	return nil
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

	var usersResponse getUsersResponse
	resp, err := s.client.Do(ctx, req, &usersResponse)
	if err != nil {
		return nil, resp, err
	}
	return usersResponse.Users, resp, err
}

// GetUsersWithDetails returns a list of all users in the system with
// details (username, enabled status, superuser status, permissions, and roles)
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/listUsersDetailed
func (s *SecurityService) GetUsersWithDetails(ctx context.Context) ([]UserDetails, *Response, error) {
	u := "admin/users/list"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var userListWithDetails getUsersWithDetailsResponse
	resp, err := s.client.Do(ctx, req, &userListWithDetails)
	if err != nil {
		return nil, resp, err
	}
	return userListWithDetails.Users, resp, err
}

// GetUserPermissions returns the permissions explicitly assigned to user. Permissions granted to a user via role assignment
// will not be contained in the response. Use [SecurityService.GetUserEffectivePermissions] for that.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getUserPermissions
func (s *SecurityService) GetUserPermissions(ctx context.Context, username string) ([]Permission, *Response, error) {
	u := fmt.Sprintf("admin/permissions/user/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var getUserPermissionsResponse userPermissionsResponse
	resp, err := s.client.Do(ctx, request, &getUserPermissionsResponse)
	if err != nil {
		return nil, resp, err
	}
	return getUserPermissionsResponse.Permissions, resp, nil
}

// GetUserEffectivePermissions returns permissions explicitly assigned to a user and via role assignment.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getEffectiveUserPermissions
func (s *SecurityService) GetUserEffectivePermissions(ctx context.Context, username string) ([]EffectivePermission, *Response, error) {
	u := fmt.Sprintf("admin/permissions/effective/user/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var getUsersEffectivePermissionsResponse getUsersEffectivePermissionsResponse
	resp, err := s.client.Do(ctx, request, &getUsersEffectivePermissionsResponse)
	if err != nil {
		return nil, resp, err
	}

	return getUsersEffectivePermissionsResponse.EffectivePermissions, resp, nil
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

	var getUserDetailsResponse UserDetails
	resp, err := s.client.Do(ctx, request, &getUserDetailsResponse)
	if err != nil {
		return nil, resp, err
	}

	return &getUserDetailsResponse, resp, nil
}

// IsSuperuser returns whether the user is a superuser or not
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

	var isSuperuserResponse isSuperuserResponse
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

	var isEnabledResponse isEnabledResponse
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

	credentials := createUserRequest{
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

	reqBody := changeUserPasswordRequest{
		Password: password,
	}
	request, err := s.client.NewRequest(http.MethodPut, u, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, request, nil)
}

// EnableUser enables a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *SecurityService) EnableUser(ctx context.Context, username string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/enabled", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := enableUserRequest{
		Enabled: true,
	}

	req, err := s.client.NewRequest(http.MethodPut, url, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// DisableUser disables a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *SecurityService) DisableUser(ctx context.Context, username string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/enabled", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := enableUserRequest{
		Enabled: false,
	}

	req, err := s.client.NewRequest(http.MethodPut, url, &headerOpts, reqBody)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// GrantUserPermission grants a permission a user.
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

	var listUsersResponse getUsersResponse
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

// OverwriteRoles overwrites the the list roles assigned to a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserRoles
func (s *SecurityService) OverwriteRoles(ctx context.Context, username string, roles []string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles", username)
	headerOpts := requestHeaderOptions{
		ContentType: mediaTypeApplicationJSON,
	}
	reqBody := overwriteRolesRequest{
		Roles: roles,
	}
	req, err := s.client.NewRequest(http.MethodPut, url, &headerOpts, reqBody)
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

// GetRolesWithPermissions returns a list of roles in the system with their permissions
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/listRolesDetailed
func (s *SecurityService) GetRolesWithDetails(ctx context.Context) ([]RoleDetails, *Response, error) {
	u := "admin/roles/list"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var getRolesWithDetailsResponse getRolesWithDetailsResponse
	resp, err := s.client.Do(ctx, req, &getRolesWithDetailsResponse)
	if err != nil {
		return nil, resp, err
	}
	return getRolesWithDetailsResponse.Roles, resp, nil
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
func (s *SecurityService) GetRolePermissions(ctx context.Context, rolename string) ([]Permission, *Response, error) {
	url := fmt.Sprintf("admin/permissions/role/%s", rolename)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, url, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var rolePermissionsResponse rolePermissionsResponse
	resp, err := s.client.Do(ctx, req, &rolePermissionsResponse)
	if err != nil {
		return nil, resp, err
	}
	return rolePermissionsResponse.Permissions, resp, nil
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
