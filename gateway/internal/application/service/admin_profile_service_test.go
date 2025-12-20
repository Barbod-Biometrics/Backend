package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/admin"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/repository"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestListProfiles_Success(t *testing.T) {
	tctx := context.Background()
	mockRepo := mocks.NewMockProfileRepository(t)
	svc := NewAdminProfileService(mockRepo)

	now := time.Now()
	personal := &entity.Profile{
		ProfileID:          1,
		UserID:             100,
		ProfileType:        entity.ProfileTypePersonal,
		ProfileName:        "John Personal",
		VerificationStatus: entity.StatusPending,
		IsActive:           true,
		CreatedAt:          now,
		PersonDetails: &entity.ProfilePersonDetails{
			FirstName:    "John",
			LastName:     "Doe",
			NationalID:   "1234567890",
			MobileNumber: "09120000000",
			DOB:          time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}
	businessNationalID := "BN-999"
	business := &entity.Profile{
		ProfileID:          2,
		UserID:             101,
		ProfileType:        entity.ProfileTypeBusiness,
		ProfileName:        "ACME Inc",
		VerificationStatus: entity.StatusVerified,
		IsActive:           true,
		CreatedAt:          now,
		BusinessDetails: &entity.ProfileBusinessDetails{
			RepFirstName:       "Jane",
			RepLastName:        "Smith",
			RepNationalID:      "0987654321",
			RepMobileNumber:    "09121111111",
			RepDOB:             time.Date(1985, 12, 12, 0, 0, 0, 0, time.UTC),
			BusinessNationalID: &businessNationalID,
		},
	}

	paginated := &repository.PaginatedResult{
		Profiles:   []*entity.Profile{personal, business},
		TotalCount: 2,
		Page:       1,
		PageSize:   10,
		TotalPages: 1,
	}

	mockRepo.
		On("ListWithFilters", mock.Anything, mock.Anything, mock.MatchedBy(func(s repository.ProfileSort) bool { return s.Field == "created_at" && s.Order == "desc" }), mock.MatchedBy(func(p repository.Pagination) bool { return p.Page == 1 && p.PageSize == 10 })).
		Return(paginated, nil)

	resp, err := svc.ListProfiles(tctx, admin.ListProfilesRequest{})
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp.Profiles))

	// Check first item (personal)
	p1 := resp.Profiles[0]
	assert.Equal(t, uint64(1), p1.ID)
	assert.Equal(t, "personal", p1.ProfileType)
	assert.Equal(t, "John Doe", p1.OwnerName)
	assert.Equal(t, "1234567890", p1.NationalID)
	assert.Equal(t, "09120000000", p1.MobileNumber)

	// Check second item (business)
	p2 := resp.Profiles[1]
	assert.Equal(t, uint64(2), p2.ID)
	assert.Equal(t, "business", p2.ProfileType)
	assert.Equal(t, "Jane Smith", p2.OwnerName)
	assert.Equal(t, "0987654321", p2.NationalID)
	assert.Equal(t, "09121111111", p2.MobileNumber)

	mockRepo.AssertExpectations(t)
}

