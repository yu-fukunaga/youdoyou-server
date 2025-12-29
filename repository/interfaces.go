package repository

import (
	"context"
	"youdoyou-server/model"
)

// ChatRepository - Firestore
type ChatRepository interface {
	GetLatestUserMessage(ctx context.Context, threadID string) (*model.ChatMessage, error)
	GetConversationHistory(ctx context.Context, threadID string) ([]model.ChatMessage, error)
	SaveMessage(ctx context.Context, message *model.ChatMessage) (string, error)
	UpdateMessageStatus(ctx context.Context, threadID, messageID string, status string) error
}

// CalendarRepository - Google Calendar
type CalendarRepository interface {
	GetEvents(ctx context.Context, timeRange string, timezone string) ([]model.CalendarEvent, error)
}

// NotionRepository - Notion
type NotionRepository interface {
	QueryDatabase(ctx context.Context, databaseID string, filter map[string]interface{}) ([]model.NotionPage, error)
	CreatePage(ctx context.Context, databaseID string, properties map[string]interface{}) (string, error)
}
