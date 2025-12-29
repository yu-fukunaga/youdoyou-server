# YouDoYou Server: ディレクトリ構造 & 実装（パターン A）

## ディレクトリ構造

```
youdoyou-server/
├── cmd/
│   └── server/
│       └── main.go                  ← エントリーポイント（DI, Bootstrap）
│
├── handler/
│   └── chat_handler.go              ← HTTP handler（Eventarc, Cloud Scheduler受け取り）
│
├── service/
│   ├── chat_service.go              ← Workflow orchestration
│   └── workflow_router.go           ← Workflow routing logic
│
├── repository/
│   ├── firestore.go                 ← Firestore (ChatRepository)
│   ├── calendar.go                  ← Google Calendar API
│   ├── notion.go                    ← Notion API
│   └── interfaces.go                ← Repository interfaces
│
├── tool/
│   ├── calendar_tool.go             ← Calendar Tool定義
│   ├── notion_tool.go               ← Notion Tool定義
│   └── tool_factory.go              ← Tool 生成
│
├── model/
│   └── types.go                     ← Domain types (ChatMessage, etc)
│
├── config/
│   └── config.go                    ← Configuration
│
├── test/
│   ├── mocks.go                     ← Mock implementations
│   └── fixtures.go                  ← Test data
│
├── go.mod
├── go.sum
└── Dockerfile
```

---

## 1. model/types.go（Domain types）

```go
package model

import "time"

type ChatMessage struct {
    ID        string
    ThreadID  string
    Role      string  // "user" or "assistant"
    Content   string
    ToolCalls []ToolCall
    Status    string  // "pending" | "processing" | "completed" | "error"
    CreatedAt time.Time
}

type ToolCall struct {
    Name       string                 `json:"name"`
    Parameters map[string]interface{} `json:"parameters"`
    Result     string                 `json:"result"`
}

type ChatThread struct {
    ID        string
    UserID    string
    CreatedAt time.Time
    UpdatedAt time.Time
}

type WorkflowRequest struct {
    ThreadID string
    UserMsg  string
}

type WorkflowResponse struct {
    Response  string
    ToolCalls []ToolCall
}

// Calendar types
type CalendarEvent struct {
    ID        string
    Summary   string
    StartTime time.Time
    EndTime   time.Time
    Location  string
}

// Notion types
type NotionPage struct {
    ID         string
    Title      string
    Properties map[string]interface{}
}
```

---

## 2. repository/interfaces.go（Repository インターフェース）

```go
package repository

import (
    "context"
    "youdoyou/model"
)

// ChatRepository - Firestore
type ChatRepository interface {
    GetLatestUserMessage(ctx context.Context, threadID string) (*model.ChatMessage, error)
    GetConversationHistory(ctx context.Context, threadID string) ([]model.ChatMessage, error)
    SaveMessage(ctx context.Context, message *model.ChatMessage) (string, error)
    UpdateMessageStatus(ctx context.Context, messageID string, status string) error
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
```

---

## 3. repository/firestore.go（Firestore 実装）

```go
package repository

import (
    "context"
    "fmt"
    "cloud.google.com/go/firestore"
    "youdoyou/model"
)

type FirestoreChatRepository struct {
    client *firestore.Client
}

func NewFirestoreChatRepository(client *firestore.Client) ChatRepository {
    return &FirestoreChatRepository{client: client}
}

func (r *FirestoreChatRepository) GetLatestUserMessage(ctx context.Context, threadID string) (*model.ChatMessage, error) {
    query := r.client.Collection("chat_threads").Doc(threadID).
        Collection("messages").
        Where("role", "==", "user").
        Where("status", "==", "pending").
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
    docs[0].DataTo(&msg)
    msg.ID = docs[0].Ref.ID

    return &msg, nil
}

func (r *FirestoreChatRepository) GetConversationHistory(ctx context.Context, threadID string) ([]model.ChatMessage, error) {
    query := r.client.Collection("chat_threads").Doc(threadID).
        Collection("messages").
        OrderBy("createdAt", firestore.Asc)

    docs, err := query.Documents(ctx).GetAll()
    if err != nil {
        return nil, err
    }

    var messages []model.ChatMessage
    for _, doc := range docs {
        var msg model.ChatMessage
        doc.DataTo(&msg)
        msg.ID = doc.Ref.ID
        messages = append(messages, msg)
    }

    return messages, nil
}

func (r *FirestoreChatRepository) SaveMessage(ctx context.Context, message *model.ChatMessage) (string, error) {
    docRef, _, err := r.client.Collection("chat_threads").Doc(message.ThreadID).
        Collection("messages").
        Add(ctx, message)

    if err != nil {
        return "", err
    }

    return docRef.ID, nil
}

func (r *FirestoreChatRepository) UpdateMessageStatus(ctx context.Context, messageID string, status string) error {
    // messageID から threadID を特定する必要がある（例：メタデータに保持）
    // 簡略化のため省略
    return nil
}
```

