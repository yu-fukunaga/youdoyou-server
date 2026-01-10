package handler

// AgentChatRequest は、Schedulerや手動実行時のリクエストボディ定義です。
// Schedulerからの実行などでBodyが空の場合は、ThreadIDが空文字になります。
type AgentChatRequest struct {
	ThreadID string `json:"threadId"`
}

// レスポンス用の構造体は、単純なJSONを返すだけなら定義しなくても
// w.Write([]byte(`{"status":"ok"}`)) で十分ですが、
// 拡張性を考えるなら定義しておいてもOKです。
type AgentChatResponse struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}
