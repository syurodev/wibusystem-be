package response

// Deprecated: use StandardResponse with pagination fields in Meta.
type PaginatedResponse struct {
    Success    bool        `json:"success"`
    Data       interface{} `json:"data"`
    Pagination Pagination  `json:"pagination"`
}

// Pagination represents pagination metadata
type Pagination struct {
    Page       int   `json:"page"`
    PageSize   int   `json:"page_size"`
    TotalPages int   `json:"total_pages"`
    TotalItems int64 `json:"total_items"`
}
