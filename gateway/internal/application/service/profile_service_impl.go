package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/pkg/storage"
)

type ProfileService struct {
	profileRepo repository.ProfileRepository
	storage     *storage.MinioClient
}

func NewProfileService(profileRepo repository.ProfileRepository, storage *storage.MinioClient) usecase.ProfileUsecase {
	return &ProfileService{
		profileRepo: profileRepo,
		storage:     storage,
	}
}

func (s *ProfileService) CreateDraft(ctx context.Context, userID uint64, req profile.CreateProfileRequest) (*profile.ProfileResponse, error) {
	if req.ProfileType == string(entity.ProfileTypePersonal) {
		existing, err := s.profileRepo.GetPersonalProfileByUserID(ctx, userID)
		if err != nil {
			return nil, fmt.Errorf("failed to check existing profiles: %w", err)
		}
		if existing != nil {
			return nil, fmt.Errorf("personal profile already exists for user ID %d", userID)
		}
	}

	newProfile := &entity.Profile{
		UserID:             userID,
		ProfileType:        entity.ProfileType(req.ProfileType),
		ProfileName:        req.ProfileName,
		VerificationStatus: entity.StatusDraft,
		IsActive:           true,
		CreatedAt:          time.Now(),
	}

	// NOTE: We intentionally do NOT create PersonDetails/BusinessDetails here.
	// They will be created when the user provides actual data via UpdateDraft.
	// Creating them with empty values would violate unique constraints on NationalID/MobileNumber.

	if err := s.profileRepo.Create(ctx, newProfile); err != nil {
		return nil, fmt.Errorf("failed to create profile: %w", err)
	}

	return s.mapEntityToResponse(newProfile), nil
}

func (s *ProfileService) UpdateDraft(ctx context.Context, userID uint64, profileID uint64, req profile.UpdateProfileRequest) (*profile.ProfileResponse, error) {
	existing, err := s.getProfileForUser(ctx, userID, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile: %w", err)
	}

	if existing.VerificationStatus != entity.StatusDraft && existing.VerificationStatus != entity.StatusRejected {
		return nil, fmt.Errorf("cannot update a profile that is peding or verified")
	}

	if req.ProfileName != nil {
		existing.ProfileName = *req.ProfileName
	}

	if existing.ProfileType == entity.ProfileTypePersonal && req.PersonDetails != nil {
		// Create PersonDetails if it doesn't exist yet (draft was created without it)
		if existing.PersonDetails == nil {
			existing.PersonDetails = &entity.ProfilePersonDetails{
				ProfileID: existing.ProfileID,
			}
		}
		err = s.updatePersonalDetails(existing.PersonDetails, req.PersonDetails)
		if err != nil {
			return nil, fmt.Errorf("failed to update personal details: %w", err)
		}
	}

	if existing.ProfileType == entity.ProfileTypeBusiness && req.BusinessDetails != nil {
		// Create BusinessDetails if it doesn't exist yet (draft was created without it)
		if existing.BusinessDetails == nil {
			existing.BusinessDetails = &entity.ProfileBusinessDetails{
				ProfileID: existing.ProfileID,
			}
		}
		err = s.updateBusinessDetails(existing.BusinessDetails, req.BusinessDetails)
		if err != nil {
			return nil, fmt.Errorf("failed to update business details: %w", err)
		}
	}

	if err := s.profileRepo.Update(ctx, existing); err != nil {
		return nil, fmt.Errorf("failed to update profile: %w", err)
	}

	return s.mapEntityToResponse(existing), nil
}

