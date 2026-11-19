package dto

// PaginationResponse contains metadata shared by paginated API responses.
type PaginationResponse struct {
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Total int64 `json:"total"`
}

// PaginatedResponse is the common data shape for endpoints that return a list.
type PaginatedResponse[T any] struct {
	Items      []T                `json:"items"`
	Pagination PaginationResponse `json:"pagination"`
}