---

## 4. repository/calendar.go（Google Calendar 実装）

```go
package repository

import (
    "context"
    "time"
    "google.golang.org/api/calendar/v3"
    "youdoyou/model"
)

type GoogleCalendarRepository struct {
    service *calendar.Service
}

func NewGoogleCalendarRepository(service *calendar.Service) CalendarRepository {
    return &GoogleCalendarRepository{service: service}
}

func (r *GoogleCalendarRepository) GetEvents(ctx context.Context, timeRange string, timezone string) ([]model.CalendarEvent, error) {
    // timeRange: "this week", "next 7 days", "today" etc
    startTime, endTime := parseTimeRange(timeRange, timezone)

    events, err := r.service.Events.List("primary").
        TimeMin(startTime.Format(time.RFC3339)).
        TimeMax(endTime.Format(time.RFC3339)).
        Context(ctx).
        Do()

    if err != nil {
        return nil, err
    }

    var result []model.CalendarEvent
    for _, item := range events.Items {
        start, _ := time.Parse(time.RFC3339, item.Start.DateTime)
        end, _ := time.Parse(time.RFC3339, item.End.DateTime)

        result = append(result, model.CalendarEvent{
            ID:        item.Id,
            Summary:   item.Summary,
            StartTime: start,
            EndTime:   end,
            Location:  item.Location,
        })
    }

    return result, nil
}

func parseTimeRange(timeRange string, timezone string) (time.Time, time.Time) {
    // Implementation of time range parsing
    // e.g., "this week", "next 7 days" → start, end
    loc, _ := time.LoadLocation(timezone)
    now := time.Now().In(loc)

    switch timeRange {
    case "today":
        start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
        end := start.AddDate(0, 0, 1)
        return start, end
    case "this week":
        start := now.AddDate(0, 0, -int(now.Weekday()))
        end := start.AddDate(0, 0, 7)
        return start, end
    case "next week":
        start := now.AddDate(0, 0, -int(now.Weekday())+7)
        end := start.AddDate(0, 0, 7)
        return start, end
    case "next 7 days":
        start := now
        end := start.AddDate(0, 0, 7)
        return start, end
    default:
        return now, now.AddDate(0, 0, 7)
    }
}
```

---

## 5. repository/notion.go（Notion 実装）

