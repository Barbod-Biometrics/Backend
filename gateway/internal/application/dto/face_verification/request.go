package face_verification

type FaceVerificationRequest struct {
	Photo []byte `json:"photo" form:"photo"`
	Video []byte `json:"video" form:"video"`
}

type CropImageRequest struct {
	Image []byte `json:"image" form:"image"`
}

type HealthCheckRequest struct {
	// no fields
}
