package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"youdoyou-server/service"
)

type ChatHandler struct {
	chatService *service.ChatService
}

func NewChatHandler(chatService *service.ChatService) *ChatHandler {
	return &ChatHandler{chatService: chatService}
}

// Eventarc トリガー + Cloud Scheduler トリガー の両対応
func (h *ChatHandler) HandleMessage(w http.ResponseWriter, r *http.Request) {
	// 1. Check for Eventarc (CloudEvents) header
	// Google CloudEvents usually have "Ce-Type" header
	ceType := r.Header.Get("Ce-Type")

	var threadID string

	if ceType != "" {
		// --- Case A: Eventarc / CloudEvents ---
		// Expected type: google.cloud.firestore.document.v1.written (or created)
		// Payload structure differs from simple JSON
		var event ChatProcessEvent

		if err := json.NewDecoder(r.Body).Decode(&event); err != nil {
			http.Error(w, "invalid cloud event", http.StatusBadRequest)
			return
		}

		// Extract ThreadID from Document Name/Path
		// Format: .../chat_threads/{threadId}/messages/{messageId}
		threadID = extractThreadIDFromPath(event.Data.Value.Name)
		if threadID == "" {
			// Note: It might be an event for a different collection if filtering is weak,
			// or parsing failed. We just log/ignore or return 200 to ack.
			// Returning 200 OK to avoid Eventarc retries on irrelevant events.
			w.WriteHeader(http.StatusOK)
			return
		}

	} else {
		// --- Case B: Direct Call / Cloud Scheduler ---
		// Simple JSON: {"threadId": "..."}
		var req ChatProcessRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "invalid request", http.StatusBadRequest)
			return
		}
		threadID = req.ThreadID
	}

	if threadID == "" {
		http.Error(w, "threadId missing", http.StatusBadRequest)
		return
	}

	// Service を呼び出す（non-blocking）
	go func() {
		log.Printf("Starting process for thread: %s", threadID)
		if err := h.chatService.ProcessMessage(context.Background(), threadID); err != nil {
			log.Printf("Error processing message: %v", err)
		} else {
			log.Printf("Successfully processed message for thread: %s", threadID)
		}
	}()

	// 即座に response を返す（非同期処理）
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusAccepted) // 202 Accepted
	json.NewEncoder(w).Encode(ChatProcessResponse{})
}

func extractThreadIDFromPath(path string) string {
	// path example: projects/my-p/databases/(default)/documents/threads/THREAD_123/messages/MSG_456
	// Simple regex or string splitting
	// Let's use strings.Split
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "threads" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
