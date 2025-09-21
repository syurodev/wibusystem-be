package dto

// CreateTenantRequest represents the request to create a new tenant
type CreateTenantRequest struct {
	Name string `json:"name" validate:"required,max=150"`
}
