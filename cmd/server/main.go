package main

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/jomei/notionapi"

	"youdoyou-server/config"
	"youdoyou-server/handler"
	"youdoyou-server/repository"
	"youdoyou-server/service"
	"youdoyou-server/tool"
)

func main() {
	ctx := context.Background()

	// ===== LOAD CONFIG =====
	cfg := config.LoadConfig()

	// ===== INITIALIZE CLIENTS =====

	// Firestore
	firestoreClient, err := firestore.NewClient(ctx, cfg.FirestoreProjectID)
	if err != nil {
		log.Fatal(err)
	}
	defer func() {
		if err := firestoreClient.Close(); err != nil {
			log.Printf("Failed to close Firestore client: %v", err)
		}
	}()

	// Google Calendar - Disabled for now
	// calendarService, err := calendar.NewService(ctx, option.WithScopes(calendar.CalendarScope))
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// Notion
	notionClient := notionapi.NewClient(notionapi.Token(cfg.NotionToken))

	// Genkit
	// Create Google AI plugin instance
	googleAI := &googlegenai.GoogleAI{
		APIKey: cfg.GoogleGenaiApiKey,
	}

	g := genkit.Init(ctx,
		genkit.WithPlugins(googleAI),
		genkit.WithDefaultModel("googleai/gemini-3-flash-preview"),
	)

	// ===== INITIALIZE REPOSITORIES =====

	chatRepo := repository.NewFirestoreChatRepository(firestoreClient)
	// calendarRepo := repository.NewGoogleCalendarRepository(calendarService)
	notionRepo := repository.NewNotionRepository(notionClient)

	// ===== INITIALIZE TOOLS =====
	toolFactory := tool.NewToolFactory(g, chatRepo, nil, notionRepo)
	tools := toolFactory.CreateAllTools()

	// ===== INITIALIZE SERVICES =====

	chatService := service.NewChatService(
		chatRepo,
		nil,
		notionRepo,
		g,
		tools,
	)

	// ===== INITIALIZE HANDLERS =====

	chatHandler := handler.NewChatHandler(chatService)

	// ===== HTTP ROUTING =====

	mux := http.NewServeMux()

	// Eventarc トリガー + Cloud Scheduler トリガー
	mux.HandleFunc("POST /v1/chat/process", chatHandler.HandleMessage)

	// ===== START SERVER =====

	log.Printf("Starting server on :%s", cfg.Port)
	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
