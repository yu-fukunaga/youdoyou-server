package seeds

import "youdoyou-server/model"

func init() {
	Register("basic", SeedData{
		Thread: model.ChatThread{
			ID:             "basic-thread",
			UserID:         "demo-user-001",
			FirstMessage:   "こんにちは！",
			UnreadCount:    0,
			LastReadAt:     JSTTime(2025, 12, 20, 14, 11),
			ReplyCount:     6,
			IsPrivate:      false,
			IsArchived:     false,
			SessionMemory:  "",
			MemorizedUntil: JSTTime(1970, 1, 1, 0, 0),
			CreatedAt:      JSTTime(2025, 12, 20, 14, 0),
		},
		Messages: []model.ChatMessage{
			{
				Role:      "user",
				Content:   "こんにちは！",
				CreatedAt: JSTTime(2025, 12, 20, 14, 0),
			},
			{
				Role:      "assistant",
				Content:   "こんにちは！何かお手伝いできることはありますか？",
				CreatedAt: JSTTime(2025, 12, 20, 14, 1),
			},
			{
				Role:      "user",
				Content:   "今日の天気はどうかな？",
				CreatedAt: JSTTime(2025, 12, 20, 14, 5),
			},
			{
				Role:      "assistant",
				Content:   "申し訳ありませんが、私はリアルタイムの天気情報に直接アクセスすることはできません。お住まいの地域のニュースサイトや天気予報アプリをご確認いただくのが確実です。他にご質問はありますか？",
				CreatedAt: JSTTime(2025, 12, 20, 14, 6),
			},
			{
				Role:      "user",
				Content:   "わかった、ありがとう。じゃあ、何か面白い雑学を教えて。",
				CreatedAt: JSTTime(2025, 12, 20, 14, 10),
			},
			{
				Role:      "assistant",
				Content:   "もちろんです！面白い雑学ですね。「シロクマの毛は白ではなく、実は透明」というのはご存知でしたか？中が空洞になっていて光を反射するため白く見えるんですよ。さらに、その下の皮膚は黒色で、太陽の熱を吸収しやすくできているんです。",
				CreatedAt: JSTTime(2025, 12, 20, 14, 11),
			},
		},
	})
}
