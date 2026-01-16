package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"youdoyou-server/cmd/seed/seeds"
	"youdoyou-server/config"
	"youdoyou-server/repository"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

func main() {
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
		log.Fatalf("Failed to create client: %v", err)
	}
	defer func() {
		if err := client.Close(); err != nil {
			log.Printf("Failed to close Firestore client: %v", err)
		}
	}()

	repo := repository.NewFirestoreChatRepository(client)

	// Determine which seed names to use
	seedName := "all"
	if len(os.Args) > 1 {
		seedName = os.Args[1]
	}

	var seedsToRun []seeds.SeedData
	if seedName == "all" {
		for _, s := range seeds.Registry {
			seedsToRun = append(seedsToRun, s)
		}
	} else {
		s, ok := seeds.Registry[seedName]
		if !ok {
			log.Fatalf("Seed '%s' not found in registry", seedName)
		}
		seedsToRun = []seeds.SeedData{s}
	}

	for _, seed := range seedsToRun {
		threadID := seed.Thread.ID
		if threadID == "" {
			log.Fatalf("Seed thread must have an ID")
		}

		// --- Idempotency: Delete existing data ---
		if err := deleteThreadAndMessages(ctx, client, threadID); err != nil {
			log.Fatalf("Failed to clear existing data for thread %s: %v", threadID, err)
		}

		// --- Seeding ---
		thread := seed.Thread
		thread.CreatedAt = ensureTime(thread.CreatedAt)
		thread.LastReadAt = ensureTime(thread.LastReadAt)
		thread.MemorizedUntil = ensureTime(thread.MemorizedUntil)

		_, err = client.Collection("threads").Doc(threadID).Set(ctx, map[string]interface{}{
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
		if err != nil {
			log.Fatalf("Failed to create thread: %v", err)
		}

		for _, msg := range seed.Messages {
			msg.ThreadID = threadID
			msg.CreatedAt = ensureTime(msg.CreatedAt)

			msgID, err := repo.SaveMessage(ctx, &msg)
			if err != nil {
				log.Fatalf("Failed to save message: %v", err)
			}
			fmt.Printf("[%s] Saved message: %s\n", threadID, msgID)
		}

		fmt.Printf("Successfully seeded data for thread: %s\n", threadID)
	}
}

func deleteThreadAndMessages(ctx context.Context, client *firestore.Client, threadID string) error {
	threadRef := client.Collection("threads").Doc(threadID)

	// Delete all messages in the sub-collection
	msgIter := threadRef.Collection("messages").DocumentRefs(ctx)
	for {
		docRef, err := msgIter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return err
		}
		if _, err := docRef.Delete(ctx); err != nil {
			return err
		}
	}

	// Delete the thread itself
	_, err := threadRef.Delete(ctx)
	return err
}

func ensureTime(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().In(time.Local)
	}
	return t
}