func (s *ProfileService) mapEntityToResponse(e *entity.Profile) *profile.ProfileResponse {
	resp := &profile.ProfileResponse{
		ID:                 fmt.Sprintf("%d", e.ProfileID),
		Type:               string(e.ProfileType),
		Name:               e.ProfileName,
		Balance:            e.Balance,
		VerificationStatus: string(e.VerificationStatus),
		IsActive:           e.IsActive,
		CreatedAt:          e.CreatedAt,
	}

	if e.PersonDetails != nil {
		resp.PersonDetails = &profile.PersonDetailsResponse{
			FirstName:    e.PersonDetails.FirstName,
			LastName:     e.PersonDetails.LastName,
			NationalID:   e.PersonDetails.NationalID,
			DOB:          &e.PersonDetails.DOB,
			MobileNumber: e.PersonDetails.MobileNumber,
		}

		resp.PersonDetails.LocationInfo = s.mapLocationEntityToDTO(e.PersonDetails.LocationInfo)
		resp.PersonDetails.BusinessInfo = s.mapBusinessMetaEntityToDTO(e.PersonDetails.BusinessInfo)
	}

	if e.BusinessDetails != nil {
		resp.BusinessDetails = &profile.BusinessDetailsResponse{
			RepFirstName:       e.BusinessDetails.RepFirstName,
			RepLastName:        e.BusinessDetails.RepLastName,
			RepNationalID:      e.BusinessDetails.RepNationalID,
			RepDOB:             &e.BusinessDetails.RepDOB,
			RepMobileNumber:    e.BusinessDetails.RepMobileNumber,
			BusinessNationalID: e.BusinessDetails.BusinessNationalID,
		}

		resp.BusinessDetails.LocationInfo = s.mapLocationEntityToDTO(e.BusinessDetails.LocationInfo)
		resp.BusinessDetails.BusinessInfo = s.mapBusinessMetaEntityToDTO(e.BusinessDetails.BusinessInfo)

		for _, signatory := range e.BusinessDetails.Signatories {
			signatoryResp := profile.SignatoryResponse{
				SignatoryID:  signatory.SignatoryID,
				FirstName:    signatory.FirstName,
				LastName:     signatory.LastName,
				NationalID:   signatory.NationalID,
				DOB:          signatory.DOB,
				MobileNumber: signatory.MobileNumber,
			}
			resp.BusinessDetails.Signatories = append(resp.BusinessDetails.Signatories, signatoryResp)
		}
	}

	return resp
}

func (s *ProfileService) mapLocationEntityToDTO(e *entity.LocationInfo) *profile.LocationDTO {
	if e == nil {
		return nil
	}
	return &profile.LocationDTO{
		PostalCode:  e.PostalCode,
		Province:    e.Province,
		City:        e.City,
		Address:     e.Address,
		PlateNumber: e.PlateNumber,
		Unit:        e.Unit,
		FixedPhone:  e.FixedPhone,
	}
}

func (s *ProfileService) mapBusinessMetaEntityToDTO(e *entity.BusinessMetaData) *profile.BusinessMetaDTO {
	if e == nil {
		return nil
	}
	return &profile.BusinessMetaDTO{
		BrandName:   e.BrandName,
		FieldOfWork: e.FieldOfWork,
		WebsiteURL:  e.WebsiteURL,
	}
}

func (s *ProfileService) getProfileForUser(ctx context.Context, userID uint64, profileID uint64) (*entity.Profile, error) {
	existing, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("failed to get profile by ID: %w", err)
	}
	if existing == nil || existing.UserID != userID {
		return nil, fmt.Errorf("profile not found for user ID %d and profile ID %d", userID, profileID)
	}
	return existing, nil
}

func (s *ProfileService) updatePersonalDetails(target *entity.ProfilePersonDetails, source *profile.PersonDetailsDTO) error {
	if source.FirstName != nil {
		target.FirstName = *source.FirstName
	}
	if source.LastName != nil {
		target.LastName = *source.LastName
	}
	if source.NationalID != nil {
		target.NationalID = *source.NationalID
	}
	if source.DOB != nil {
		parsedTime, err := time.Parse("2006-01-02", *source.DOB)
		if err != nil {
			return fmt.Errorf("invalid date format for DOB: %w", err)
		}
		target.DOB = parsedTime
	}

	if source.MobileNumber != nil {
		target.MobileNumber = *source.MobileNumber
	}

	if source.LocationInfo != nil {
		target.LocationInfo = &entity.LocationInfo{
			PostalCode:  source.LocationInfo.PostalCode,
			Province:    source.LocationInfo.Province,
			City:        source.LocationInfo.City,
			Address:     source.LocationInfo.Address,
			PlateNumber: source.LocationInfo.PlateNumber,
			Unit:        source.LocationInfo.Unit,
			FixedPhone:  source.LocationInfo.FixedPhone,
		}
	}

	if source.BusinessInfo != nil {
		target.BusinessInfo = &entity.BusinessMetaData{
			BrandName:   source.BusinessInfo.BrandName,
			FieldOfWork: source.BusinessInfo.FieldOfWork,
			WebsiteURL:  source.BusinessInfo.WebsiteURL,
		}
	}

	return nil
}

