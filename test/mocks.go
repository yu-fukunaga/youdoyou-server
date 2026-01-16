package test

import (
	"context"
	"time"

	"youdoyou-server/model"
	"youdoyou-server/repository"
)

// Mock ChatRepository
type MockChatRepository struct {
}

// Ensure interface compliance
var _ repository.ChatRepository = &MockChatRepository{}

func (m *MockChatRepository) GetUnmemorizedMessages(ctx context.Context, threadID string) ([]model.ChatMessage, error) {
	return []model.ChatMessage{}, nil
}

func (m *MockChatRepository) GetThread(ctx context.Context, threadID string) (*model.ChatThread, error) {
	return &model.ChatThread{
		ID:             threadID,
		SessionMemory:  "Mock session memory for testing",
		MemorizedUntil: time.Now().Add(-1 * time.Hour), // One hour ago
	}, nil
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, message *model.ChatMessage) (string, error) {
	return "mock_id", nil
}

func (m *MockChatRepository) CreateThread(ctx context.Context, thread *model.ChatThread) error {
	return nil
}

// Mock CalendarRepository
type MockCalendarRepository struct {
	events []model.CalendarEvent
}

// Ensure interface compliance
var _ repository.CalendarRepository = &MockCalendarRepository{}

func (m *MockCalendarRepository) GetEvents(ctx context.Context, timeRange string, timezone string) ([]model.CalendarEvent, error) {
	return m.events, nil
}

// Mock NotionRepository
type MockNotionRepository struct{}

// Ensure interface compliance
var _ repository.NotionRepository = &MockNotionRepository{}

func (m *MockNotionRepository) QueryDatabase(ctx context.Context, databaseID string, filter map[string]interface{}) ([]model.NotionPage, error) {
	return []model.NotionPage{}, nil
}

func (m *MockNotionRepository) CreatePage(ctx context.Context, databaseID string, properties map[string]interface{}) (string, error) {
	return "page_id", nil
}
