package repository

import (
	"context"
	"fmt"

	"youdoyou-server/model"

	"cloud.google.com/go/firestore"
)

type FirestoreChatRepository struct {
	client *firestore.Client
}

func NewFirestoreChatRepository(client *firestore.Client) ChatRepository {
	return &FirestoreChatRepository{client: client}
}

func (r *FirestoreChatRepository) GetLatestUserMessage(ctx context.Context, threadID string) (*model.ChatMessage, error) {
	query := r.client.Collection("threads").Doc(threadID).
		Collection("messages").
		Where("role", "==", "user").
		Where("status", "==", "unread").
		OrderBy("createdAt", firestore.Desc).
		Limit(1)

	docs, err := query.Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	if len(docs) == 0 {
		return nil, fmt.Errorf("no pending message found")
	}

	var msg model.ChatMessage
	if err := docs[0].DataTo(&msg); err != nil {
		return nil, fmt.Errorf("failed to parse message data: %v", err)
	}
	msg.ID = docs[0].Ref.ID

	return &msg, nil
}

func (r *FirestoreChatRepository) GetConversationHistory(ctx context.Context, threadID string) ([]model.ChatMessage, error) {
	query := r.client.Collection("threads").Doc(threadID).
		Collection("messages").
		OrderBy("createdAt", firestore.Asc)

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
	docRef, _, err := r.client.Collection("threads").Doc(message.ThreadID).
		Collection("messages").
		Add(ctx, message)

	if err != nil {
		return "", err
	}

	return docRef.ID, nil
}

func (r *FirestoreChatRepository) UpdateMessageStatus(ctx context.Context, threadID, messageID string, status string) error {
	_, err := r.client.Collection("threads").Doc(threadID).
		Collection("messages").Doc(messageID).
		Update(ctx, []firestore.Update{
			{Path: "status", Value: status},
		})
	return err
}
