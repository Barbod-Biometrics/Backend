package contact

// SubmitContactSalesRequest represents the incoming payload for contact sales
// swagger:model SubmitContactSalesRequest
// The captcha_token should come from Google reCAPTCHA v2 client integration on the landing page form.
type SubmitContactSalesRequest struct {
	FirstName    string `json:"first_name" binding:"required,min=2,max=100"`
	LastName     string `json:"last_name" binding:"required,min=2,max=100"`
	Phone        string `json:"phone" binding:"required,len=11,numeric,startswith=09"`
	BusinessName string `json:"business_name" binding:"required,min=2,max=150"`
	Email        string `json:"email" binding:"omitempty,email,max=150"`
	Description  string `json:"description" binding:"required,min=5,max=2000"`
	CaptchaToken string `json:"captcha_token"`
}

// AdminListContactSalesRequest holds filters for admin listing
// swagger:model AdminListContactSalesRequest
type AdminListContactSalesRequest struct {
	Page      int    `form:"page" binding:"omitempty,min=1"`
	PageSize  int    `form:"page_size" binding:"omitempty,min=1,max=100"`
	Status    string `form:"status" binding:"omitempty,oneof=new read"`
	Search    string `form:"search" binding:"omitempty,max=150"`
	HasEmail  *bool  `form:"has_email"`
	FromDate  string `form:"from_date" binding:"omitempty,datetime=2006-01-02"` // YYYY-MM-DD
	ToDate    string `form:"to_date" binding:"omitempty,datetime=2006-01-02"`   // YYYY-MM-DD
	SortBy    string `form:"sort_by" binding:"omitempty,oneof=created_at status"`
	SortOrder string `form:"sort_order" binding:"omitempty,oneof=asc desc"`
}
