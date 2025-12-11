package user

// CreateRequest is the request body for creating a user
type CreateRequest struct {
	Name  string `json:"name" binding:"required"`
	Email string `json:"email" binding:"required,email"`
}

// Response is the API response for a user
type Response struct {
	ID        int32  `json:"id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	CreatedAt string `json:"created_at,omitempty"`
}
