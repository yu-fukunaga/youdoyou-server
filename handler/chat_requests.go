package handler

// ChatProcessRequest represents the JSON body for direct process requests
type ChatProcessRequest struct {
	ThreadID string `json:"threadId"`
}

// ChatProcessResponse represents the empty JSON response
type ChatProcessResponse struct{}

// ChatProcessEvent represents the Eventarc/CloudEvents payload structure
type ChatProcessEvent struct {
	Data struct {
		Value struct {
			Name string `json:"name"` // "projects/.../documents/chat_threads/{threadId}/messages/{msgId}"
		} `json:"value"`
	} `json:"data"`
}
