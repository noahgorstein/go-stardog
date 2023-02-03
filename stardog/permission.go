package stardog

import (
	"strings"
)

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

// EffectivePermission represents a permission assigned implicitly via role assignment or explicitly.
type EffectivePermission struct {
	Permission
	// whether the permission is explictly assigned to user or implicitly via role assignment
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
