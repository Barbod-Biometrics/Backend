package profile

import "time"

type LocationDTO struct {
	PostalCode  string `json:"postal_code"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Address     string `json:"address"`
	PlateNumber string `json:"plate_number"`
	Unit        string `json:"unit"`
	FixedPhone  string `json:"fixed_phone"`
}

type BusinessMetaDTO struct {
	BrandName   string `json:"brand_name"`
	FieldOfWork string `json:"field_of_work"`
	WebsiteURL  string `json:"website_url"`
}

type PersonalDocumentsDTO struct {
	NationalCardFront string `json:"national_card_front"`
	NationalCardBack  string `json:"national_card_back"`
	IdBookPageOne     string `json:"id_book_page_one"`
}

type CreateProfileRequest struct {
	ProfileType string `json:"profile_type" binding:"required,oneof=personal business"`
	ProfileName string `json:"profile_name" binding:"required,min=3,max=100"`
}

type UpdateProfileRequest struct {
	ProfileName *string `json:"profile_name" binding:"omitempty,min=3,max=100"`

	PersonDetails   *PersonDetailsDTO   `json:"person_details,omitempty"`
	BusinessDetails *BusinessDetailsDTO `json:"business_details,omitempty"`
}

type PersonDetailsDTO struct {
	FirstName    *string `json:"first_name" binding:"required,min=2,max=50"`
	LastName     *string `json:"last_name" binding:"required,min=2,max=50"`
	NationalID   *string `json:"national_id" binding:"required,len=10,numeric"`
	DOB          *string `json:"dob" binding:"required,datetime=2006-01-02"`
	MobileNumber *string `json:"mobile_number" binding:"required,len=11,numeric"`

	BusinessInfo *BusinessMetaDTO      `json:"business_info,omitempty"`
	LocationInfo *LocationDTO          `json:"location_info,omitempty"`
	Documents    *PersonalDocumentsDTO `json:"documents,omitempty"`
}

type BusinessDetailsDTO struct {
	RepFirstName       *string `json:"rep_first_name" binding:"required,min=2,max=50"`
	RepLastName        *string `json:"rep_last_name" binding:"required,min=2,max=50"`
	RepNationalID      *string `json:"rep_national_id" binding:"required,len=10,numeric,omitempty"`
	RepDOB             *string `json:"rep_dob" binding:"required,datetime=2006-01-02,omitempty"`
	RepMobileNumber    *string `json:"rep_mobile_number" binding:"required,len=11,numeric,omitempty"`
	BusinessNationalID *string `json:"business_national_id" binding:"omitempty,len=11,numeric"`

	BusinessInfo *BusinessMetaDTO `json:"business_info,omitempty"`
	LocationInfo *LocationDTO     `json:"location_info,omitempty"`

	Signatories []SignatoryDTO `json:"signatories,omitempty"`
}

type SignatoryDTO struct {
	FirstName    string `json:"first_name" binding:"required,min=2,max=50"`
	LastName     string `json:"last_name" binding:"required,min=2,max=50"`
	NationalID   string `json:"national_id" binding:"required,len=10,numeric,omitempty"`
	DOB          string `json:"dob" binding:"required,datetime=2006-01-02,omitempty"`
	MobileNumber string `json:"mobile_number" binding:"required,len=11,numeric,omitempty"`

	Documents *PersonalDocumentsDTO `json:"documents,omitempty"`
}

type SaveDocumentRequest struct {
	DocumentType string `json:"document_type" binding:"required,oneof=national_card_front national_card_back id_book_page_one establishment_notice statutes introduction_letter official_gazette"`
	FileUrl      string `json:"file_url" binding:"required,url"`
}

type GetUploadUrlRequest struct {
	DocumentType  string `json:"document_type" binding:"required,oneof=national_card_front national_card_back id_book_page_one establishment_notice statutes introduction_letter official_gazette"`
	FileExtension string `json:"file_extension" binding:"required,oneof=.jpg .jpeg .png .pdf"`
}

type GetUsageSummaryRequest struct {
	ProfileID uint64 `json:"profile_id" form:"profile_id" binding:"required"`
	// for probable filters in the future
	FromDate *time.Time `json:"from_date,omitempty" query:"from_date"`
	ToDate   *time.Time `json:"to_date,omitempty" query:"to_date"`
}
