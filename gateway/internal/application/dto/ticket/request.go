package ticket

// CreateTicketRequest represents the request to create a new ticket
type CreateTicketRequest struct {
	Service     string `json:"service" binding:"required,oneof=ocr face_auth liveness other"`
	Title       string `json:"title" binding:"required,min=3,max=255"`
	Description string `json:"description" binding:"required,min=10,max=5000"`
	Attachment  string `json:"attachment,omitempty"` // MinIO object key (optional)
}

// SendMessageRequest represents the request to send a message
type SendMessageRequest struct {
	Message string `json:"message" binding:"required,min=1,max=5000"`
}