```go
package repository

import (
    "context"
    "github.com/jomei/notionapi"
    "youdoyou/model"
)

type NotionRepository struct {
    client *notionapi.Client
}

func NewNotionRepository(client *notionapi.Client) NotionRepository {
    return &NotionRepository{client: client}
}

func (r *NotionRepository) QueryDatabase(ctx context.Context, databaseID string, filter map[string]interface{}) ([]model.NotionPage, error) {
    // Convert databaseID to notionapi.DatabaseID
    dbID := notionapi.DatabaseID(databaseID)

    // Build query
    query := &notionapi.DatabaseQueryRequest{
        Filter: buildNotionFilter(filter),
    }

    response, err := r.client.Database.Query(ctx, dbID, query)
    if err != nil {
        return nil, err
    }

    var result []model.NotionPage
    for _, page := range response.Results {
        result = append(result, model.NotionPage{
            ID:         page.ID.String(),
            Title:      extractTitle(page),
            Properties: page.Properties,
        })
    }

    return result, nil
}

func (r *NotionRepository) CreatePage(ctx context.Context, databaseID string, properties map[string]interface{}) (string, error) {
    dbID := notionapi.DatabaseID(databaseID)

    createRequest := &notionapi.PageCreateRequest{
        Parent: notionapi.Parent{
            DatabaseID: dbID,
        },
        Properties: buildNotionProperties(properties),
    }

    page, err := r.client.Page.Create(ctx, createRequest)
    if err != nil {
        return "", err
    }

    return page.ID.String(), nil
}

func buildNotionFilter(filter map[string]interface{}) interface{} {
    // Implementation: convert generic filter to notionapi filter
    return nil
}

func buildNotionProperties(props map[string]interface{}) notionapi.Properties {
    // Implementation: convert generic properties to notionapi properties
    return notionapi.Properties{}
}

func extractTitle(page *notionapi.Page) string {
    // Extract title from page properties
    return ""
}
```

---

## 6. tool/calendar_tool.go（Calendar Tool定義）

```go
package tool

import (
    "context"
    "encoding/json"
    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "youdoyou/model"
    "youdoyou/repository"
)

type CalendarToolInput struct {
    TimeRange string `json:"timeRange" jsonschema_description:"Time range like 'today', 'this week', 'next 7 days'"`
    Timezone  string `json:"timezone" jsonschema_description:"Timezone like 'Asia/Tokyo'"`
}

func CreateCalendarTool(calendarRepo repository.CalendarRepository) *ai.Tool {
    return genkit.DefineTool(
        "getCalendar",
        "Retrieves calendar events for the specified time range",
        func(ctx context.Context, input CalendarToolInput) (string, error) {
            // Repository を使って Calendar データを取得
            events, err := calendarRepo.GetEvents(ctx, input.TimeRange, input.Timezone)
            if err != nil {
                return "", err
            }

            // Tool の結果は文字列で返す（AI が読める形式）
            result := formatCalendarResult(events)
            return result, nil
        },
    )
}

func formatCalendarResult(events []model.CalendarEvent) string {
    // Format events as human-readable string for AI
    var result string
    for _, event := range events {
        result += event.Summary + " (" + event.StartTime.Format("2006-01-02 15:04") + ")\n"
    }
    return result
}
```

---

## 7. tool/notion_tool.go（Notion Tool定義）

```go
package tool

import (
    "context"
    "encoding/json"
    "github.com/firebase/genkit/go/ai"
    "github.com/firebase/genkit/go/genkit"
    "youdoyou/model"
    "youdoyou/repository"
)

type NotionToolInput struct {
    DatabaseID string                 `json:"databaseId" jsonschema_description:"Notion database ID"`
    Filter     map[string]interface{} `json:"filter" jsonschema_description:"Query filter"`
}

func CreateNotionTool(notionRepo repository.NotionRepository) *ai.Tool {
    return genkit.DefineTool(
        "getNotion",
        "Queries Notion database and returns pages matching the filter",
        func(ctx context.Context, input NotionToolInput) (string, error) {
            // Repository を使って Notion データを取得
            pages, err := notionRepo.QueryDatabase(ctx, input.DatabaseID, input.Filter)
            if err != nil {
                return "", err
            }

            result := formatNotionResult(pages)
            return result, nil
        },
    )
}

type NotionCreateInput struct {
    DatabaseID string                 `json:"databaseId" jsonschema_description:"Notion database ID"`
    Properties map[string]interface{} `json:"properties" jsonschema_description:"Page properties to create"`
}

func CreateNotionWriteTool(notionRepo repository.NotionRepository) *ai.Tool {
    return genkit.DefineTool(
        "createNotionPage",
        "Creates a new page in Notion database",
        func(ctx context.Context, input NotionCreateInput) (string, error) {
            pageID, err := notionRepo.CreatePage(ctx, input.DatabaseID, input.Properties)
            if err != nil {
                return "", err
            }

            return "Page created with ID: " + pageID, nil
        },
    )
}

func formatNotionResult(pages []model.NotionPage) string {
    var result string
    for _, page := range pages {
        result += page.Title + " (ID: " + page.ID + ")\n"
    }
    return result
}
```

