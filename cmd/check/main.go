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
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close Firestore client: %v", err)
		}
	}()

	repo := repository.NewFirestoreChatRepository(client)

	threadID := "test-thread-e2e"
	if len(os.Args) > 1 {
		threadID = os.Args[1]
	}

	history, err := repo.GetUnmemorizedMessages(ctx, threadID)
	if err != nil {
		log.Fatalf("Failed to get unmemorized messages: %v", err)
	}

	fmt.Printf("--- Unmemorized Messages for %s ---\n", threadID)
	for _, msg := range history {
		fmt.Printf("[%s] %s ID:%s : %s\n", msg.CreatedAt.Format("15:04:05"), msg.Role, msg.ID, msg.Content)

	}
	fmt.Printf("---------------------------------------\n")
}
