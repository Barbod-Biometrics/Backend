package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/application/dto/ticket"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/test/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCreateTicket_Success(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		tk := a.Get(1).(*entity.Ticket)
		tk.TicketID = 123
	}).Return(nil)

	req := ticket.CreateTicketRequest{
		Service:     "ocr",
		Title:       "Test Ticket",
		Description: "Test Description",
	}

	resp, err := svc.CreateTicket(ctx, 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "123", resp.ID)
	assert.Equal(t, "OCR مدارک", resp.Service)
	assert.Equal(t, "Test Ticket", resp.Title)
	assert.Equal(t, "pending", resp.Status)

	ticketRepo.AssertExpectations(t)
}

func TestCreateTicket_WithAttachment(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("Create", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		tk := a.Get(1).(*entity.Ticket)
		tk.TicketID = 124
		assert.NotNil(t, tk.AttachmentKey)
		assert.Equal(t, "tickets/1/attachment.png", *tk.AttachmentKey)
	}).Return(nil)

	req := ticket.CreateTicketRequest{
		Service:     "face_auth",
		Title:       "Face Auth Issue",
		Description: "Can't verify face",
		Attachment:  "tickets/1/attachment.png",
	}

	resp, err := svc.CreateTicket(ctx, 1, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "124", resp.ID)

	ticketRepo.AssertExpectations(t)
}

func TestCreateTicket_Error(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("Create", mock.Anything, mock.Anything).Return(errors.New("database error"))

	req := ticket.CreateTicketRequest{
		Service:     "ocr",
		Title:       "Test Ticket",
		Description: "Test Description",
	}

	resp, err := svc.CreateTicket(ctx, 1, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "failed to create ticket")

	ticketRepo.AssertExpectations(t)
}

func TestGetUserTickets_Success(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	now := time.Now()
	tickets := []*entity.Ticket{
		{TicketID: 1, UserID: 1, Service: enum.TicketServiceOCR, Title: "Ticket 1", Status: enum.TicketStatusPending, CreatedAt: now},
		{TicketID: 2, UserID: 1, Service: enum.TicketServiceFaceAuth, Title: "Ticket 2", Status: enum.TicketStatusAnswered, CreatedAt: now},
	}
	ticketRepo.On("GetByUserID", mock.Anything, uint64(1)).Return(tickets, nil)

	resp, err := svc.GetUserTickets(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "1", resp[0].ID)
	assert.Equal(t, "OCR مدارک", resp[0].Service)
	assert.Equal(t, "2", resp[1].ID)
	assert.Equal(t, "احراز هویت چهره", resp[1].Service)

	ticketRepo.AssertExpectations(t)
}

func TestGetUserTickets_Empty(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("GetByUserID", mock.Anything, uint64(1)).Return([]*entity.Ticket{}, nil)

	resp, err := svc.GetUserTickets(ctx, 1)
	assert.NoError(t, err)
	assert.Len(t, resp, 0)

	ticketRepo.AssertExpectations(t)
}

func TestGetAllTickets_Success(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	now := time.Now()
	email := "user@example.com"
	user := &entity.User{UserID: 1, PhoneNumber: "09121234567", Email: &email}

	tickets := []*entity.Ticket{
		{TicketID: 1, UserID: 1, Service: enum.TicketServiceOCR, Title: "Ticket 1", Status: enum.TicketStatusPending, CreatedAt: now, User: user},
		{TicketID: 2, UserID: 1, Service: enum.TicketServiceFaceAuth, Title: "Ticket 2", Status: enum.TicketStatusAnswered, CreatedAt: now, User: user},
	}
	ticketRepo.On("GetAllWithUsers", mock.Anything).Return(tickets, nil)
	ticketRepo.On("GetLastMessageByTicketID", mock.Anything, uint64(1)).Return((*entity.TicketMessage)(nil), nil)
	ticketRepo.On("GetLastMessageByTicketID", mock.Anything, uint64(2)).Return(&entity.TicketMessage{CreatedAt: now}, nil)

	resp, err := svc.GetAllTickets(ctx)
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "tk_1", resp[0].ID)
	assert.Equal(t, "09121234567", resp[0].UserContact.Phone)
	assert.Equal(t, "user@example.com", resp[0].UserContact.Email)
	assert.Nil(t, resp[0].LastMessageAt)
	assert.NotNil(t, resp[1].LastMessageAt)

	ticketRepo.AssertExpectations(t)
}

