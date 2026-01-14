package repository

import (
	"context"
	"crypto/rand"
	"fmt"
	"time"

	"youdoyou-server/model"

	"cloud.google.com/go/firestore"
	"github.com/oklog/ulid/v2"
)

type FirestoreChatRepository struct {
	client *firestore.Client
}

func NewFirestoreChatRepository(client *firestore.Client) ChatRepository {
	return &FirestoreChatRepository{client: client}
}

func (r *FirestoreChatRepository) GetLatestUserMessage(ctx context.Context, threadID string) (*model.ChatMessage, error) {
	// Query optimization: fetch all messages (ordered by ID implicitly via ULID or simple fetch)
	// User requested to remove WHERE clauses to avoid composite index requirement.
	// Since we are moving to standard fetch, we can reuse GetConversationHistory logic and filter in memory.

	messages, err := r.GetConversationHistory(ctx, threadID)
	if err != nil {
		return nil, err
	}

	// Filter in memory for the latest user message that is 'unread'
	// Iterate backwards since we want the latest
	for i := len(messages) - 1; i >= 0; i-- {
		msg := messages[i]
		if msg.Role == "user" && msg.Status == "unread" {
			// Return a pointer to the message
			return &msg, nil
		}
	}

	return nil, fmt.Errorf("no pending message found")
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

func (r *FirestoreChatRepository) GetConversationHistory(ctx context.Context, threadID string) ([]model.ChatMessage, error) {
	query := r.client.Collection("threads").Doc(threadID).
		Collection("messages").
		OrderBy(firestore.DocumentID, firestore.Asc)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	var messages []model.ChatMessage
	for _, doc := range docs {
		var msg model.ChatMessage
		if err := doc.DataTo(&msg); err != nil {
			return nil, fmt.Errorf("failed to parse history message data: %v", err)
		}
		msg.ID = doc.Ref.ID
		messages = append(messages, msg)
	}

	return messages, nil
}

func (r *FirestoreChatRepository) SaveMessage(ctx context.Context, message *model.ChatMessage) (string, error) {
	// Generate ULID
	entropy := ulid.Monotonic(rand.Reader, 0)
	id := ulid.MustNew(ulid.Timestamp(time.Now()), entropy).String()

	// Use Doc(id).Set instead of Add
	_, err := r.client.Collection("threads").Doc(message.ThreadID).
		Collection("messages").Doc(id).
		Set(ctx, message)

	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *FirestoreChatRepository) UpdateMessageStatus(ctx context.Context, threadID, messageID string, status string) error {
	_, err := r.client.Collection("threads").Doc(threadID).
		Collection("messages").Doc(messageID).
		Update(ctx, []firestore.Update{
			{Path: "status", Value: status},
		})
	return err
}
