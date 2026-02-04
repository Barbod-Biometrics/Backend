package contact

import "time"

// SubmitContactSalesResponse returned on successful submission
// swagger:model SubmitContactSalesResponse
type SubmitContactSalesResponse struct {
	ID      uint64 `json:"id"`
	Message string `json:"message"`
}

// ContactSalesItem used in admin list
// swagger:model ContactSalesItem
type ContactSalesItem struct {
	ID           uint64     `json:"id"`
	FirstName    string     `json:"first_name"`
	LastName     string     `json:"last_name"`
	Phone        string     `json:"phone"`
	BusinessName string     `json:"business_name"`
	Email        string     `json:"email,omitempty"`
	Description  string     `json:"description"`
	Status       string     `json:"status"`
	ReadAt       *time.Time `json:"read_at,omitempty"`
	CreatedAt    time.Time  `json:"created_at"`
}

// ContactSalesListResponse paginated admin list
// swagger:model ContactSalesListResponse
type ContactSalesListResponse struct {
	Items      []ContactSalesItem `json:"items"`
	TotalCount int64              `json:"total_count"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// ContactSalesActionResponse generic action response
// swagger:model ContactSalesActionResponse
type ContactSalesActionResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
