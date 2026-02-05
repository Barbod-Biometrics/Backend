package workflow_config

type SaveConfigRequest struct {
	Name                  string `json:"name" binding:"required"`
	Instruction           string `json:"instruction"`
	LivenessSentence      string `json:"liveness_sentence"`
	OCRAcceptanceRequired bool   `json:"ocr_acceptance_required"`
}

type UpdateConfigRequest struct {
	Name                  string `json:"name"`
	Instruction           string `json:"instruction"`
	LivenessSentence      string `json:"liveness_sentence"`
	OCRAcceptanceRequired bool   `json:"ocr_acceptance_required"`
}

type ConfigResponse struct {
	ID                    uint64 `json:"id"`
	ProfileID             uint64 `json:"profile_id"`
	Name                  string `json:"name"`
	Instruction           string `json:"instruction"`
	LivenessSentence      string `json:"liveness_sentence"`
	OCRAcceptanceRequired bool   `json:"ocr_acceptance_required"`
	CreatedAt             string `json:"created_at"`
	UpdatedAt             string `json:"updated_at"`
}
