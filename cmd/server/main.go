package main

import (
	"context"
	"log"
	"net/http"

	"cloud.google.com/go/firestore"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jomei/notionapi"

	"youdoyou-server/config"
	"youdoyou-server/handler"
	"youdoyou-server/repository"
	"youdoyou-server/service"
	"youdoyou-server/tool"
)

func main() {
	ctx := context.Background()
	cfg := config.LoadConfig()

	// --- 1. Clients & SDKs Initialization ---

	firestoreClient, err := firestore.NewClient(ctx, cfg.FirestoreProjectID)
	if err != nil {
		log.Fatal(err)
	}
	defer firestoreClient.Close()

	// Notion
	notionClient := notionapi.NewClient(notionapi.Token(cfg.NotionToken))

	// Genkit
	googleAI := &googlegenai.GoogleAI{
		APIKey: cfg.GoogleGenaiApiKey,
	}

	g := genkit.Init(ctx,
		genkit.WithPlugins(googleAI),
		genkit.WithDefaultModel("googleai/gemini-3-flash-preview"),
	)

	// --- 2. Dependency Injection (DI) ---

	chatRepo := repository.NewFirestoreChatRepository(firestoreClient)
	notionRepo := repository.NewNotionRepository(notionClient)

	toolFactory := tool.NewToolFactory(g, chatRepo, nil, notionRepo)
	tools := toolFactory.CreateAllTools()

	agentService := service.NewAgentService(chatRepo, nil, notionRepo, g, tools)
	agentHandler := handler.NewAgentHandler(agentService)

	// --- 3. HTTP Routing with chi ---

	r := chi.NewRouter()

	// A. グローバルミドルウェア (chi標準のもの)
	r.Use(chimiddleware.Logger)    // 全リクエストをログ出力
	r.Use(chimiddleware.Recoverer) // パニック発生時にサーバー停止を防ぐ
	r.Use(chimiddleware.RealIP)    // プロキシ配下でもクライアントIPを正しく取得

	// B. ルーティングの構築
	r.Route("/v1", func(r chi.Router) {

		// デバッグ/手動実行用 (人間が叩く)
		r.Post("/agent/chat", agentHandler.HandleAgentChat)

		// システム連携用 (Eventarc / Scheduler)
		r.Post("/hooks/firestore", agentHandler.HandleFirestoreTrigger)

		// ヘルスチェック
		r.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			_, err := w.Write([]byte("pong"))
			if err != nil {
				log.Printf("Failed to write response: %v", err)
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
		})
	})

	// --- 4. Start Server ---
	log.Printf("Starting server on :%s", cfg.Port)
	// nosemgrep: go.lang.security.audit.net.use-tls.use-tls
	if err := http.ListenAndServe(":"+cfg.Port, r); err != nil {
		log.Fatal(err)
	}
}
