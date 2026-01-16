package service

import (
	"context"
	"fmt"
	"log"
	"time"

	"youdoyou-server/model"
	"youdoyou-server/repository"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type AgentService struct {
	chatRepo     repository.ChatRepository
	calendarRepo repository.CalendarRepository
	notionRepo   repository.NotionRepository
	genkitClient *genkit.Genkit
	tools        []ai.Tool
}

func NewAgentService(
	chatRepo repository.ChatRepository,
	calendarRepo repository.CalendarRepository,
	notionRepo repository.NotionRepository,
	genkitClient *genkit.Genkit,
	tools []ai.Tool,
) *AgentService {
	return &AgentService{
		chatRepo:     chatRepo,
		calendarRepo: calendarRepo,
		notionRepo:   notionRepo,
		genkitClient: genkitClient,
		tools:        tools,
	}
}

func (s *AgentService) Chat(ctx context.Context, threadID string) error {
	log.Printf("ProcessMessage started for thread: %s", threadID)

	// 1. Get Thread for SessionMemory
	thread, err := s.chatRepo.GetThread(ctx, threadID)
	if err != nil {
		log.Printf("Warning: Failed to get thread: %v", err)
	}
	var sessionMemory string
	if thread != nil {
		sessionMemory = thread.SessionMemory
	}

	// 2. Get unmemorized messages (messages after memorizedUntil)
	history, err := s.chatRepo.GetUnmemorizedMessages(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get unmemorized messages: %w", err)
	}
	log.Printf("Retrieved %d unmemorized messages for thread %s", len(history), threadID)

	// 3. Build Genkit Messages (System + SessionMemory + History)
	messages := s.buildHistoryMessages(history, sessionMemory)

	// 4. Provide Tools
	// Map map[string]ai.Tool for efficient execution
	toolMap := make(map[string]ai.Tool)
	var toolRefs []ai.ToolRef
	for _, t := range s.tools {
		toolMap[t.Name()] = t
		toolRefs = append(toolRefs, t)
	}

	// 5. Look up the model
	m := genkit.LookupModel(s.genkitClient, "googleai/gemini-3-flash-preview")
	if m == nil {
		return fmt.Errorf("model not found")
	}

	// 6. Agent Loop
	// We will loop until the model stops generating tool calls
	maxTurns := 5
	var finalContent string

	for i := 0; i < maxTurns; i++ {
		log.Printf("Turn %d: Generating...", i)
		resp, err := genkit.Generate(ctx, s.genkitClient,
			ai.WithModel(m),
			// ai.WithConfig(&ai.GenerationCommonConfig{Temperature: 0}),
			ai.WithMessages(messages...),
			ai.WithTools(toolRefs...),
		)
		if err != nil {
			return fmt.Errorf("genkit call failed: %w", err)
		}

		// Append the model's response to history
		messages = append(messages, resp.Message)

		toolReqs := resp.ToolRequests()
		if len(toolReqs) == 0 {
			// No tools called, this is the final response
			finalContent = resp.Text()
			break
		}

		// Handle Tool Calls
		log.Printf("Turn %d: Model requested %d tools", i, len(toolReqs))
		var toolParts []*ai.Part
		for _, req := range toolReqs {
			t, ok := toolMap[req.Name]
			if !ok {
				log.Printf("Tool not found: %s", req.Name)
				toolParts = append(toolParts, ai.NewToolResponsePart(&ai.ToolResponse{
					Name:   req.Name,
					Ref:    req.Ref,
					Output: fmt.Sprintf("Error: Tool %s not found", req.Name),
				}))
				continue
			}

			// Run Tool
			log.Printf("Running tool: %s", req.Name)
			out, err := t.RunRaw(ctx, req.Input)
			if err != nil {
				log.Printf("Tool execution failed: %v", err)
				toolParts = append(toolParts, ai.NewToolResponsePart(&ai.ToolResponse{
					Name:   req.Name,
					Ref:    req.Ref,
					Output: fmt.Sprintf("Error: %v", err),
				}))
				continue
			}

			toolParts = append(toolParts, ai.NewToolResponsePart(&ai.ToolResponse{
				Name:   req.Name,
				Ref:    req.Ref,
				Output: out,
			}))
		}

		// Append Tool Response Message
		toolMsg := ai.NewMessage(ai.RoleTool, nil, toolParts...)
		messages = append(messages, toolMsg)
	}

	if finalContent == "" {
		finalContent = "申し訳ありません、処理を完了できませんでした (Max turns reached)."
	}

	// 7. Save response to Firestore
	responseMsg := &model.ChatMessage{
		ThreadID:  threadID,
		Role:      "assistant",
		Content:   finalContent,
		CreatedAt: time.Now(),
	}

	_, err = s.chatRepo.SaveMessage(ctx, responseMsg)
	if err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	log.Printf("Response saved successfully for thread %s", threadID)
	return nil
}

func (s *AgentService) buildHistoryMessages(history []model.ChatMessage, sessionMemory string) []*ai.Message {
	var messages []*ai.Message

	// System Prompt
	systemPrompt := `あなたは業務自動化アシスタント YouDoYou です。
ユーザーの業務をサポートするため、以下の能力があります：
- Notion database へのアクセス（タスク管理）

ユーザーの要望に応じて、必要なツールを使用してサポートしてください。
回答は日本語で、簡潔かつ分かりやすく。`

	if sessionMemory != "" {
		systemPrompt = fmt.Sprintf("【これまでの要約】\n%s\n\n%s", sessionMemory, systemPrompt)
	}

	messages = append(messages, ai.NewSystemTextMessage(systemPrompt))

	// History
	for _, msg := range history {
		if msg.Role == "user" {
			messages = append(messages, ai.NewUserTextMessage(msg.Content))
		} else {
			// Assistant message
			messages = append(messages, ai.NewModelTextMessage(msg.Content))
		}
	}

	return messages
}
