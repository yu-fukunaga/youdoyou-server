package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"youdoyou-server/config"
	"youdoyou-server/model"
	"youdoyou-server/repository"

	"cloud.google.com/go/firestore"
)

func main() {
	// Parse flags
	threadID := flag.String("thread-id", "", "Thread ID (optional, creates new thread if not provided)")
	message := flag.String("message", "", "Message content (required)")
	userID := flag.String("user-id", "default-user", "User ID for new threads")
	flag.Parse()

	// Validate required flags
	if *message == "" {
		fmt.Println("Error: --message is required")
		fmt.Println("")
		fmt.Println("Usage:")
		fmt.Println("  # Create new thread:")
		fmt.Println("  go run cmd/create-message --message '今日のスケジュールを教えて'")
		fmt.Println("")
		fmt.Println("  # Add to existing thread:")
		fmt.Println("  go run cmd/create-message --thread-id test-thread-001 --message '今日のスケジュールを教えて'")
		flag.PrintDefaults()
		return
	}

	ctx := context.Background()
	cfg := config.LoadConfig()

	// Set JST as default timezone
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		log.Printf("Warning: Failed to load Asia/Tokyo location: %v. Using UTC.", err)
		jst = time.UTC
	}
	time.Local = jst

	// Initialize Firestore
	client, err := firestore.NewClient(ctx, cfg.FirestoreProjectID)
	if err != nil {
		log.Fatalf("Failed to create Firestore client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close Firestore client: %v", err)
		}
	}()

	repo := repository.NewFirestoreChatRepository(client)

	// Determine thread ID
	finalThreadID := *threadID
	if finalThreadID == "" {
		// Create new thread
		finalThreadID = fmt.Sprintf("thread-%d", time.Now().Unix())
		log.Printf("Creating new thread: %s", finalThreadID)

		thread := &model.ChatThread{
			ID:            finalThreadID,
			UserID:        *userID,
			Title:         "CLI Created Thread",
			IsPrivate:     false,
			IsArchived:    false,
			LastMessage:   "",
			LastMessageAt: time.Now(),
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}

		if err := createThread(ctx, client, thread); err != nil {
			log.Fatalf("Failed to create thread: %v", err)
		}
		log.Printf("✅ Thread created: %s", finalThreadID)
	} else {
		log.Printf("Using existing thread: %s", finalThreadID)
	}

	// Create message
	msg := &model.ChatMessage{
		ThreadID:  finalThreadID,
		Role:      "user",
		Content:   *message,
		Status:    "unread",
		CreatedAt: time.Now(),
	}

	messageID, err := repo.SaveMessage(ctx, msg)
	if err != nil {
		log.Fatalf("Failed to create message: %v", err)
	}

	log.Printf("✅ Message created successfully!")
	log.Printf("   Thread ID: %s", finalThreadID)
	log.Printf("   Message ID: %s", messageID)
	log.Printf("   Content: %s", *message)
	log.Printf("")
	log.Printf("The message will be processed by the agent via Eventarc trigger (production environment).")
}

func createThread(ctx context.Context, client *firestore.Client, thread *model.ChatThread) error {
	_, err := client.Collection("threads").Doc(thread.ID).Set(ctx, map[string]interface{}{
		"userId":          thread.UserID,
		"title":           thread.Title,
		"isPrivate":       thread.IsPrivate,
		"isArchived":      thread.IsArchived,
		"lastMessage":     thread.LastMessage,
		"lastMessageAt":   thread.LastMessageAt,
		"summary":         thread.Summary,
		"summarizedUntil": thread.SummarizedUntil,
		"createdAt":       thread.CreatedAt,
		"updatedAt":       thread.UpdatedAt,
	})
	return err
}
