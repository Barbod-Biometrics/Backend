package service

import (
	"context"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/profile"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateDraft_Personal_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	// No existing personal profile
	repo.On("GetPersonalProfileByUserID", mock.Anything, uint64(1)).Return((*entity.Profile)(nil), nil)

	// Create called; we capture the profile
	repo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		p := a.Get(1).(*entity.Profile)
		p.ProfileID = 123
	}).Return(nil)

	req := profile.CreateProfileRequest{ProfileType: string(entity.ProfileTypePersonal), ProfileName: "My Personal"}
	resp, err := svc.CreateDraft(ctx, 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "123", resp.ID)
	assert.Equal(t, "personal", resp.Type)
	assert.Equal(t, "My Personal", resp.Name)

	repo.AssertExpectations(t)
}

func TestCreateDraft_Personal_AlreadyExists(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	existing := &entity.Profile{ProfileID: 99, UserID: 1, ProfileType: entity.ProfileTypePersonal}
	repo.On("GetPersonalProfileByUserID", mock.Anything, uint64(1)).Return(existing, nil)

	req := profile.CreateProfileRequest{ProfileType: string(entity.ProfileTypePersonal), ProfileName: "X"}
	resp, err := svc.CreateDraft(ctx, 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "personal profile already exists")

	repo.AssertExpectations(t)
}

func TestUpdateDraft_Personal_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	now := time.Now()
	existing := &entity.Profile{ProfileID: 200, UserID: 2, ProfileType: entity.ProfileTypePersonal, ProfileName: "Old Name", VerificationStatus: entity.StatusDraft, CreatedAt: now}
	// existing had no PersonDetails; update will create it
	repo.On("GetByID", mock.Anything, uint64(200)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	fname := "John"
	lname := "Doe"
	nid := "1234567890"
	dob := "1990-01-02"
	mobile := "09123333333"

	req := profile.UpdateProfileRequest{
		ProfileName: nil,
		PersonDetails: &profile.PersonDetailsDTO{
			FirstName:    &fname,
			LastName:     &lname,
			NationalID:   &nid,
			DOB:          &dob,
			MobileNumber: &mobile,
		},
	}

	resp, err := svc.UpdateDraft(ctx, 2, 200, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "200", resp.ID)
	assert.Equal(t, "John", resp.PersonDetails.FirstName)
	assert.Equal(t, "Doe", resp.PersonDetails.LastName)

	repo.AssertExpectations(t)
}