func (s *ProfileService) updateBusinessDetails(target *entity.ProfileBusinessDetails, source *profile.BusinessDetailsDTO) error {
	if source.RepFirstName != nil {
		target.RepFirstName = *source.RepFirstName
	}
	if source.RepLastName != nil {
		target.RepLastName = *source.RepLastName
	}
	if source.RepNationalID != nil {
		target.RepNationalID = *source.RepNationalID
	}
	if source.RepDOB != nil {
		parsedTime, err := time.Parse("2006-01-02", *source.RepDOB)
		if err != nil {
			return fmt.Errorf("invalid date format for RepDOB: %w", err)
		}
		target.RepDOB = parsedTime
	}

	if source.RepMobileNumber != nil {
		target.RepMobileNumber = *source.RepMobileNumber
	}

	if source.BusinessNationalID != nil {
		target.BusinessNationalID = *source.BusinessNationalID
	}

	if source.BusinessInfo != nil {
		target.BusinessInfo = &entity.BusinessMetaData{
			BrandName:   source.BusinessInfo.BrandName,
			FieldOfWork: source.BusinessInfo.FieldOfWork,
			WebsiteURL:  source.BusinessInfo.WebsiteURL,
		}
	}

	if source.LocationInfo != nil {
		target.LocationInfo = &entity.LocationInfo{
			PostalCode:  source.LocationInfo.PostalCode,
			Province:    source.LocationInfo.Province,
			City:        source.LocationInfo.City,
			Address:     source.LocationInfo.Address,
			PlateNumber: source.LocationInfo.PlateNumber,
			Unit:        source.LocationInfo.Unit,
			FixedPhone:  source.LocationInfo.FixedPhone,
		}
	}

	if len(source.Signatories) > 0 {
		target.Signatories = make([]entity.AuthorizedSignatory, len(source.Signatories))
		for i, signatory := range source.Signatories {
			var dobTime *time.Time
			if signatory.DOB != "" {
				parsedTime, err := time.Parse("2006-01-02", signatory.DOB)
				if err != nil {
					return fmt.Errorf("invalid date format for signatory DOB: %w", err)
				}
				dobTime = &parsedTime
			}

			target.Signatories[i] = entity.AuthorizedSignatory{
				FirstName:    signatory.FirstName,
				LastName:     signatory.LastName,
				NationalID:   signatory.NationalID,
				DOB:          dobTime,
				MobileNumber: signatory.MobileNumber,
			}

			if signatory.Documents != nil {
				target.Signatories[i].Documents = &entity.PersonalDocuments{
					NationalCardFront: signatory.Documents.NationalCardFront,
					NationalCardBack:  signatory.Documents.NationalCardBack,
					IdBookPageOne:     signatory.Documents.IdBookPageOne,
				}
			}
		}
	}

	return nil
}

func (s *ProfileService) SaveDocument(ctx context.Context, userID uint64, profileID uint64, req profile.SaveDocumentRequest) error {
	existing, err := s.getProfileForUser(ctx, userID, profileID)
	if err != nil {
		return err
	}

	if existing.VerificationStatus != entity.StatusDraft && existing.VerificationStatus != entity.StatusRejected {
		return errors.New("cannot upload documents to a locked profile")
	}

	if existing.ProfileType == entity.ProfileTypePersonal {
		if existing.PersonDetails.Documents == nil {
			existing.PersonDetails.Documents = &entity.PersonalDocuments{}
		}
		switch req.DocumentType {
		case "national_card_front":
			existing.PersonDetails.Documents.NationalCardFront = req.FileUrl
		case "national_card_back":
			existing.PersonDetails.Documents.NationalCardBack = req.FileUrl
		case "id_book_page_one":
			existing.PersonDetails.Documents.IdBookPageOne = req.FileUrl
		default:
			return fmt.Errorf("invalid document type for personal profile: %s", req.DocumentType)
		}
	} else if existing.ProfileType == entity.ProfileTypeBusiness {
		if existing.BusinessDetails.SupplementaryDocs == nil {
			existing.BusinessDetails.SupplementaryDocs = &entity.BusinessDocuments{}
		}
		switch req.DocumentType {
		case "establishment_notice":
			existing.BusinessDetails.SupplementaryDocs.EstablishmentNotice = req.FileUrl
		case "statutes":
			existing.BusinessDetails.SupplementaryDocs.Statutes = req.FileUrl
		case "introduction_letter":
			existing.BusinessDetails.SupplementaryDocs.IntroductionLetter = req.FileUrl
		case "official_gazette":
			existing.BusinessDetails.SupplementaryDocs.OfficialGazette = req.FileUrl
		default:
			return fmt.Errorf("invalid document type for business profile: %s", req.DocumentType)
		}
	}

	return s.profileRepo.Update(ctx, existing)
}

