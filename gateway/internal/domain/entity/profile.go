package entity

import "time"

// Enums for ProfileType and VerificationStatus

type ProfileType string

const (
	ProfileTypePersonal ProfileType = "personal"
	ProfileTypeBusiness ProfileType = "business"
)

type VerificationStatus string

const (
	StatusDraft    VerificationStatus = "draft"
	StatusPending  VerificationStatus = "pending"
	StatusVerified VerificationStatus = "verified"
	StatusRejected VerificationStatus = "rejected"
)

// Value Objects

type LocationInfo struct {
	PostalCode  string `json:"postal_code"`
	Province    string `json:"province"`
	City        string `json:"city"`
	Address     string `json:"address"`
	PlateNumber string `json:"plate_number"`
	Unit        string `json:"unit"`
	FixedPhone  string `json:"fixed_phone"`
}

type BusinessMetaData struct {
	BrandName   string `json:"brand_name"`
	FieldOfWork string `json:"field_of_work"`
	WebsiteURL  string `json:"website_url"`
}

type PersonalDocuments struct {
	NationalCardFront string `json:"national_card_front"`
	NationalCardBack  string `json:"national_card_back"`
	IdBookPageOne     string `json:"id_book_page_one"`
}

type BusinessDocuments struct {
	EstablishmentNotice string `json:"establishment_notice"` // Agahi taasis
	Statutes            string `json:"statutes"`             // Asase nameh
	IntroductionLetter  string `json:"introduction_letter"`  // Nameh moarefi
	OfficialGazette     string `json:"official_gazette"`     // Rouznameh rasmi
}

// Database Entity

type Profile struct {
	ProfileID          uint64             `gorm:"primaryKey;autoIncrement" json:"profile_id"`
	UserID             uint64             `gorm:"not null;index" json:"user_id"`
	ProfileType        ProfileType        `gorm:"type:varchar(20);not null" json:"profile_type"`
	ProfileName        string             `gorm:"type:varchar(255);not null" json:"profile_name"`
	Balance            uint64             `gorm:"type:bigint;default:0;not null" json:"balance"`
	VerificationStatus VerificationStatus `gorm:"type:varchar(20);default:'draft';not null" json:"verification_status"`
	IsActive           bool               `gorm:"default:true;not null" json:"is_active"`
	CreatedAt          time.Time          `gorm:"not null;default:now()" json:"created_at"`

	PersonDetails   *ProfilePersonDetails   `gorm:"foreignKey:ProfileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"person_details,omitempty"`
	BusinessDetails *ProfileBusinessDetails `gorm:"foreignKey:ProfileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"business_details,omitempty"`
}

func (Profile) TableName() string {
	return "profiles"
}

type ProfilePersonDetails struct {
	ProfileID    uint64    `gorm:"primaryKey;not null" json:"profile_id"`
	FirstName    string    `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName     string    `gorm:"type:varchar(100);not null" json:"last_name"`
	NationalID   string    `gorm:"type:varchar(10);not null;unique" json:"national_code"`
	DOB          time.Time `gorm:"type:date;not null" json:"dob"`
	MobileNumber string    `gorm:"type:varchar(15);not null;unique" json:"mobile_number"`

	BusinessInfo *BusinessMetaData  `gorm:"type:jsonb;serializer:json" json:"business_info"`
	LocationInfo *LocationInfo      `gorm:"type:jsonb;serializer:json" json:"location_info"`
	Documents    *PersonalDocuments `gorm:"type:jsonb;serializer:json" json:"documents"`
}

func (ProfilePersonDetails) TableName() string {
	return "profile_person_details"
}

type ProfileBusinessDetails struct {
	ProfileID       uint64    `gorm:"primaryKey;not null" json:"profile_id"`
	RepFirstName    string    `gorm:"type:varchar(100);not null" json:"rep_first_name"`
	RepLastName     string    `gorm:"type:varchar(100);not null" json:"rep_last_name"`
	RepNationalID   string    `gorm:"type:varchar(10);not null;unique" json:"rep_national_code"`
	RepDOB          time.Time `gorm:"type:date;not null" json:"rep_dob"`
	RepMobileNumber string    `gorm:"type:varchar(15);not null;unique" json:"rep_mobile_number"`

	BusinessNationalID string `gorm:"type:varchar(15);not null;unique" json:"business_national_id"`

	BusinessInfo      *BusinessMetaData  `gorm:"type:jsonb;serializer:json" json:"business_info"`
	LocationInfo      *LocationInfo      `gorm:"type:jsonb;serializer:json" json:"location_info"`
	SupplementaryDocs *BusinessDocuments `gorm:"type:jsonb;serializer:json" json:"supplementary_docs"`

	Signatories []AuthorizedSignatory `gorm:"foreignKey:BusinessProfileID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;" json:"signatories,omitempty"`
}

type AuthorizedSignatory struct {
	SignatoryID       uint64 `gorm:"primaryKey;autoIncrement" json:"signatory_id"`
	BusinessProfileID uint64 `gorm:"not null;index" json:"business_profile_id"`

	FirstName    string     `gorm:"type:varchar(100);not null" json:"first_name"`
	LastName     string     `gorm:"type:varchar(100);not null" json:"last_name"`
	NationalID   string     `gorm:"type:varchar(10);not null;unique" json:"national_code"`
	DOB          *time.Time `gorm:"type:date;not null" json:"dob"`
	MobileNumber string     `gorm:"type:varchar(15);not null;unique" json:"mobile_number"`

	Documents *PersonalDocuments `gorm:"type:jsonb;serializer:json" json:"documents"`
}

func (AuthorizedSignatory) TableName() string {
	return "authorized_signatories"
}