func TestGetTicketDetail_Success(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	now := time.Now()
	email := "user@example.com"
	user := &entity.User{UserID: 1, PhoneNumber: "09121234567", Email: &email}
	tk := &entity.Ticket{
		TicketID:    1,
		UserID:      1,
		Service:     enum.TicketServiceOCR,
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      enum.TicketStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
		User:        user,
	}
	ticketRepo.On("GetByIDWithUser", mock.Anything, uint64(1)).Return(tk, nil)

	resp, err := svc.GetTicketDetail(ctx, 1)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "tk_1", resp.ID)
	assert.Equal(t, "u_1", resp.UserID)
	assert.Equal(t, "ocr", resp.Service)
	assert.Equal(t, "Test Ticket", resp.Title)
	assert.Equal(t, "Test Description", resp.Message)
	assert.Equal(t, "09121234567", resp.UserContact.Phone)
	assert.Equal(t, "user@example.com", resp.UserContact.Email)

	ticketRepo.AssertExpectations(t)
}

func TestGetTicketDetail_NotFound(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("GetByIDWithUser", mock.Anything, uint64(999)).Return((*entity.Ticket)(nil), nil)

	resp, err := svc.GetTicketDetail(ctx, 999)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTicketNotFound)

	ticketRepo.AssertExpectations(t)
}

func TestGetTicketMessages_Success_Owner(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	now := time.Now()
	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	messages := []*entity.TicketMessage{
		{MessageID: 1, TicketID: 1, Sender: enum.MessageSenderUser, Message: "Hello", CreatedAt: now},
		{MessageID: 2, TicketID: 1, Sender: enum.MessageSenderAdmin, Message: "Hi there", CreatedAt: now.Add(time.Minute)},
	}

	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)
	ticketRepo.On("GetMessagesByTicketID", mock.Anything, uint64(1)).Return(messages, nil)

	resp, err := svc.GetTicketMessages(ctx, 1, 1, false) // user owns ticket
	assert.NoError(t, err)
	assert.Len(t, resp, 2)
	assert.Equal(t, "msg_1", resp[0].ID)
	assert.Equal(t, "user", resp[0].Sender)
	assert.Equal(t, "msg_2", resp[1].ID)
	assert.Equal(t, "admin", resp[1].Sender)

	ticketRepo.AssertExpectations(t)
}

func TestGetTicketMessages_Success_Admin(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	now := time.Now()
	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	messages := []*entity.TicketMessage{
		{MessageID: 1, TicketID: 1, Sender: enum.MessageSenderUser, Message: "Hello", CreatedAt: now},
	}

	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)
	ticketRepo.On("GetMessagesByTicketID", mock.Anything, uint64(1)).Return(messages, nil)

	// Admin (userID 99) accessing ticket of user 1
	resp, err := svc.GetTicketMessages(ctx, 1, 99, true)
	assert.NoError(t, err)
	assert.Len(t, resp, 1)

	ticketRepo.AssertExpectations(t)
}

func TestGetTicketMessages_Unauthorized(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)

	// User 2 trying to access ticket of user 1 (not admin)
	resp, err := svc.GetTicketMessages(ctx, 1, 2, false)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrUnauthorized)

	ticketRepo.AssertExpectations(t)
}

func TestGetTicketMessages_NotFound(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("GetByID", mock.Anything, uint64(999)).Return((*entity.Ticket)(nil), nil)

	resp, err := svc.GetTicketMessages(ctx, 999, 1, false)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTicketNotFound)

	ticketRepo.AssertExpectations(t)
}

func TestSendMessage_Success_User(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)
	ticketRepo.On("CreateMessage", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		msg := a.Get(1).(*entity.TicketMessage)
		msg.MessageID = 10
		assert.Equal(t, enum.MessageSenderUser, msg.Sender)
	}).Return(nil)
	ticketRepo.On("UpdateStatus", mock.Anything, uint64(1), enum.TicketStatusPending).Return(nil)

	req := ticket.SendMessageRequest{Message: "Test message"}
	resp, err := svc.SendMessage(ctx, 1, 1, false, req)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "msg_10", resp.ID)
	assert.Equal(t, "user", resp.Sender)
	assert.Equal(t, "Test message", resp.Message)

	ticketRepo.AssertExpectations(t)
}

