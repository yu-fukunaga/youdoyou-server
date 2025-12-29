package model

import "time"

type ChatMessage struct {
	ID          string       `firestore:"-"`
	ThreadID    string       `firestore:"threadId"` // Note: Schema doesn't explicitly have threadId in message fields, but it is useful for application logic.
	Role        string       `firestore:"role"`
	Content     string       `firestore:"content"`
	Status      string       `firestore:"status"` // unread, received, generating, completed, error
	Attachments []Attachment `firestore:"attachments,omitempty"`
	AIMetadata  *AIMetadata  `firestore:"aiMetadata,omitempty"`
	CreatedAt   time.Time    `firestore:"createdAt"`
}

type Attachment struct {
	Type     string `firestore:"type"` // image, text, document, audio, video
	URL      string `firestore:"url"`
	MimeType string `firestore:"mimeType"`
	Name     string `firestore:"name"`
	Size     int64  `firestore:"size"`
}

type AIMetadata struct {
	Model        string  `firestore:"model"`
	Usage        AIUsage `firestore:"usage"`
	FinishReason string  `firestore:"finishReason"`
	ResponseID   string  `firestore:"responseId"`
}

type AIUsage struct {
	PromptTokens     float64 `firestore:"promptTokens"` // Using float64 because schema says "number", usually int is fine but safe for dynamic types? Schema type: number. JSON/Firestore numbers often decode to float64 or int64. Let's use int for tokens generally, but float64 is safer for "number" type. Actually tokens are ints. I'll stick to int but checks might fail if it decodes to float. Firestore Go client handles int/float conversion well.
	CompletionTokens float64 `firestore:"completionTokens"`
	TotalTokens      float64 `firestore:"totalTokens"`
}

type ChatThread struct {
	ID              string    `firestore:"-"`
	UserID          string    `firestore:"userId"`
	Title           string    `firestore:"title"`
	IsPrivate       bool      `firestore:"isPrivate"`
	IsArchived      bool      `firestore:"isArchived"`
	LastMessage     string    `firestore:"lastMessage"`
	LastMessageAt   time.Time `firestore:"lastMessageAt"`
	Summary         string    `firestore:"summary"`
	SummarizedUntil time.Time `firestore:"summarizedUntil"`
	CreatedAt       time.Time `firestore:"createdAt"`
	UpdatedAt       time.Time `firestore:"updatedAt"`
}

type ToolCall struct {
	Name       string                 `json:"name" firestore:"name"`
	Parameters map[string]interface{} `json:"parameters" firestore:"parameters"`
	Result     string                 `json:"result" firestore:"result"`
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
