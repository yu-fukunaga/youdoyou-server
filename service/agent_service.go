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

	// 1. Get Thread for Summary
	thread, err := s.chatRepo.GetThread(ctx, threadID)
	if err != nil {
		log.Printf("Warning: Failed to get thread summary: %v", err)
	}
	var summary string
	if thread != nil {
		summary = thread.Summary
	}

	// 2. Get conversation history (all messages)
	// We optimize by fetching once and filtering in memory
	history, err := s.chatRepo.GetConversationHistory(ctx, threadID)
	if err != nil {
		return fmt.Errorf("failed to get history: %w", err)
	}
	log.Printf("Retrieved conversation history for thread %s, %d messages found", threadID, len(history))

	// 3. Find latest unread user message
	var userMsg *model.ChatMessage
	for i := len(history) - 1; i >= 0; i-- {
		if history[i].Role == "user" && history[i].Status == "unread" {
			userMsg = &history[i]
			break
		}
	}

	if userMsg == nil {
		log.Printf("No pending user message found")
		return fmt.Errorf("no pending message found")
	}

	log.Printf("Found user message: %s", userMsg.ID)

	// 4. Mark as processing
	if err := s.chatRepo.UpdateMessageStatus(ctx, threadID, userMsg.ID, "generating"); err != nil {
		log.Printf("Failed to update status to generating: %v", err)
	}
	log.Printf("Message %s status updated to 'generating'", userMsg.ID)

	// 5. Build Genkit Messages (System + History)
	// No longer limiting to last 5 messages, passing full history as requested.
	messages := s.buildHistoryMessages(history, summary)

	// 6. Provide Tools
	// Map map[string]ai.Tool for efficient execution
	toolMap := make(map[string]ai.Tool)
	var toolRefs []ai.ToolRef
	for _, t := range s.tools {
		toolMap[t.Name()] = t
		toolRefs = append(toolRefs, t)
	}

	// 7. Look up the model
	m := genkit.LookupModel(s.genkitClient, "googleai/gemini-3-flash-preview")
	if m == nil {
		if err := s.chatRepo.UpdateMessageStatus(ctx, threadID, userMsg.ID, "error"); err != nil {
			log.Printf("Failed to update status to error: %v", err)
		}
		return fmt.Errorf("model not found")
	}

	// 8. Agent Loop
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
			if updateErr := s.chatRepo.UpdateMessageStatus(ctx, threadID, userMsg.ID, "error"); updateErr != nil {
				log.Printf("Failed to update status to error: %v", updateErr)
			}
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

	// 9. Save response to Firestore
	responseMsg := &model.ChatMessage{
		ThreadID:  threadID,
		Role:      "assistant",
		Content:   finalContent,
		Status:    "completed",
		CreatedAt: time.Now(),
	}

	_, err = s.chatRepo.SaveMessage(ctx, responseMsg)
	if err != nil {
		return fmt.Errorf("failed to save response: %w", err)
	}

	if err := s.chatRepo.UpdateMessageStatus(ctx, threadID, userMsg.ID, "completed"); err != nil {
		log.Printf("Failed to update status to completed: %v", err)
	}

	return nil
}

func (s *AgentService) buildHistoryMessages(history []model.ChatMessage, summary string) []*ai.Message {
	var messages []*ai.Message

	// System Prompt
	systemPrompt := `あなたは業務自動化アシスタント YouDoYou です。
ユーザーの業務をサポートするため、以下の能力があります：
- Notion database へのアクセス（タスク管理）

ユーザーの要望に応じて、必要なツールを使用してサポートしてください。
回答は日本語で、簡潔かつ分かりやすく。`

	if summary != "" {
		systemPrompt = fmt.Sprintf("【これまでの要約】\n%s\n\n%s", summary, systemPrompt)
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