---

## 8. tool/tool_factory.go（Tool 生成）

```go
package tool

import (
    "github.com/firebase/genkit/go/ai"
    "youdoyou/repository"
)

type ToolFactory struct {
    chatRepo     repository.ChatRepository
    calendarRepo repository.CalendarRepository
    notionRepo   repository.NotionRepository
}

func NewToolFactory(
    chatRepo repository.ChatRepository,
    calendarRepo repository.CalendarRepository,
    notionRepo repository.NotionRepository,
) *ToolFactory {
    return &ToolFactory{
        chatRepo:     chatRepo,
        calendarRepo: calendarRepo,
        notionRepo:   notionRepo,
    }
}

// 複数の Tool を一度に返す
func (f *ToolFactory) CreateAllTools() []*ai.Tool {
    return []*ai.Tool{
        CreateCalendarTool(f.calendarRepo),
        CreateNotionTool(f.notionRepo),
        CreateNotionWriteTool(f.notionRepo),
    }
}

// 特定の Tool だけ返す
func (f *ToolFactory) CreateToolsByDependencies(deps []string) []*ai.Tool {
    var tools []*ai.Tool

    for _, dep := range deps {
        switch dep {
        case "calendar":
            tools = append(tools, CreateCalendarTool(f.calendarRepo))
        case "notion":
            tools = append(tools,
                CreateNotionTool(f.notionRepo),
                CreateNotionWriteTool(f.notionRepo),
            )
        }
    }

    return tools
}
```

---

## 9. service/chat_service.go（メイン Service）

```go
package service

import (
    "context"
    "fmt"
    "github.com/firebase/genkit/go/genkit"
    "youdoyou/model"
    "youdoyou/repository"
    "youdoyou/tool"
)

type ChatService struct {
    chatRepo     repository.ChatRepository
    calendarRepo repository.CalendarRepository
    notionRepo   repository.NotionRepository
    toolFactory  *tool.ToolFactory
    genkitClient genkit.Client  // Genkit client
}

func NewChatService(
    chatRepo repository.ChatRepository,
    calendarRepo repository.CalendarRepository,
    notionRepo repository.NotionRepository,
    genkitClient genkit.Client,
) *ChatService {
    toolFactory := tool.NewToolFactory(chatRepo, calendarRepo, notionRepo)

    return &ChatService{
        chatRepo:     chatRepo,
        calendarRepo: calendarRepo,
        notionRepo:   notionRepo,
        toolFactory:  toolFactory,
        genkitClient: genkitClient,
    }
}

func (s *ChatService) ProcessMessage(ctx context.Context, threadID string) error {
    // 1. Get latest user message
    userMsg, err := s.chatRepo.GetLatestUserMessage(ctx, threadID)
    if err != nil {
        return fmt.Errorf("failed to get message: %w", err)
    }

    // 2. Mark as processing
    s.chatRepo.UpdateMessageStatus(ctx, userMsg.ID, "processing")

    // 3. Get conversation history
    history, err := s.chatRepo.GetConversationHistory(ctx, threadID)
    if err != nil {
        return fmt.Errorf("failed to get history: %w", err)
    }

    // 4. Build prompt from message and history
    prompt := s.buildPrompt(userMsg.Content, history)

    // 5. Create all tools (Tool factory)
    tools := s.toolFactory.CreateAllTools()

    // 6. Call Genkit with tools
    // Genkit が自動的に AI に Tool を渡し、AI が必要に応じて実行
    resp, err := genkit.Generate(ctx, s.genkitClient,
        genkit.WithPrompt(prompt),
        genkit.WithTools(tools...),
    )
    if err != nil {
        s.chatRepo.UpdateMessageStatus(ctx, userMsg.ID, "error")
        return fmt.Errorf("genkit call failed: %w", err)
    }

    // 7. Create response message
    responseMsg := &model.ChatMessage{
        ThreadID:  threadID,
        Role:      "assistant",
        Content:   resp.Text(),
        ToolCalls: extractToolCalls(resp),
        Status:    "completed",
    }

    // 8. Save response to Firestore
    _, err = s.chatRepo.SaveMessage(ctx, responseMsg)
    if err != nil {
        return fmt.Errorf("failed to save response: %w", err)
    }

    return nil
}

func (s *ChatService) buildPrompt(userMsg string, history []model.ChatMessage) string {
    systemPrompt := `あなたは業務自動化アシスタント YouDoYou です。
