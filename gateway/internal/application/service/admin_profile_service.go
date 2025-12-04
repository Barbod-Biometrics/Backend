package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/usecase"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
)

type AdminProfileService struct {
	profileRepo repository.ProfileRepository
}

func NewAdminProfileService(profileRepo repository.ProfileRepository) usecase.AdminProfileUsecase {
	return &AdminProfileService{
		profileRepo: profileRepo,
	}
}

func (s *AdminProfileService) ListProfiles(ctx context.Context, req admin.ListProfilesRequest) (*admin.PaginatedProfilesResponse, error) {
	// Build filter
	filter := repository.ProfileFilter{}
	if req.Status != "" {
		status := entity.VerificationStatus(req.Status)
		filter.Status = &status
	}
	if req.ProfileType != "" {
		profileType := entity.ProfileType(req.ProfileType)
		filter.ProfileType = &profileType
	}
	if req.Search != "" {
		filter.SearchQuery = &req.Search
	}

	// Build sort
	sort := repository.ProfileSort{
		Field: req.SortBy,
		Order: req.SortOrder,
	}
	if sort.Field == "" {
		sort.Field = "created_at"
	}
	if sort.Order == "" {
		sort.Order = "desc"
	}

	// Build pagination
	pagination := repository.Pagination{
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}

	// Execute query
	result, err := s.profileRepo.ListWithFilters(ctx, filter, sort, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to list profiles: %w", err)
	}

	// Map to response
	items := make([]admin.ProfileListItem, 0, len(result.Profiles))
	for _, p := range result.Profiles {
		item := admin.ProfileListItem{
			ID:                 p.ProfileID,
			UserID:             p.UserID,
			ProfileType:        string(p.ProfileType),
			ProfileName:        p.ProfileName,
			VerificationStatus: string(p.VerificationStatus),
			IsActive:           p.IsActive,
			CreatedAt:          p.CreatedAt,
		}

		// Add summary fields based on profile type
		if p.ProfileType == entity.ProfileTypePersonal && p.PersonDetails != nil {
			item.OwnerName = p.PersonDetails.FirstName + " " + p.PersonDetails.LastName
			item.NationalID = p.PersonDetails.NationalID
			item.MobileNumber = p.PersonDetails.MobileNumber
		} else if p.ProfileType == entity.ProfileTypeBusiness && p.BusinessDetails != nil {
			item.OwnerName = p.BusinessDetails.RepFirstName + " " + p.BusinessDetails.RepLastName
			item.NationalID = p.BusinessDetails.RepNationalID
			item.MobileNumber = p.BusinessDetails.RepMobileNumber
		}

		items = append(items, item)
	}

	return &admin.PaginatedProfilesResponse{
		Profiles:   items,
		TotalCount: result.TotalCount,
		Page:       result.Page,
		PageSize:   result.PageSize,
		TotalPages: result.TotalPages,
	}, nil
}

