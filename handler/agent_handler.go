package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"

	"youdoyou-server/service"

	"github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"google.golang.org/protobuf/proto"
)

type AgentHandler struct {
	agentService *service.AgentService
}

func NewAgentHandler(agentService *service.AgentService) *AgentHandler {
	return &AgentHandler{agentService: agentService}
}

// ==========================================
// 1. Generic Agent Chat (Scheduler / Manual)
// URL: POST /v1/agent/chat
// ==========================================
func (h *AgentHandler) HandleAgentChat(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// 定義した AgentChatRequest を使用
	var req AgentChatRequest

	// Bodyがあればデコード (Schedulerは空ボディで来る可能性があるためエラーは許容)
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&req)
	}

	var err error

	if req.ThreadID == "" {
		log.Printf("⏰ Agent Chat Triggered (Initiate Mode)")
	}

	err = h.agentService.Chat(ctx, req.ThreadID)
	if err != nil {
		log.Printf("❌ Agent run failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 成功レスポンス
	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(AgentChatResponse{Status: "ok"})
	if err != nil {
		log.Printf("Failed to encode response: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}
}

// ==========================================
// 2. Firestore Trigger (Eventarc)
// URL: POST /hooks/firestore
// ==========================================
func (h *AgentHandler) HandleFirestoreTrigger(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		log.Printf("Failed to read body: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	// ★ 自前定義の ChatProcessEvent ではなく、公式ライブラリを使用
	var eventData firestoredata.DocumentEventData
	if err := proto.Unmarshal(body, &eventData); err != nil {
		log.Printf("Failed to unmarshal event: %v", err)
		http.Error(w, "Bad Request", http.StatusBadRequest)
		return
	}

	fullPath := eventData.GetValue().GetName()
	log.Printf("🔥 Firestore Triggered: %s", fullPath)

	threadID := extractThreadIDFromPath(fullPath)
	if threadID == "" {
		// 対象外のパスなら正常終了扱いで無視
		w.WriteHeader(http.StatusOK)
		return
	}

	// Check if the message is from a user (only process user messages)
	fields := eventData.GetValue().GetFields()
	if roleField, ok := fields["role"]; ok {
		role := roleField.GetStringValue()
		if role != "user" {
			// assistantメッセージなら処理せず正常終了
			log.Printf("Skipping non-user message (role=%s)", role)
			w.WriteHeader(http.StatusOK)
			return
		}
	}

	if err := h.agentService.Chat(ctx, threadID); err != nil {
		log.Printf("❌ Firestore trigger failed: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// ==========================================
// Helper Functions
// ==========================================

func extractThreadIDFromPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "threads" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
