package entity


type OCRRecord struct {
	Success bool   `json:"success"`
	Message string `json:"message"`

	NationalID     string `json:"شماره_ملی,omitempty"`
	FirstName      string `json:"نام,omitempty"`
	LastName       string `json:"نام_خانوادگی,omitempty"`
	FatherName     string `json:"نام_پدر,omitempty"`
	BirthDate      string `json:"تاریخ_تولد,omitempty"`
	ExpirationDate string `json:"پایان_اعتبار,omitempty"`

	Stats map[string]interface{} `json:"stats,omitempty" gorm:"type:jsonb"`
}
