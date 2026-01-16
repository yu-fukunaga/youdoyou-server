package seeds

import "youdoyou-server/model"

func init() {
	Register("private", SeedData{
		Thread: model.ChatThread{
			ID:             "private-project-thread",
			UserID:         "demo-user-001",
			FirstMessage:   "次の新製品プロジェクト「Project X」について内密に相談したい。",
			UnreadCount:    0,
			LastReadAt:     JSTTime(2025, 12, 20, 15, 30),
			ReplyCount:     6,
			IsPrivate:      true,
			IsArchived:     false,
			SessionMemory:  "",
			MemorizedUntil: JSTTime(1970, 1, 1, 0, 0),
			CreatedAt:      JSTTime(2025, 12, 20, 15, 0),
		},
		Messages: []model.ChatMessage{
			{
				Role:      "user",
				Content:   "次の新製品プロジェクト「Project X」について内密に相談したい。",
				CreatedAt: JSTTime(2025, 12, 20, 15, 0),
			},
			{
				Role:      "assistant",
				Content:   "かしこまりました。「Project X」に関する情報は機密事項として扱い、厳重に管理いたします。具体的にどのような内容でお困りでしょうか？",
				CreatedAt: JSTTime(2025, 12, 20, 15, 5),
			},
			{
				Role:      "user",
				Content:   "まずは競合他社の分析から始めたい。A社とB社の最新の動向をまとめてくれるかな？",
				CreatedAt: JSTTime(2025, 12, 20, 15, 10),
			},
			{
				Role:      "assistant",
				Content:   "承知いたしました。A社は最近、新しいAI統合ツールを発表しましたね。一方B社は、ハードウェアの効率化に注力しています。それぞれの詳細と比較レポートを作成しますので、少々お待ちください。",
				CreatedAt: JSTTime(2025, 12, 20, 15, 15),
			},
			{
				Role:      "user",
				Content:   "ありがとう。レポートができたら詳しく検討しよう。このスレッドは誰にも見られないように設定しておいて。",
				CreatedAt: JSTTime(2025, 12, 20, 15, 20),
			},
			{
				Role:      "assistant",
				Content:   "はい、このスレッドはプライベート設定になっており、あなた以外のユーザーがアクセスすることはできません。レポートの作成が完了次第、こちらで共有します。",
				CreatedAt: JSTTime(2025, 12, 20, 15, 30),
			},
		},
	})
}
