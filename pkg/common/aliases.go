package common

import (
    d "wibusystem/pkg/common/dto"
    m "wibusystem/pkg/common/model"
    r "wibusystem/pkg/common/response"
)

// Models
type User = m.User
type Tenant = m.Tenant
type Membership = m.Membership
type Permission = m.Permission
type Role = m.Role
type RolePermission = m.RolePermission
type RoleAssignment = m.RoleAssignment
type AuthType = m.AuthType
const (
    AuthTypePassword = m.AuthTypePassword
    AuthTypeOAuth    = m.AuthTypeOAuth
    AuthTypeOIDC     = m.AuthTypeOIDC
    AuthTypeSAML     = m.AuthTypeSAML
    AuthTypeWebAuthn = m.AuthTypeWebAuthn
    AuthTypeTOTP     = m.AuthTypeTOTP
    AuthTypePasskey  = m.AuthTypePasskey
)
type Credential = m.Credential
type Device = m.Device
type Session = m.Session
type OAuth2Client = m.OAuth2Client
type StringArray = m.StringArray

// DTOs
type CreateUserRequest = d.CreateUserRequest
type UpdateUserRequest = d.UpdateUserRequest
type CreateTenantRequest = d.CreateTenantRequest
type LoginRequest = d.LoginRequest
type LoginResponse = d.LoginResponse
type UserInfo = d.UserInfo

// Responses
type ErrorDetail = r.ErrorDetail
type StandardResponse = r.StandardResponse
// Deprecated passthroughs
type APIResponse = r.APIResponse
type ErrorResponse = r.ErrorResponse
type PaginatedResponse = r.PaginatedResponse
type Pagination = r.Pagination

