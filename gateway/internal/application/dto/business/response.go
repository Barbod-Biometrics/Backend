package business

type NewAPIKeyResponse struct {
	ProfileID uint64 `json:"profile_id"`
	APIKey    string `json:"api_key"`
}