ユーザーの業務をサポートするため、以下の能力があります：
- Google Calendar へのアクセス（スケジュール確認）
- Notion database へのアクセス（タスク管理）

ユーザーの要望に応じて、必要なツールを使用してサポートしてください。
回答は日本語で、簡潔かつ分かりやすく。`

    // Build conversation context
    conversationContext := ""
    for _, msg := range history {
        role := msg.Role
        if role == "user" {
            conversationContext += "User: " + msg.Content + "\n"
        } else {
            conversationContext += "Assistant: " + msg.Content + "\n"
        }
    }

    return systemPrompt + "\n\n" + conversationContext + "\nUser: " + userMsg
}

func extractToolCalls(resp *genkit.Response) []model.ToolCall {
    // Extract tool calls from Genkit response
    // Implementation depends on Genkit's response structure
    return []model.ToolCall{}
}
```

---

## 10. handler/chat_handler.go（HTTP Handler）

```go
package handler

import (
    "encoding/json"
    "net/http"
    "youdoyou/service"
)

type ChatHandler struct {
    chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
    return &ChatHandler{chatService: chatService}
}

// Eventarc トリガー + Cloud Scheduler トリガー の両対応
func (h *ChatHandler) HandleMessage(w http.ResponseWriter, r *http.Request) {
    var req struct {
        ThreadID string `json:"threadId"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "invalid request", http.StatusBadRequest)
        return
    }

    // Service を呼び出す（non-blocking）
    go func() {
        if err := h.chatService.ProcessMessage(r.Context(), req.ThreadID); err != nil {
            // Logging
        }
    }()

    // 即座に response を返す（非同期処理）
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "status": "processing",
        "threadId": req.ThreadID,
    })
}
```

---

## 11. cmd/server/main.go（Bootstrap & DI）

```go
package main

import (
    "context"
    "log"
    "net/http"

    "cloud.google.com/go/firestore"
    "github.com/firebase/genkit/go/genkit"
    "google.golang.org/api/calendar/v3"
    "google.golang.org/api/option"
    "github.com/jomei/notionapi"

    "youdoyou/handler"
    "youdoyou/repository"
    "youdoyou/service"
)

func main() {
    ctx := context.Background()

    // ===== INITIALIZE CLIENTS =====

    // Firestore
    firestoreClient, err := firestore.NewClient(ctx, "youdoyou-project")
    if err != nil {
        log.Fatal(err)
    }
    defer firestoreClient.Close()

    // Google Calendar
    calendarService, err := calendar.NewService(ctx, option.WithScopes(calendar.CalendarScope))
    if err != nil {
        log.Fatal(err)
    }

    // Notion
    notionClient := notionapi.NewClient(notionapi.Token("your-notion-token"))

    // Genkit
    genkitClient, err := genkit.Init(ctx)
    if err != nil {
        log.Fatal(err)
    }

    // ===== INITIALIZE REPOSITORIES =====

    chatRepo := repository.NewFirestoreChatRepository(firestoreClient)
    calendarRepo := repository.NewGoogleCalendarRepository(calendarService)
    notionRepo := repository.NewNotionRepository(notionClient)

    // ===== INITIALIZE SERVICES =====

    chatService := service.NewChatService(
        chatRepo,
        calendarRepo,
        notionRepo,
        genkitClient,
    )

    // ===== INITIALIZE HANDLERS =====

    chatHandler := handler.NewChatHandler(chatService)

    // ===== HTTP ROUTING =====

    mux := http.NewServeMux()

    // Eventarc トリガー + Cloud Scheduler トリガー
    mux.HandleFunc("POST /v1/chat/process", chatHandler.HandleMessage)

    // ===== START SERVER =====

    log.Println("Starting server on :8080")
    if err := http.ListenAndServe(":8080", mux); err != nil {
        log.Fatal(err)
    }
}
```

---

## 12. test/mocks.go（Mock implementations）

```go
package test

