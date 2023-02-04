package stardog

import (
	"testing"
)

func TestPermissionAction_Valid(t *testing.T) {
	r := PermissionAction(100)
	if r.Valid() {
		t.Errorf("should be an invalid PermissionAction")
	}
	if r.String() != PermissionActionUnknown.String() {
		t.Errorf("PermissionAction string value should be unknown")
	}
}

func TestPermissionResourceType_Valid(t *testing.T) {
	r := PermissionResourceType(100)
	if r.Valid() {
		t.Errorf("should be an invalid PermissionResourceType")
	}
	if r.String() != PermissionResourceTypeUnknown.String() {
		t.Errorf("PermissionResourceType string value should be unknown")
	}
}

func TestPermissionAction_UnmarshalText(t *testing.T) {
	r := PermissionActionWrite
	r.UnmarshalText([]byte("write"))
	if r != PermissionActionWrite {
		t.Error("should still be PermissionActionWrite")
	}
	r.UnmarshalText([]byte("trite"))
	if r.Valid() {
		t.Error("should be an invalid PermissionAction")
	}
}

func TestPermissionResourceType_UnmarshalText(t *testing.T) {
	r := PermissionResourceTypeDatabaseAdmin
	r.UnmarshalText([]byte("admin"))
	if r != PermissionResourceTypeDatabaseAdmin {
		t.Error("should still be PermissionResourceTypeDatabaseAdmin")
	}
	r.UnmarshalText([]byte("trite"))
	if r.Valid() {
		t.Error("should be an invalid PermissionResourceType")
	}
}
