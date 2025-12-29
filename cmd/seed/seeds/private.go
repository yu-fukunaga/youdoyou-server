package seeds

import "youdoyou-server/model"

func init() {
	Register("private", SeedData{
		Thread: model.ChatThread{
			ID:            "private-project-thread",
			UserID:        "demo-user-001",
			Title:         "機密プロジェクト X の相談",
			IsPrivate:     true,
			IsArchived:    false,
			LastMessage:   "はい、このスレッドはプライベート設定になっており、あなた以外のユーザーがアクセスすることはできません。レポートの作成が完了次第、こちらで共有します。",
			LastMessageAt: JSTTime(2025, 12, 20, 15, 30),
			CreatedAt:     JSTTime(2025, 12, 20, 15, 0),
			UpdatedAt:     JSTTime(2025, 12, 20, 15, 30),
		},
		Messages: []model.ChatMessage{
			{
				Role:      "user",
				Content:   "次の新製品プロジェクト「Project X」について内密に相談したい。",
				Status:    "completed",
				CreatedAt: JSTTime(2025, 12, 20, 15, 0),
			},
			{
				Role:      "assistant",
				Content:   "かしこまりました。「Project X」に関する情報は機密事項として扱い、厳重に管理いたします。具体的にどのような内容でお困りでしょうか？",
				Status:    "completed",
				CreatedAt: JSTTime(2025, 12, 20, 15, 5),
			},
			{
				Role:      "user",
				Content:   "まずは競合他社の分析から始めたい。A社とB社の最新の動向をまとめてくれるかな？",
				Status:    "completed",
				CreatedAt: JSTTime(2025, 12, 20, 15, 10),
			},
			{
				Role:      "assistant",
				Content:   "承知いたしました。A社は最近、新しいAI統合ツールを発表しましたね。一方B社は、ハードウェアの効率化に注力しています。それぞれの詳細と比較レポートを作成しますので、少々お待ちください。",
				Status:    "completed",
				CreatedAt: JSTTime(2025, 12, 20, 15, 15),
			},
			{
				Role:      "user",
				Content:   "ありがとう。レポートができたら詳しく検討しよう。このスレッドは誰にも見られないように設定しておいて。",
				Status:    "completed",
				CreatedAt: JSTTime(2025, 12, 20, 15, 20),
			},
			{
				Role:      "assistant",
				Content:   "はい、このスレッドはプライベート設定になっており、あなた以外のユーザーがアクセスすることはできません。レポートの作成が完了次第、こちらで共有します。",
				Status:    "completed",
				CreatedAt: JSTTime(2025, 12, 20, 15, 30),
			},
		},
	})
}