func TestSendMessage_Success_Admin(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)
	ticketRepo.On("CreateMessage", mock.Anything, mock.Anything).Run(func(a mock.Arguments) {
		msg := a.Get(1).(*entity.TicketMessage)
		msg.MessageID = 11
		assert.Equal(t, enum.MessageSenderAdmin, msg.Sender)
	}).Return(nil)
	ticketRepo.On("UpdateStatus", mock.Anything, uint64(1), enum.TicketStatusAnswered).Return(nil)

	req := ticket.SendMessageRequest{Message: "Admin response"}
	resp, err := svc.SendMessage(ctx, 1, 99, true, req) // admin (userID 99)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "msg_11", resp.ID)
	assert.Equal(t, "admin", resp.Sender)

	ticketRepo.AssertExpectations(t)
}

func TestSendMessage_TicketClosed(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusClosed}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)

	req := ticket.SendMessageRequest{Message: "Test message"}
	resp, err := svc.SendMessage(ctx, 1, 1, false, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTicketClosed)

	ticketRepo.AssertExpectations(t)
}

func TestSendMessage_Unauthorized(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)

	req := ticket.SendMessageRequest{Message: "Test message"}
	resp, err := svc.SendMessage(ctx, 1, 2, false, req) // user 2 trying to message on user 1's ticket
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrUnauthorized)

	ticketRepo.AssertExpectations(t)
}

func TestSendMessage_NotFound(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("GetByID", mock.Anything, uint64(999)).Return((*entity.Ticket)(nil), nil)

	req := ticket.SendMessageRequest{Message: "Test message"}
	resp, err := svc.SendMessage(ctx, 999, 1, false, req)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTicketNotFound)

	ticketRepo.AssertExpectations(t)
}

func TestCloseTicket_Success_User(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)
	ticketRepo.On("UpdateStatus", mock.Anything, uint64(1), enum.TicketStatusClosed).Return(nil)

	resp, err := svc.CloseTicket(ctx, 1, 1, false)
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "tk_1", resp.TicketID)
	assert.Equal(t, "closed", resp.Status)

	ticketRepo.AssertExpectations(t)
}

func TestCloseTicket_Success_Admin(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusAnswered}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)
	ticketRepo.On("UpdateStatus", mock.Anything, uint64(1), enum.TicketStatusClosed).Return(nil)

	resp, err := svc.CloseTicket(ctx, 1, 99, true) // admin closing user's ticket
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "tk_1", resp.TicketID)
	assert.Equal(t, "closed", resp.Status)

	ticketRepo.AssertExpectations(t)
}

func TestCloseTicket_AlreadyClosed(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusClosed}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)

	resp, err := svc.CloseTicket(ctx, 1, 1, false)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTicketClosed)

	ticketRepo.AssertExpectations(t)
}

func TestCloseTicket_Unauthorized(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	tk := &entity.Ticket{TicketID: 1, UserID: 1, Status: enum.TicketStatusPending}
	ticketRepo.On("GetByID", mock.Anything, uint64(1)).Return(tk, nil)

	resp, err := svc.CloseTicket(ctx, 1, 2, false) // user 2 trying to close user 1's ticket
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrUnauthorized)

	ticketRepo.AssertExpectations(t)
}

func TestCloseTicket_NotFound(t *testing.T) {
	ctx := context.Background()
	ticketRepo := mocks.NewMockTicketRepository(t)
	userRepo := mocks.NewMockUserRepository(t)
	svc := NewTicketService(ticketRepo, userRepo, nil)

	ticketRepo.On("GetByID", mock.Anything, uint64(999)).Return((*entity.Ticket)(nil), nil)

	resp, err := svc.CloseTicket(ctx, 999, 1, false)
	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.ErrorIs(t, err, ErrTicketNotFound)

	ticketRepo.AssertExpectations(t)
}

func TestToJalaliDate(t *testing.T) {
	// Test that the function produces valid Jalali date format
	testDate := time.Date(2024, 3, 21, 0, 0, 0, 0, time.UTC)
	result := toJalaliDate(testDate)

	// Verify the format is correct (YYYY/MM/DD)
	assert.Regexp(t, `^\d{4}/\d{2}/\d{2}$`, result)

	// Test another date
	testDate2 := time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC)
	result2 := toJalaliDate(testDate2)
	assert.Regexp(t, `^\d{4}/\d{2}/\d{2}$`, result2)

	// Year should be around 1402-1404 for years 2024-2025
	assert.Contains(t, result, "140")
	assert.Contains(t, result2, "140")
}
