package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"youdoyou-server/config"
	"youdoyou-server/repository"

	"cloud.google.com/go/firestore"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	// Initialize Firestore
	client, err := firestore.NewClient(ctx, cfg.FirestoreProjectID)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	repo := repository.NewFirestoreChatRepository(client)

	threadID := "test-thread-e2e"
	if len(os.Args) > 1 {
		threadID = os.Args[1]
	}

	history, err := repo.GetConversationHistory(ctx, threadID)
	if err != nil {
		log.Fatalf("Failed to get history: %v", err)
	}

	fmt.Printf("--- Conversation History for %s ---\n", threadID)
	for _, msg := range history {
		fmt.Printf("[%s] %s (%s) ID:%s : %s\n", msg.CreatedAt.Format("15:04:05"), msg.Role, msg.Status, msg.ID, msg.Content)

	}
	fmt.Printf("---------------------------------------\n")
}