func (s *ProfileService) SubmitProfile(ctx context.Context, userID uint64, profileID uint64) (*profile.ProfileResponse, error) {
	existing, err := s.getProfileForUser(ctx, userID, profileID)
	if err != nil {
		return nil, err
	}

	if err := s.validateProfileCompleteness(existing); err != nil {
		return nil, err
	}

	existing.VerificationStatus = entity.StatusPending

	if err := s.profileRepo.Update(ctx, existing); err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(existing), nil
}

func (s *ProfileService) GetByID(ctx context.Context, userID uint64, profileID uint64) (*profile.ProfileResponse, error) {
	existing, err := s.getProfileForUser(ctx, userID, profileID)
	if err != nil {
		return nil, err
	}

	return s.mapEntityToResponse(existing), nil
}

func (s *ProfileService) validateProfileCompleteness(p *entity.Profile) error {
	if p == nil {
		return fmt.Errorf("profile is nil")
	}

	switch p.ProfileType {
	case entity.ProfileTypePersonal:
		pd := p.PersonDetails
		if pd == nil {
			return fmt.Errorf("person details are required for personal profile")
		}
		if pd.FirstName == "" {
			return fmt.Errorf("first name is required for personal profile")
		}
		if pd.LastName == "" {
			return fmt.Errorf("last name is required for personal profile")
		}
		if pd.NationalID == "" {
			return fmt.Errorf("national ID is required for personal profile")
		}
		if pd.MobileNumber == "" {
			return fmt.Errorf("mobile number is required for personal profile")
		}
		if pd.DOB.IsZero() {
			return fmt.Errorf("date of birth is required for personal profile")
		}
		// Documents are optional for personal profile

	case entity.ProfileTypeBusiness:
		bd := p.BusinessDetails
		if bd == nil {
			return fmt.Errorf("business details are required for business profile")
		}
		if bd.RepFirstName == "" {
			return fmt.Errorf("representative first name is required for business profile")
		}
		if bd.RepLastName == "" {
			return fmt.Errorf("representative last name is required for business profile")
		}
		if bd.RepNationalID == "" {
			return fmt.Errorf("representative national ID is required for business profile")
		}
		if bd.RepMobileNumber == "" {
			return fmt.Errorf("representative mobile number is required for business profile")
		}
		if bd.RepDOB.IsZero() {
			return fmt.Errorf("representative date of birth is required for business profile")
		}
		// BusinessNationalID is optional
		// SupplementaryDocs are optional
		// Signatories are optional
		// Validate signatories only if provided (signatories are optional)
		for i, sgn := range bd.Signatories {
			idx := i + 1
			if sgn.FirstName == "" {
				return fmt.Errorf("first name is required for signatory %d", idx)
			}
			if sgn.LastName == "" {
				return fmt.Errorf("last name is required for signatory %d", idx)
			}
			if sgn.NationalID == "" {
				return fmt.Errorf("national ID is required for signatory %d", idx)
			}
			if sgn.MobileNumber == "" {
				return fmt.Errorf("mobile number is required for signatory %d", idx)
			}
			if sgn.DOB == nil || sgn.DOB.IsZero() {
				return fmt.Errorf("date of birth is required for signatory %d", idx)
			}
			// Documents are optional for signatories
		}

	default:
		return fmt.Errorf("unknown profile type: %s", p.ProfileType)
	}

	return nil
}

func (s *ProfileService) GetUploadUrl(ctx context.Context, userID uint64, req profile.GetUploadUrlRequest) (*profile.UploadUrlResponse, error) {
	timestamp := time.Now().Unix()
	objectName := fmt.Sprintf("user_%d/%s_%d%s", userID, req.DocumentType, timestamp, req.FileExtension)

	url, err := s.storage.GeneratePresignedUploadURL(ctx, objectName, 15*time.Minute)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return &profile.UploadUrlResponse{
		UploadUrl: url,
		FileKey:   objectName,
		ExpiresAt: time.Now().Add(15 * time.Minute).Format(time.RFC3339),
	}, nil
}
