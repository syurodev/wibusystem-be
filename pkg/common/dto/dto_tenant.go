package dto

// CreateTenantRequest represents the request to create a new tenant
type CreateTenantRequest struct {
	Name        string                 `json:"name" validate:"required,max=150"`
	Slug        string                 `json:"slug,omitempty" validate:"omitempty,max=50"`
	Description *string                `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}

// UpdateTenantRequest represents the request to update a tenant
type UpdateTenantRequest struct {
	Name        *string                `json:"name,omitempty" validate:"omitempty,max=150"`
	Slug        *string                `json:"slug,omitempty" validate:"omitempty,max=50"`
	Description *string                `json:"description,omitempty"`
	Settings    map[string]interface{} `json:"settings,omitempty"`
}