import (
    "context"
    "youdoyou/model"
    "youdoyou/repository"
)

// Mock ChatRepository
type MockChatRepository struct {
    messages map[string]*model.ChatMessage
}

func (m *MockChatRepository) GetLatestUserMessage(ctx context.Context, threadID string) (*model.ChatMessage, error) {
    // Mock implementation
    return &model.ChatMessage{
        ID:       "msg_1",
        ThreadID: threadID,
        Content:  "テストメッセージ",
        Role:     "user",
    }, nil
}

func (m *MockChatRepository) GetConversationHistory(ctx context.Context, threadID string) ([]model.ChatMessage, error) {
    return []model.ChatMessage{}, nil
}

func (m *MockChatRepository) SaveMessage(ctx context.Context, message *model.ChatMessage) (string, error) {
    return "mock_id", nil
}

func (m *MockChatRepository) UpdateMessageStatus(ctx context.Context, messageID string, status string) error {
    return nil
}

// Mock CalendarRepository
type MockCalendarRepository struct {
    events []model.CalendarEvent
}

func (m *MockCalendarRepository) GetEvents(ctx context.Context, timeRange string, timezone string) ([]model.CalendarEvent, error) {
    return m.events, nil
}

// Mock NotionRepository
type MockNotionRepository struct{}

func (m *MockNotionRepository) QueryDatabase(ctx context.Context, databaseID string, filter map[string]interface{}) ([]model.NotionPage, error) {
    return []model.NotionPage{}, nil
}

func (m *MockNotionRepository) CreatePage(ctx context.Context, databaseID string, properties map[string]interface{}) (string, error) {
    return "page_id", nil
}
```

---

## 実行フロー

```
1. Eventarc or Cloud Scheduler
   ↓
   HTTP POST /v1/chat/process
   {
     "threadId": "thread_abc123"
   }

2. ChatHandler.HandleMessage()
   ↓ (Repository から message を取得)

3. ChatService.ProcessMessage()
   ├─ chatRepo.GetLatestUserMessage()
   ├─ chatRepo.GetConversationHistory()
   ├─ toolFactory.CreateAllTools()
   │  ├─ Calendar Tool（calendarRepo 注入）
   │  ├─ Notion Tool（notionRepo 注入）
   │  └─ Tool が必要なら AI が自動実行
   └─ genkit.Generate()
      ↓

4. Genkit (Google Gemini)
   ├─ Analyze user message
   ├─ Decide if tools needed
   ├─ Auto-execute tools if needed
   │  ├─ Tool が Repository を呼ぶ
   │  └─ 結果を AI に返す
   └─ Generate response

5. ChatService
   └─ chatRepo.SaveMessage() → Firestore

6. Firestore listener（SwiftUI）
   └─ Response を表示
```

---

## Pattern A の利点（このアーキテクチャ）

✅ **Repository と Tool が統一**
  - 同じ data access layer を使う
  - 重複がない

✅ **シンプル**
  - Service から Tool factory へ Repository を渡すだけ
  - Tool 内で Repository を直接使える

✅ **テスト可能**
  - Mock Repository を渡して Service をテスト
  - Tool も同じ Mock Repository でテスト

```go
func TestChatService(t *testing.T) {
    mockRepo := &MockChatRepository{...}
    service := NewChatService(mockRepo, ...)

    err := service.ProcessMessage(context.Background(), "thread_1")
    if err != nil {
        t.Fail()
    }
}
```

✅ **拡張が簡単**
  - 新しい Tool を追加 = tool_factory に追加 + DefineTool で定義
  - 新しい Repository = repository インターフェース実装

---

## 次のステップ

これで良いですか？

もしくは：
- 実装の詳細（Genkit との連携、エラーハンドリング等）
- テスト例
- Deployment（Dockerfile, Cloud Run設定）

どれが欲しいですか？
