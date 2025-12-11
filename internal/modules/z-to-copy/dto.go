package ztocopy

// CreateRequest is the request body for creating a resource
type CreateRequest struct {
	Name string `json:"name" binding:"required"`
}

// Response is the API response
type Response struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at,omitempty"`
}
