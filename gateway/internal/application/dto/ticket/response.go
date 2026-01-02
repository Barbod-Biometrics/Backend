package ticket

import "time"

// TicketListItemResponse represents a ticket in the list view (for users)
type TicketListItemResponse struct {
	ID        string `json:"id"`
	Service   string `json:"service"`
	Title     string `json:"title"`
	Status    string `json:"status"`
	CreatedAt string `json:"createdAt"` // Jalali date format: 1403/10/05
}

// TicketResponse represents a ticket response after creation
type TicketResponse struct {
	ID          string `json:"id"`
	Service     string `json:"service"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	Attachment  string `json:"attachment,omitempty"`
	CreatedAt   string `json:"createdAt"` // Jalali date format: 1403/10/05
}

// UserContactResponse represents user contact information
type UserContactResponse struct {
	Phone string `json:"phone,omitempty"`
	Email string `json:"email,omitempty"`
}

// AdminTicketListItemResponse represents a ticket in the list view (for admin)
type AdminTicketListItemResponse struct {
	ID            string               `json:"id"`
	Service       string               `json:"service"`
	Title         string               `json:"title"`
	Status        string               `json:"status"`
	CreatedAt     time.Time            `json:"createdAt"`
	UserContact   *UserContactResponse `json:"userContact"`
	LastMessageAt *time.Time           `json:"lastMessageAt,omitempty"`
}

// AdminTicketDetailResponse represents detailed ticket info for admin
type AdminTicketDetailResponse struct {
	ID          string               `json:"id"`
	UserID      string               `json:"userId"`
	UserContact *UserContactResponse `json:"userContact"`
	Service     string               `json:"service"`
	Title       string               `json:"title"`
	Message     string               `json:"message"`
	FileURL     string               `json:"fileUrl,omitempty"`
	Status      string               `json:"status"`
	CreatedAt   time.Time            `json:"createdAt"`
	UpdatedAt   time.Time            `json:"updatedAt"`
}

// FileURLResponse represents the file URL response
type FileURLResponse struct {
	URL string `json:"url"`
}

// MessageResponse represents a message in a ticket
type MessageResponse struct {
	ID        string    `json:"id"`
	TicketID  string    `json:"ticketId"`
	Sender    string    `json:"sender"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"createdAt"`
}

// CloseTicketResponse represents the response when closing a ticket
type CloseTicketResponse struct {
	TicketID string `json:"ticketId"`
	Status   string `json:"status"`
}

// SuccessResponse is a generic success response wrapper
type SuccessResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// SuccessListResponse is a success response for lists
type SuccessListResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}