func TestGetProfileDetail_PersonalAndBusiness(t *testing.T) {
	tctx := context.Background()
	mockRepo := mocks.NewMockProfileRepository(t)
	svc := NewAdminProfileService(mockRepo)

	now := time.Now()
	// Personal profile case
	personal := &entity.Profile{
		ProfileID:          1,
		UserID:             200,
		ProfileType:        entity.ProfileTypePersonal,
		ProfileName:        "Alice",
		Balance:            5000,
		VerificationStatus: entity.StatusVerified,
		IsActive:           true,
		CreatedAt:          now,
		PersonDetails: &entity.ProfilePersonDetails{
			FirstName:    "Alice",
			LastName:     "Wonder",
			NationalID:   "2222222222",
			DOB:          time.Date(1992, 5, 10, 0, 0, 0, 0, time.UTC),
			MobileNumber: "09123333333",
			BusinessInfo: &entity.BusinessMetaData{BrandName: "AliceBrand", FieldOfWork: "Ecommerce", WebsiteURL: "https://example.com"},
			LocationInfo: &entity.LocationInfo{PostalCode: "12345", Province: "Tehran", City: "City", Address: "Somewhere"},
			Documents:    &entity.PersonalDocuments{NationalCardFront: "front.png", NationalCardBack: "back.png", IdBookPageOne: "id.png"},
		},
	}

	mockRepo.On("GetByID", mock.Anything, uint64(1)).Return(personal, nil)
	resp, err := svc.GetProfileDetail(tctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, uint64(1), resp.ID)
	assert.NotNil(t, resp.PersonDetails)
	assert.Equal(t, "AliceBrand", resp.PersonDetails.BusinessInfo.BrandName)
	assert.Equal(t, "front.png", resp.PersonDetails.Documents.NationalCardFront)

	// Business profile case
	businessNationalID := "BUS-123"
	signatoryDOB := time.Date(2000, 2, 2, 0, 0, 0, 0, time.UTC)
	business := &entity.Profile{
		ProfileID:          2,
		UserID:             201,
		ProfileType:        entity.ProfileTypeBusiness,
		ProfileName:        "BizCo",
		Balance:            7000,
		VerificationStatus: entity.StatusVerified,
		IsActive:           true,
		CreatedAt:          now,
		BusinessDetails: &entity.ProfileBusinessDetails{
			RepFirstName:       "Bob",
			RepLastName:        "Builder",
			RepNationalID:      "3333333333",
			RepMobileNumber:    "09125555555",
			RepDOB:             time.Date(1995, 3, 3, 0, 0, 0, 0, time.UTC),
			BusinessNationalID: &businessNationalID,
			BusinessInfo:       &entity.BusinessMetaData{BrandName: "BizCo", FieldOfWork: "Construction", WebsiteURL: "https://biz.example"},
			LocationInfo:       &entity.LocationInfo{PostalCode: "54321", Province: "Tehran", City: "Capital", Address: "HQ"},
			Signatories:        []entity.AuthorizedSignatory{{SignatoryID: 1, FirstName: "Sig", LastName: "One", NationalID: "4444444444", DOB: &signatoryDOB, MobileNumber: "09126666666", Documents: &entity.PersonalDocuments{NationalCardFront: "sfront.png"}}},
		},
	}
	mockRepo.On("GetByID", mock.Anything, uint64(2)).Return(business, nil)
	resp2, err := svc.GetProfileDetail(tctx, 2)
	assert.NoError(t, err)
	assert.NotNil(t, resp2)
	assert.Equal(t, uint64(2), resp2.ID)
	assert.NotNil(t, resp2.BusinessDetails)
	assert.Equal(t, "BizCo", resp2.BusinessDetails.BusinessInfo.BrandName)
	assert.Equal(t, 1, len(resp2.BusinessDetails.Signatories))

	mockRepo.AssertExpectations(t)
}

func TestApproveRejectProfileFlows(t *testing.T) {
	tctx := context.Background()
	mockRepo := mocks.NewMockProfileRepository(t)
	svc := NewAdminProfileService(mockRepo)

	// Approve success
	p := &entity.Profile{ProfileID: 10, VerificationStatus: entity.StatusPending}
	mockRepo.On("GetByID", mock.Anything, uint64(10)).Return(p, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Profile) bool { return p.VerificationStatus == entity.StatusVerified })).Return(nil)
	err := svc.ApproveProfile(tctx, 10, admin.ApproveProfileRequest{})
	assert.NoError(t, err)

	// Approve not pending -> error
	p2 := &entity.Profile{ProfileID: 11, VerificationStatus: entity.StatusVerified}
	mockRepo.On("GetByID", mock.Anything, uint64(11)).Return(p2, nil)
	err = svc.ApproveProfile(tctx, 11, admin.ApproveProfileRequest{})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("only pending profiles can be approved")) || err.Error() == "only pending profiles can be approved")

	// Reject success
	p3 := &entity.Profile{ProfileID: 20, VerificationStatus: entity.StatusPending}
	mockRepo.On("GetByID", mock.Anything, uint64(20)).Return(p3, nil)
	mockRepo.On("Update", mock.Anything, mock.MatchedBy(func(p *entity.Profile) bool { return p.VerificationStatus == entity.StatusRejected })).Return(nil)
	err = svc.RejectProfile(tctx, 20, admin.RejectProfileRequest{Reason: "Not valid"})
	assert.NoError(t, err)

	// Reject not pending -> error
	p4 := &entity.Profile{ProfileID: 21, VerificationStatus: entity.StatusVerified}
	mockRepo.On("GetByID", mock.Anything, uint64(21)).Return(p4, nil)
	err = svc.RejectProfile(tctx, 21, admin.RejectProfileRequest{Reason: "Nope"})
	assert.Error(t, err)
	assert.True(t, errors.Is(err, errors.New("only pending profiles can be rejected")) || err.Error() == "only pending profiles can be rejected")

	mockRepo.AssertExpectations(t)
}