func TestUpdateDraft_Personal_InvalidDOB(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	existing := &entity.Profile{ProfileID: 201, UserID: 2, ProfileType: entity.ProfileTypePersonal, VerificationStatus: entity.StatusDraft}
	repo.On("GetByID", mock.Anything, uint64(201)).Return(existing, nil)

	fname := "Jane"
	lname := "Smith"
	nid := "0123456789"
	dob := "not-a-date"
	mobile := "09124444444"
	req := profile.UpdateProfileRequest{PersonDetails: &profile.PersonDetailsDTO{FirstName: &fname, LastName: &lname, NationalID: &nid, DOB: &dob, MobileNumber: &mobile}}

	resp, err := svc.UpdateDraft(ctx, 2, 201, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid date format")

	repo.AssertExpectations(t)
}

func TestUpdateDraft_Business_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	now := time.Now()
	existing := &entity.Profile{ProfileID: 300, UserID: 3, ProfileType: entity.ProfileTypeBusiness, ProfileName: "Old Biz", VerificationStatus: entity.StatusDraft, CreatedAt: now}
	repo.On("GetByID", mock.Anything, uint64(300)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	repFirst := "Rep"
	repLast := "Leader"
	repNid := "9876543210"
	repDob := "1985-02-02"
	repMobile := "09126666666"
	bnid := "11122233344"

	req := profile.UpdateProfileRequest{BusinessDetails: &profile.BusinessDetailsDTO{
		RepFirstName:       &repFirst,
		RepLastName:        &repLast,
		RepNationalID:      &repNid,
		RepDOB:             &repDob,
		RepMobileNumber:    &repMobile,
		BusinessNationalID: &bnid,
	}}

	resp, err := svc.UpdateDraft(ctx, 3, 300, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "300", resp.ID)
	assert.Equal(t, "Rep", resp.BusinessDetails.RepFirstName)
	assert.Equal(t, "Leader", resp.BusinessDetails.RepLastName)

	repo.AssertExpectations(t)
}

func TestSaveDocument_Personal_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	existing := &entity.Profile{ProfileID: 400, UserID: 4, ProfileType: entity.ProfileTypePersonal, VerificationStatus: entity.StatusDraft, PersonDetails: &entity.ProfilePersonDetails{ProfileID: 400}}
	// PersonDetails nil -> should create PersonDetails pointer
	repo.On("GetByID", mock.Anything, uint64(400)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	req := profile.SaveDocumentRequest{DocumentType: "national_card_front", FileUrl: "https://example.com/front.png"}
	err := svc.SaveDocument(ctx, 4, 400, req)
	assert.NoError(t, err)

	// Ensure subject updated
	repo.AssertExpectations(t)
}

func TestSaveDocument_Business_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	existing := &entity.Profile{ProfileID: 401, UserID: 4, ProfileType: entity.ProfileTypeBusiness, VerificationStatus: entity.StatusDraft, BusinessDetails: &entity.ProfileBusinessDetails{ProfileID: 401}}
	repo.On("GetByID", mock.Anything, uint64(401)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	req := profile.SaveDocumentRequest{DocumentType: "establishment_notice", FileUrl: "https://example.com/est.pdf"}
	err := svc.SaveDocument(ctx, 4, 401, req)
	assert.NoError(t, err)

	repo.AssertExpectations(t)
}

func TestSaveDocument_InvalidType(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	existing := &entity.Profile{ProfileID: 402, UserID: 4, ProfileType: entity.ProfileTypePersonal, VerificationStatus: entity.StatusDraft, PersonDetails: &entity.ProfilePersonDetails{ProfileID: 402}}
	repo.On("GetByID", mock.Anything, uint64(402)).Return(existing, nil)

	req := profile.SaveDocumentRequest{DocumentType: "unknown", FileUrl: "https://example.com/a.png"}
	err := svc.SaveDocument(ctx, 4, 402, req)
	assert.Error(t, err)

	repo.AssertExpectations(t)
}

func TestSubmitProfile_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	// Create a complete personal profile
	dobTime := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := &entity.Profile{ProfileID: 500, UserID: 5, ProfileType: entity.ProfileTypePersonal, VerificationStatus: entity.StatusDraft, PersonDetails: &entity.ProfilePersonDetails{ProfileID: 500, FirstName: "A", LastName: "B", NationalID: "1234567890", DOB: dobTime, MobileNumber: "09120000000"}}
	repo.On("GetByID", mock.Anything, uint64(500)).Return(existing, nil)
	repo.On("Update", mock.Anything, mock.Anything).Return(nil)

	resp, err := svc.SubmitProfile(ctx, 5, 500)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "pending", resp.VerificationStatus)

	repo.AssertExpectations(t)
}

func TestSubmitProfile_Incomplete_Fails(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	// Missing required person fields
	existing := &entity.Profile{ProfileID: 501, UserID: 5, ProfileType: entity.ProfileTypePersonal, VerificationStatus: entity.StatusDraft, PersonDetails: &entity.ProfilePersonDetails{ProfileID: 501}}
	repo.On("GetByID", mock.Anything, uint64(501)).Return(existing, nil)

	resp, err := svc.SubmitProfile(ctx, 5, 501)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "first name is required")

	repo.AssertExpectations(t)
}

func TestGetByID_Maps(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	dob := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	existing := &entity.Profile{ProfileID: 600, UserID: 6, ProfileType: entity.ProfileTypePersonal, ProfileName: "P1", Balance: 10, VerificationStatus: entity.StatusVerified, IsActive: true, CreatedAt: time.Now(), PersonDetails: &entity.ProfilePersonDetails{FirstName: "F", LastName: "L", NationalID: "0123456789", DOB: dob, MobileNumber: "09120000001"}}
	repo.On("GetByID", mock.Anything, uint64(600)).Return(existing, nil)

	resp, err := svc.GetByID(ctx, 6, 600)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "600", resp.ID)
	assert.Equal(t, "P1", resp.Name)
	assert.NotNil(t, resp.PersonDetails)
	assert.Equal(t, "F", resp.PersonDetails.FirstName)

	repo.AssertExpectations(t)
}

func TestGetUserProfiles_Success(t *testing.T) {
	ctx := context.Background()
	repo := mocks.NewMockProfileRepository(t)
	svc := NewProfileService(repo, nil)

	now := time.Now()
	p1 := &entity.Profile{ProfileID: 700, UserID: 7, ProfileType: entity.ProfileTypePersonal, ProfileName: "P1", CreatedAt: now}
	p2 := &entity.Profile{ProfileID: 701, UserID: 7, ProfileType: entity.ProfileTypeBusiness, ProfileName: "P2", CreatedAt: now}
	repo.On("GetByUserID", mock.Anything, uint64(7)).Return([]*entity.Profile{p1, p2}, nil)

	resp, err := svc.GetUserProfiles(ctx, 7)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 2, len(resp))

	repo.AssertExpectations(t)
}
