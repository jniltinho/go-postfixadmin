package models

import "time"

const (
	TableNameRBACRole      = "rbac_roles"
	TableNameRBACPermission = "rbac_permissions"
	TableNameRBACAdminRole = "rbac_admin_roles"
)

// RBACRole is a named set of permissions that can be assigned to admins.
// Roles with System=true are built-in and cannot be deleted or have their
// permissions changed through the API.
type RBACRole struct {
	ID          int              `gorm:"column:id;primaryKey;autoIncrement:true"                                                  json:"id"`
	Name        string           `gorm:"column:name;type:varchar(64);not null;uniqueIndex"                                        json:"name"`
	Description string           `gorm:"column:description;type:varchar(255);not null;default:''"                                json:"description"`
	System      bool             `gorm:"column:system;not null;default:false"                                                     json:"system"`
	Created     time.Time        `gorm:"column:created;not null;default:'2000-01-01 00:00:00'"                                   json:"created"`
	Modified    time.Time        `gorm:"column:modified;not null;autoUpdateTime"                                                  json:"modified"`
	Permissions []RBACPermission `gorm:"many2many:rbac_role_permissions;joinForeignKey:role_id;joinReferences:permission_id"      json:"permissions,omitempty"`
}

// TableName returns the table name for RBACRole.
func (RBACRole) TableName() string { return TableNameRBACRole }

// RBACPermission is an atomic capability in the format "<resource>:<action>".
// The Name field is the canonical string used in JWT claims and middleware checks.
type RBACPermission struct {
	ID       int    `gorm:"column:id;primaryKey;autoIncrement:true"               json:"id"`
	Name     string `gorm:"column:name;type:varchar(128);not null;uniqueIndex"    json:"name"`
	Resource string `gorm:"column:resource;type:varchar(64);not null"             json:"resource"`
	Action   string `gorm:"column:action;type:varchar(32);not null"               json:"action"`
}

// TableName returns the table name for RBACPermission.
func (RBACPermission) TableName() string { return TableNameRBACPermission }

// RBACAdminRole assigns a role to an admin, optionally scoped to a single domain.
// Domain="" means the role applies globally (all domains the admin manages).
// The composite unique index prevents duplicate (username, role_id, domain) triples.
type RBACAdminRole struct {
	ID       int       `gorm:"column:id;primaryKey;autoIncrement:true"                                                                       json:"id"`
	Username string    `gorm:"column:username;type:varchar(255);not null;uniqueIndex:uq_rbac_admin_role_domain,priority:1"                  json:"username"`
	RoleID   int       `gorm:"column:role_id;not null;uniqueIndex:uq_rbac_admin_role_domain,priority:2"                                      json:"role_id"`
	Domain   string    `gorm:"column:domain;type:varchar(255);not null;uniqueIndex:uq_rbac_admin_role_domain,priority:3;default:''"          json:"domain"`
	Created  time.Time `gorm:"column:created;not null;autoCreateTime"                                                                        json:"created"`
	Role     RBACRole  `gorm:"foreignKey:RoleID;constraint:OnDelete:CASCADE"                                                                json:"role,omitempty"`
}

// TableName returns the table name for RBACAdminRole.
func (RBACAdminRole) TableName() string { return TableNameRBACAdminRole }
