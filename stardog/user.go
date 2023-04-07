package stardog

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
)

// UserService handles communication with the user related methods of the Stardog API.
type UserService service

// Represents a Stardog user
type User struct {
	Username             *string               `json:"username,omitempty"`
	Enabled              bool                  `json:"enabled"`
	Superuser            bool                  `json:"superuser"`
	Roles                []string              `json:"roles"`
	EffectivePermissions []EffectivePermission `json:"permissions"`
}

// response for ListNames
type listUserNamesResponse struct {
	Users []string `json:"users"`
}

// response for List
type listUsersResponse struct {
	Users []User `json:"users"`
}

// response for Permissions
type userPermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// response for EffectivePermissions
type getUsersEffectivePermissionsResponse struct {
	EffectivePermissions []EffectivePermission `json:"permissions"`
}

// response for IsSuperuser
type isSuperuserResponse struct {
	Superuser bool `json:"superuser"`
}

// response for IsEnabled
type isEnabledResponse struct {
	Enabled bool `json:"enabled"`
}

// request for Create
type createUserRequest struct {
	Username string   `json:"username"`
	Password []string `json:"password"`
}

// request for ChangePassword
type changePasswordRequest struct {
	Password string `json:"password"`
}

// request for Enable/Disable
type enableUserRequest struct {
	Enabled bool `json:"enabled"`
}

// request for AssignRole
type assignRoleRequest struct {
	Rolename string `json:"rolename"`
}

// request for OverwriteRoles
type overwriteRolesRequest struct {
	Roles []string `json:"roles"`
}

// WhoAmI returns the authenticated user's username
func (s *UserService) WhoAmI(ctx context.Context) (*string, *Response, error) {
	u := "admin/status/whoami"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypePlainText,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var buf bytes.Buffer
	resp, err := s.client.Do(ctx, req, &buf)
	if err != nil {
		return nil, resp, err
	}
	username := buf.String()
	return &username, resp, nil
}

// ListNames returns the name of all users in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/GetUsers/operation/listUsers
func (s *UserService) ListNames(ctx context.Context) ([]string, *Response, error) {
	u := "admin/users"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var listUserNamesResponse listUserNamesResponse
	resp, err := s.client.Do(ctx, req, &listUserNamesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listUserNamesResponse.Users, resp, err
}

// List returns all Users in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/listUsersDetailed
func (s *UserService) List(ctx context.Context) ([]User, *Response, error) {
	u := "admin/users/list"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var userList listUsersResponse
	resp, err := s.client.Do(ctx, req, &userList)
	if err != nil {
		return nil, resp, err
	}
	return userList.Users, resp, err
}

// Permissions returns the permissions explicitly assigned to user. Permissions granted to a user via role assignment
// will not be contained in the response. Use [UserService.UserEffectivePermissions] for that.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getUserPermissions
func (s *UserService) Permissions(ctx context.Context, username string) ([]Permission, *Response, error) {
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

// EffectivePermissions returns permissions explicitly assigned to a user and via role assignment.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getEffectiveUserPermissions
func (s *UserService) EffectivePermissions(ctx context.Context, username string) ([]EffectivePermission, *Response, error) {
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

// Get returns a User in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUser
func (s *UserService) Get(ctx context.Context, username string) (*User, *Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}

	request, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var user User
	resp, err := s.client.Do(ctx, request, &user)
	if err != nil {
		return nil, resp, err
	}

	return &user, resp, nil
}

// IsSuperuser returns whether the user is a superuser or not
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/isSuper
func (s *UserService) IsSuperuser(ctx context.Context, username string) (*bool, *Response, error) {
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
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/userEnabled
func (s *UserService) IsEnabled(ctx context.Context, username string) (*bool, *Response, error) {
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

// Create adds a user to the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUser
func (s *UserService) Create(ctx context.Context, username string, password string) (*Response, error) {
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

// Delete deletes a user from the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/deleteUser
func (s *UserService) Delete(ctx context.Context, username string) (*Response, error) {
	u := fmt.Sprintf("admin/users/%s", username)
	request, err := s.client.NewRequest(http.MethodDelete, u, nil, nil)
	if err != nil {
		return nil, err
	}

	return s.client.Do(ctx, request, nil)
}

// ChangePassword changes a user's password.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/changePassword
func (s *UserService) ChangePassword(ctx context.Context, username string, password string) (*Response, error) {
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

// Enable enables a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *UserService) Enable(ctx context.Context, username string) (*Response, error) {
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

// Disable disables a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserEnabled
func (s *UserService) Disable(ctx context.Context, username string) (*Response, error) {
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

// GrantPermission grants a permission a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *UserService) GrantPermission(ctx context.Context, username string, permission Permission) (*Response, error) {
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

// RevokePermission revokes a permission from a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteUserPermission
func (s *UserService) RevokePermission(ctx context.Context, username string, permission Permission) (*Response, error) {
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

// ListNamesAssignedRole returns all the names of users assigned a given role.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/getUsersWithRole
func (s *UserService) ListNamesAssignedRole(ctx context.Context, rolename string) ([]string, *Response, error) {
	u := fmt.Sprintf("admin/roles/%s/users", rolename)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}

	var listUserNamesResponse listUserNamesResponse
	resp, err := s.client.Do(ctx, req, &listUserNamesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listUserNamesResponse.Users, resp, err
}

// AssignRole assigns a role to a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/addUserRole
func (s *UserService) AssignRole(ctx context.Context, username string, rolename string) (*Response, error) {
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
func (s *UserService) UnassignRole(ctx context.Context, username string, rolename string) (*Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles/%s", username, rolename)
	req, err := s.client.NewRequest(http.MethodDelete, url, nil, nil)
	if err != nil {
		return nil, err
	}
	return s.client.Do(ctx, req, nil)
}

// OverwriteRoles overwrites the the list roles assigned to a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/setUserRoles
func (s *UserService) OverwriteRoles(ctx context.Context, username string, roles []string) (*Response, error) {
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

// Roles returns the names of all roles assigned to a user.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Users/operation/getUserRoles
func (s *UserService) Roles(ctx context.Context, username string) ([]string, *Response, error) {
	url := fmt.Sprintf("admin/users/%s/roles", username)
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, url, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var listRolesResponse *listRoleNamesResponse
	resp, err := s.client.Do(ctx, req, &listRolesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listRolesResponse.Roles, resp, err
}
