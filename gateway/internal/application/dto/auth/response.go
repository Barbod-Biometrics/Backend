package auth

type UserInfoResponse struct {
	AccessToken   string `json:"access_token"`
	RefereshToken string `json:"refresh_token"`
	PhoneNumber   string `json:"phone_number"`
	IsAdmin       bool   `json:"is_admin"`
}
