package repository

import (
	"context"
	"fmt"

	"youdoyou-server/model"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

type FirestoreChatRepository struct {
	client *firestore.Client
}

func NewFirestoreChatRepository(client *firestore.Client) ChatRepository {
	return &FirestoreChatRepository{client: client}
}

func (r *FirestoreChatRepository) GetUnmemorizedMessages(ctx context.Context, threadID string) ([]model.ChatMessage, error) {
	// Get thread to check memorizedUntil
	thread, err := r.GetThread(ctx, threadID)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}

	// Query for messages created after memorizedUntil
	query := r.client.Collection("threads").Doc(threadID).
		Collection("messages").
		Where("createdAt", ">", thread.MemorizedUntil).
		OrderBy("createdAt", firestore.Asc)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("failed to query messages: %w", err)
	}

	var messages []model.ChatMessage
	for _, doc := range docs {
		var msg model.ChatMessage
		if err := doc.DataTo(&msg); err != nil {
			return nil, fmt.Errorf("failed to parse message data: %w", err)
		}
		msg.ID = doc.Ref.ID
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *FirestoreChatRepository) GetThread(ctx context.Context, threadID string) (*model.ChatThread, error) {
	doc, err := r.client.Collection("threads").Doc(threadID).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get thread: %w", err)
	}
	var thread model.ChatThread
	if err := doc.DataTo(&thread); err != nil {
		return nil, fmt.Errorf("failed to parse thread data: %w", err)
	}
	thread.ID = doc.Ref.ID
	return &thread, nil
}

func (r *FirestoreChatRepository) SaveMessage(ctx context.Context, message *model.ChatMessage) (string, error) {
	// Generate UUID v7
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("failed to generate UUID v7: %w", err)
	}
	idStr := id.String()

	// Use Doc(id).Set instead of Add
	_, err = r.client.Collection("threads").Doc(message.ThreadID).
		Collection("messages").Doc(idStr).
		Set(ctx, message)

	if err != nil {
		return "", err
	}

	return idStr, nil
}

func (r *FirestoreChatRepository) CreateThread(ctx context.Context, thread *model.ChatThread) error {
	// Generate UUID v7 for thread ID if not set
	if thread.ID == "" {
		id, err := uuid.NewV7()
		if err != nil {
			return fmt.Errorf("failed to generate UUID v7: %w", err)
		}
		thread.ID = id.String()
	}

	_, err := r.client.Collection("threads").Doc(thread.ID).Set(ctx, map[string]interface{}{
		"userId":         thread.UserID,
		"firstMessage":   thread.FirstMessage,
		"unreadCount":    thread.UnreadCount,
		"lastReadAt":     thread.LastReadAt,
		"replyCount":     thread.ReplyCount,
		"isPrivate":      thread.IsPrivate,
		"isArchived":     thread.IsArchived,
		"sessionMemory":  thread.SessionMemory,
		"memorizedUntil": thread.MemorizedUntil,
		"createdAt":      thread.CreatedAt,
	})
	return err
}
