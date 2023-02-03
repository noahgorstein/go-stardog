package stardog

import (
	"context"
	"fmt"
	"net/http"
)

// RoleService handles communication with the role related methods of the Stardog API.
type RoleService service

// Role represents a a Stardog role that can be assigned to a user to implicitly assign permissions
type Role struct {
	// name of the role
	Name string `json:"rolename"`
	// permissions assigned to the role
	Permissions []Permission `json:"permissions"`
}

// response for ListNames
type listRoleNamesResponse struct {
	Roles []string `json:"roles"`
}

// response for List
type listRolesResponse struct {
	Roles []Role `json:"roles"`
}

// request for Create
type createRoleRequest struct {
	Rolename string `json:"rolename"`
}

// response for Permissions
type rolePermissionsResponse struct {
	Permissions []Permission `json:"permissions"`
}

// DeleteRoleOptions specifies the optional parameters to the [RoleService.Delete] method.
type DeleteRoleOptions struct {
	// remove the role if currently assigned to users
	Force bool `url:"force"`
}

// ListNames returns the names of all roles in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/GetRoles/operation/listRoles
func (s *RoleService) ListNames(ctx context.Context) ([]string, *Response, error) {
	u := "admin/roles"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var listRolesResponse listRoleNamesResponse
	resp, err := s.client.Do(ctx, req, &listRolesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listRolesResponse.Roles, resp, nil
}

// List returns all Roles in the system
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/listRolesDetailed
func (s *RoleService) List(ctx context.Context) ([]Role, *Response, error) {
	u := "admin/roles/list"
	headerOpts := requestHeaderOptions{
		Accept: mediaTypeApplicationJSON,
	}
	req, err := s.client.NewRequest(http.MethodGet, u, &headerOpts, nil)
	if err != nil {
		return nil, nil, err
	}
	var listRolesResponse listRolesResponse
	resp, err := s.client.Do(ctx, req, &listRolesResponse)
	if err != nil {
		return nil, resp, err
	}
	return listRolesResponse.Roles, resp, nil
}

// Create adds a role to the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/addRole
func (s *RoleService) Create(ctx context.Context, rolename string) (*Response, error) {
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

// Permissions returns the permissions assigned to a role.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/getRolePermissions
func (s *RoleService) Permissions(ctx context.Context, rolename string) ([]Permission, *Response, error) {
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

// GrantPermission grants a permission to a role.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/addUserPermission
func (s *RoleService) GrantPermission(ctx context.Context, rolename string, permission Permission) (*Response, error) {
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

// RevokePermission revokes a permission from a role.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Permissions/operation/deleteRolePermission
func (s *RoleService) RevokePermission(ctx context.Context, rolename string, permission Permission) (*Response, error) {
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

// Delete deletes the role from the system.
//
// Stardog API: https://stardog-union.github.io/http-docs/#tag/Roles/operation/deleteRole
func (s *RoleService) Delete(ctx context.Context, rolename string, opts *DeleteRoleOptions) (*Response, error) {
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