func (s *AdminProfileService) GetProfileDetail(ctx context.Context, profileID uint64) (*admin.ProfileDetailResponse, error) {
	p, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return nil, fmt.Errorf("profile not found: %w", err)
	}

	resp := &admin.ProfileDetailResponse{
		ID:                 p.ProfileID,
		UserID:             p.UserID,
		ProfileType:        string(p.ProfileType),
		ProfileName:        p.ProfileName,
		Balance:            p.Balance,
		VerificationStatus: string(p.VerificationStatus),
		IsActive:           p.IsActive,
		CreatedAt:          p.CreatedAt,
	}

	// Map person details
	if p.PersonDetails != nil {
		dob := p.PersonDetails.DOB
		resp.PersonDetails = &profile.PersonDetailsResponse{
			FirstName:    p.PersonDetails.FirstName,
			LastName:     p.PersonDetails.LastName,
			NationalID:   p.PersonDetails.NationalID,
			DOB:          &dob,
			MobileNumber: p.PersonDetails.MobileNumber,
		}
		if p.PersonDetails.BusinessInfo != nil {
			resp.PersonDetails.BusinessInfo = &profile.BusinessMetaDTO{
				BrandName:   p.PersonDetails.BusinessInfo.BrandName,
				FieldOfWork: p.PersonDetails.BusinessInfo.FieldOfWork,
				WebsiteURL:  p.PersonDetails.BusinessInfo.WebsiteURL,
			}
		}
		if p.PersonDetails.LocationInfo != nil {
			resp.PersonDetails.LocationInfo = &profile.LocationDTO{
				PostalCode:  p.PersonDetails.LocationInfo.PostalCode,
				Province:    p.PersonDetails.LocationInfo.Province,
				City:        p.PersonDetails.LocationInfo.City,
				Address:     p.PersonDetails.LocationInfo.Address,
				PlateNumber: p.PersonDetails.LocationInfo.PlateNumber,
				Unit:        p.PersonDetails.LocationInfo.Unit,
				FixedPhone:  p.PersonDetails.LocationInfo.FixedPhone,
			}
		}
		if p.PersonDetails.Documents != nil {
			resp.PersonDetails.Documents = &profile.PersonalDocumentsDTO{
				NationalCardFront: p.PersonDetails.Documents.NationalCardFront,
				NationalCardBack:  p.PersonDetails.Documents.NationalCardBack,
				IdBookPageOne:     p.PersonDetails.Documents.IdBookPageOne,
			}
		}
	}

	// Map business details
	if p.BusinessDetails != nil {
		dob := p.BusinessDetails.RepDOB
		resp.BusinessDetails = &profile.BusinessDetailsResponse{
			RepFirstName:       p.BusinessDetails.RepFirstName,
			RepLastName:        p.BusinessDetails.RepLastName,
			RepNationalID:      p.BusinessDetails.RepNationalID,
			RepDOB:             &dob,
			RepMobileNumber:    p.BusinessDetails.RepMobileNumber,
			BusinessNationalID: p.BusinessDetails.BusinessNationalID,
		}
		if p.BusinessDetails.BusinessInfo != nil {
			resp.BusinessDetails.BusinessInfo = &profile.BusinessMetaDTO{
				BrandName:   p.BusinessDetails.BusinessInfo.BrandName,
				FieldOfWork: p.BusinessDetails.BusinessInfo.FieldOfWork,
				WebsiteURL:  p.BusinessDetails.BusinessInfo.WebsiteURL,
			}
		}
		if p.BusinessDetails.LocationInfo != nil {
			resp.BusinessDetails.LocationInfo = &profile.LocationDTO{
				PostalCode:  p.BusinessDetails.LocationInfo.PostalCode,
				Province:    p.BusinessDetails.LocationInfo.Province,
				City:        p.BusinessDetails.LocationInfo.City,
				Address:     p.BusinessDetails.LocationInfo.Address,
				PlateNumber: p.BusinessDetails.LocationInfo.PlateNumber,
				Unit:        p.BusinessDetails.LocationInfo.Unit,
				FixedPhone:  p.BusinessDetails.LocationInfo.FixedPhone,
			}
		}
		// Map signatories
		for _, sig := range p.BusinessDetails.Signatories {
			sigResp := profile.SignatoryResponse{
				SignatoryID:  sig.SignatoryID,
				FirstName:    sig.FirstName,
				LastName:     sig.LastName,
				NationalID:   sig.NationalID,
				DOB:          sig.DOB,
				MobileNumber: sig.MobileNumber,
			}
			if sig.Documents != nil {
				sigResp.Documents = &profile.PersonalDocumentsDTO{
					NationalCardFront: sig.Documents.NationalCardFront,
					NationalCardBack:  sig.Documents.NationalCardBack,
					IdBookPageOne:     sig.Documents.IdBookPageOne,
				}
			}
			resp.BusinessDetails.Signatories = append(resp.BusinessDetails.Signatories, sigResp)
		}
	}

	return resp, nil
}

func (s *AdminProfileService) ApproveProfile(ctx context.Context, profileID uint64, req admin.ApproveProfileRequest) error {
	p, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	if p.VerificationStatus != entity.StatusPending {
		return errors.New("only pending profiles can be approved")
	}

	p.VerificationStatus = entity.StatusVerified

	if err := s.profileRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("failed to approve profile: %w", err)
	}

	return nil
}

func (s *AdminProfileService) RejectProfile(ctx context.Context, profileID uint64, req admin.RejectProfileRequest) error {
	p, err := s.profileRepo.GetByID(ctx, profileID)
	if err != nil {
		return fmt.Errorf("profile not found: %w", err)
	}

	if p.VerificationStatus != entity.StatusPending {
		return errors.New("only pending profiles can be rejected")
	}

	p.VerificationStatus = entity.StatusRejected
	// Note: In a production system, you might want to store the rejection reason
	// in a separate table or field for audit purposes

	if err := s.profileRepo.Update(ctx, p); err != nil {
		return fmt.Errorf("failed to reject profile: %w", err)
	}

	return nil
}
