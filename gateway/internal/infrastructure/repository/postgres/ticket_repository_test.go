package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/entity"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/domain/enum"
	repoPostgres "github.com/Barbod-Biometrics/Backend/gateway/internal/infrastructure/repository/postgres"
	"github.com/Barbod-Biometrics/Backend/gateway/internal/test/testcontainer"
	"github.com/stretchr/testify/assert"
)

func TestTicketRepository_Create(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	// Create a user first (FK dependency)
	user := &entity.User{
		PhoneNumber: "09121234567",
		CreatedAt:   time.Now(),
	}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	ticket := &entity.Ticket{
		UserID:      user.UserID,
		Service:     enum.TicketServiceOCR,
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      enum.TicketStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = repo.Create(ctx, ticket)
	assert.NoError(t, err)
	assert.NotZero(t, ticket.TicketID)

	// Verify persistence
	var savedTicket entity.Ticket
	err = pg.DB.First(&savedTicket, ticket.TicketID).Error
	assert.NoError(t, err)
	assert.Equal(t, ticket.Title, savedTicket.Title)
	assert.Equal(t, ticket.Description, savedTicket.Description)
	assert.Equal(t, ticket.Service, savedTicket.Service)
	assert.Equal(t, ticket.Status, savedTicket.Status)
}

func TestTicketRepository_Create_WithAttachment(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234568", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	attachmentKey := "tickets/1/file.png"
	ticket := &entity.Ticket{
		UserID:        user.UserID,
		Service:       enum.TicketServiceFaceAuth,
		Title:         "Face Auth Issue",
		Description:   "Can't verify face",
		Status:        enum.TicketStatusPending,
		AttachmentKey: &attachmentKey,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = repo.Create(ctx, ticket)
	assert.NoError(t, err)

	var savedTicket entity.Ticket
	err = pg.DB.First(&savedTicket, ticket.TicketID).Error
	assert.NoError(t, err)
	assert.NotNil(t, savedTicket.AttachmentKey)
	assert.Equal(t, attachmentKey, *savedTicket.AttachmentKey)
}

func TestTicketRepository_GetByID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234569", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	// Seed
	ticket := &entity.Ticket{
		UserID:      user.UserID,
		Service:     enum.TicketServiceOCR,
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      enum.TicketStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = pg.DB.Create(ticket).Error
	assert.NoError(t, err)

	// Found
	{
		found, err := repo.GetByID(ctx, ticket.TicketID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, ticket.TicketID, found.TicketID)
		assert.Equal(t, ticket.Title, found.Title)
	}

	// Not Found
	{
		found, err := repo.GetByID(ctx, 999999)
		assert.NoError(t, err)
		assert.Nil(t, found)
	}
}

func TestTicketRepository_GetByIDWithUser(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	email := "user@example.com"
	user := &entity.User{PhoneNumber: "09121234570", Email: &email, CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	ticket := &entity.Ticket{
		UserID:      user.UserID,
		Service:     enum.TicketServiceOCR,
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      enum.TicketStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = pg.DB.Create(ticket).Error
	assert.NoError(t, err)

	// Found with user
	{
		found, err := repo.GetByIDWithUser(ctx, ticket.TicketID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, ticket.TicketID, found.TicketID)
		assert.NotNil(t, found.User)
		assert.Equal(t, user.PhoneNumber, found.User.PhoneNumber)
		assert.Equal(t, email, *found.User.Email)
	}

	// Not Found
	{
		found, err := repo.GetByIDWithUser(ctx, 999999)
		assert.NoError(t, err)
		assert.Nil(t, found)
	}
}

func TestTicketRepository_GetByUserID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user1 := &entity.User{PhoneNumber: "09121234571", CreatedAt: time.Now()}
	user2 := &entity.User{PhoneNumber: "09121234572", CreatedAt: time.Now()}
	err = pg.DB.Create(user1).Error
	assert.NoError(t, err)
	err = pg.DB.Create(user2).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	// Create tickets for user1
	tickets := []entity.Ticket{
		{UserID: user1.UserID, Service: enum.TicketServiceOCR, Title: "User1 Ticket 1", Description: "Desc 1", Status: enum.TicketStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{UserID: user1.UserID, Service: enum.TicketServiceFaceAuth, Title: "User1 Ticket 2", Description: "Desc 2", Status: enum.TicketStatusAnswered, CreatedAt: time.Now().Add(time.Hour), UpdatedAt: time.Now()},
	}
	for i := range tickets {
		err = pg.DB.Create(&tickets[i]).Error
		assert.NoError(t, err)
	}

	// Create ticket for user2
	user2Ticket := &entity.Ticket{UserID: user2.UserID, Service: enum.TicketServiceOCR, Title: "User2 Ticket", Description: "Desc", Status: enum.TicketStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = pg.DB.Create(user2Ticket).Error
	assert.NoError(t, err)

	// Get user1 tickets
	{
		found, err := repo.GetByUserID(ctx, user1.UserID)
		assert.NoError(t, err)
		assert.Len(t, found, 2)
		// Should be ordered by created_at DESC
		assert.Equal(t, "User1 Ticket 2", found[0].Title)
		assert.Equal(t, "User1 Ticket 1", found[1].Title)
	}

	// Get user2 tickets
	{
		found, err := repo.GetByUserID(ctx, user2.UserID)
		assert.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, "User2 Ticket", found[0].Title)
	}

	// No tickets for non-existent user
	{
		found, err := repo.GetByUserID(ctx, 999999)
		assert.NoError(t, err)
		assert.Len(t, found, 0)
	}
}

func TestTicketRepository_GetAll(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234573", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	// Create multiple tickets
	tickets := []entity.Ticket{
		{UserID: user.UserID, Service: enum.TicketServiceOCR, Title: "Ticket 1", Description: "Desc 1", Status: enum.TicketStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()},
		{UserID: user.UserID, Service: enum.TicketServiceFaceAuth, Title: "Ticket 2", Description: "Desc 2", Status: enum.TicketStatusAnswered, CreatedAt: time.Now().Add(time.Hour), UpdatedAt: time.Now()},
		{UserID: user.UserID, Service: enum.TicketServiceOCR, Title: "Ticket 3", Description: "Desc 3", Status: enum.TicketStatusClosed, CreatedAt: time.Now().Add(2 * time.Hour), UpdatedAt: time.Now()},
	}
	for i := range tickets {
		err = pg.DB.Create(&tickets[i]).Error
		assert.NoError(t, err)
	}

	found, err := repo.GetAll(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, 3)
	// Should be ordered by created_at DESC
	assert.Equal(t, "Ticket 3", found[0].Title)
	assert.Equal(t, "Ticket 2", found[1].Title)
	assert.Equal(t, "Ticket 1", found[2].Title)
}

func TestTicketRepository_GetAllWithUsers(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	email1 := "user1@example.com"
	email2 := "user2@example.com"
	user1 := &entity.User{PhoneNumber: "09121234574", Email: &email1, CreatedAt: time.Now()}
	user2 := &entity.User{PhoneNumber: "09121234575", Email: &email2, CreatedAt: time.Now()}
	err = pg.DB.Create(user1).Error
	assert.NoError(t, err)
	err = pg.DB.Create(user2).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	// Create tickets
	ticket1 := &entity.Ticket{UserID: user1.UserID, Service: enum.TicketServiceOCR, Title: "Ticket 1", Description: "Desc 1", Status: enum.TicketStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	ticket2 := &entity.Ticket{UserID: user2.UserID, Service: enum.TicketServiceFaceAuth, Title: "Ticket 2", Description: "Desc 2", Status: enum.TicketStatusAnswered, CreatedAt: time.Now().Add(time.Hour), UpdatedAt: time.Now()}
	err = pg.DB.Create(ticket1).Error
	assert.NoError(t, err)
	err = pg.DB.Create(ticket2).Error
	assert.NoError(t, err)

	found, err := repo.GetAllWithUsers(ctx)
	assert.NoError(t, err)
	assert.Len(t, found, 2)

	// Ticket 2 should be first (DESC by created_at)
	assert.Equal(t, "Ticket 2", found[0].Title)
	assert.NotNil(t, found[0].User)
	assert.Equal(t, user2.PhoneNumber, found[0].User.PhoneNumber)
	assert.Equal(t, email2, *found[0].User.Email)

	assert.Equal(t, "Ticket 1", found[1].Title)
	assert.NotNil(t, found[1].User)
	assert.Equal(t, user1.PhoneNumber, found[1].User.PhoneNumber)
}

func TestTicketRepository_UpdateStatus(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234576", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	ticket := &entity.Ticket{
		UserID:      user.UserID,
		Service:     enum.TicketServiceOCR,
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      enum.TicketStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = pg.DB.Create(ticket).Error
	assert.NoError(t, err)

	// Update status to answered
	err = repo.UpdateStatus(ctx, ticket.TicketID, enum.TicketStatusAnswered)
	assert.NoError(t, err)

	// Verify
	var updated entity.Ticket
	err = pg.DB.First(&updated, ticket.TicketID).Error
	assert.NoError(t, err)
	assert.Equal(t, enum.TicketStatusAnswered, updated.Status)

	// Update status to closed
	err = repo.UpdateStatus(ctx, ticket.TicketID, enum.TicketStatusClosed)
	assert.NoError(t, err)

	err = pg.DB.First(&updated, ticket.TicketID).Error
	assert.NoError(t, err)
	assert.Equal(t, enum.TicketStatusClosed, updated.Status)
}

func TestTicketRepository_Update(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234577", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	ticket := &entity.Ticket{
		UserID:      user.UserID,
		Service:     enum.TicketServiceOCR,
		Title:       "Original Title",
		Description: "Original Description",
		Status:      enum.TicketStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = pg.DB.Create(ticket).Error
	assert.NoError(t, err)

	// Update ticket
	ticket.Title = "Updated Title"
	ticket.Description = "Updated Description"
	ticket.Status = enum.TicketStatusAnswered

	err = repo.Update(ctx, ticket)
	assert.NoError(t, err)

	// Verify
	var updated entity.Ticket
	err = pg.DB.First(&updated, ticket.TicketID).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Title", updated.Title)
	assert.Equal(t, "Updated Description", updated.Description)
	assert.Equal(t, enum.TicketStatusAnswered, updated.Status)
}

func TestTicketRepository_CreateMessage(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234578", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	ticket := &entity.Ticket{
		UserID:      user.UserID,
		Service:     enum.TicketServiceOCR,
		Title:       "Test Ticket",
		Description: "Test Description",
		Status:      enum.TicketStatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	err = pg.DB.Create(ticket).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	message := &entity.TicketMessage{
		TicketID:  ticket.TicketID,
		Sender:    enum.MessageSenderUser,
		Message:   "Hello, I need help",
		CreatedAt: time.Now(),
	}

	err = repo.CreateMessage(ctx, message)
	assert.NoError(t, err)
	assert.NotZero(t, message.MessageID)

	// Verify persistence
	var savedMessage entity.TicketMessage
	err = pg.DB.First(&savedMessage, message.MessageID).Error
	assert.NoError(t, err)
	assert.Equal(t, message.TicketID, savedMessage.TicketID)
	assert.Equal(t, message.Sender, savedMessage.Sender)
	assert.Equal(t, message.Message, savedMessage.Message)
}

func TestTicketRepository_GetMessagesByTicketID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234579", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	ticket1 := &entity.Ticket{UserID: user.UserID, Service: enum.TicketServiceOCR, Title: "Ticket 1", Description: "Desc 1", Status: enum.TicketStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	ticket2 := &entity.Ticket{UserID: user.UserID, Service: enum.TicketServiceOCR, Title: "Ticket 2", Description: "Desc 2", Status: enum.TicketStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = pg.DB.Create(ticket1).Error
	assert.NoError(t, err)
	err = pg.DB.Create(ticket2).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	// Create messages for ticket1
	messages := []entity.TicketMessage{
		{TicketID: ticket1.TicketID, Sender: enum.MessageSenderUser, Message: "Message 1", CreatedAt: time.Now()},
		{TicketID: ticket1.TicketID, Sender: enum.MessageSenderAdmin, Message: "Message 2", CreatedAt: time.Now().Add(time.Minute)},
		{TicketID: ticket1.TicketID, Sender: enum.MessageSenderUser, Message: "Message 3", CreatedAt: time.Now().Add(2 * time.Minute)},
	}
	for i := range messages {
		err = pg.DB.Create(&messages[i]).Error
		assert.NoError(t, err)
	}

	// Create message for ticket2
	ticket2Msg := &entity.TicketMessage{TicketID: ticket2.TicketID, Sender: enum.MessageSenderUser, Message: "Ticket 2 Message", CreatedAt: time.Now()}
	err = pg.DB.Create(ticket2Msg).Error
	assert.NoError(t, err)

	// Get messages for ticket1
	{
		found, err := repo.GetMessagesByTicketID(ctx, ticket1.TicketID)
		assert.NoError(t, err)
		assert.Len(t, found, 3)
		// Should be ordered by created_at ASC
		assert.Equal(t, "Message 1", found[0].Message)
		assert.Equal(t, enum.MessageSenderUser, found[0].Sender)
		assert.Equal(t, "Message 2", found[1].Message)
		assert.Equal(t, enum.MessageSenderAdmin, found[1].Sender)
		assert.Equal(t, "Message 3", found[2].Message)
	}

	// Get messages for ticket2
	{
		found, err := repo.GetMessagesByTicketID(ctx, ticket2.TicketID)
		assert.NoError(t, err)
		assert.Len(t, found, 1)
		assert.Equal(t, "Ticket 2 Message", found[0].Message)
	}

	// No messages for non-existent ticket
	{
		found, err := repo.GetMessagesByTicketID(ctx, 999999)
		assert.NoError(t, err)
		assert.Len(t, found, 0)
	}
}

func TestTicketRepository_GetLastMessageByTicketID(t *testing.T) {
	ctx := context.Background()
	pg, err := testcontainer.SetupPostgres(ctx)
	if err != nil {
		t.Fatalf("failed to setup postgres container: %v", err)
	}
	defer pg.Terminate(ctx)

	user := &entity.User{PhoneNumber: "09121234580", CreatedAt: time.Now()}
	err = pg.DB.Create(user).Error
	assert.NoError(t, err)

	ticket := &entity.Ticket{UserID: user.UserID, Service: enum.TicketServiceOCR, Title: "Test Ticket", Description: "Desc", Status: enum.TicketStatusPending, CreatedAt: time.Now(), UpdatedAt: time.Now()}
	err = pg.DB.Create(ticket).Error
	assert.NoError(t, err)

	repo := repoPostgres.NewTicketRepository(pg.DB)

	// No messages yet
	{
		found, err := repo.GetLastMessageByTicketID(ctx, ticket.TicketID)
		assert.NoError(t, err)
		assert.Nil(t, found)
	}

	// Create messages
	messages := []entity.TicketMessage{
		{TicketID: ticket.TicketID, Sender: enum.MessageSenderUser, Message: "First message", CreatedAt: time.Now()},
		{TicketID: ticket.TicketID, Sender: enum.MessageSenderAdmin, Message: "Second message", CreatedAt: time.Now().Add(time.Minute)},
		{TicketID: ticket.TicketID, Sender: enum.MessageSenderUser, Message: "Last message", CreatedAt: time.Now().Add(2 * time.Minute)},
	}
	for i := range messages {
		err = pg.DB.Create(&messages[i]).Error
		assert.NoError(t, err)
	}

	// Get last message
	{
		found, err := repo.GetLastMessageByTicketID(ctx, ticket.TicketID)
		assert.NoError(t, err)
		assert.NotNil(t, found)
		assert.Equal(t, "Last message", found.Message)
		assert.Equal(t, enum.MessageSenderUser, found.Sender)
	}

	// Non-existent ticket
	{
		found, err := repo.GetLastMessageByTicketID(ctx, 999999)
		assert.NoError(t, err)
		assert.Nil(t, found)
	}
}
