package seeds

import "youdoyou-server/model"

func init() {
	Register("codeblock", SeedData{
		Thread: model.ChatThread{
			ID:         "codeblock-thread",
			UserID:     "demo-user-001",
			Title:      "Code Block Example",
			IsPrivate:  false,
			IsArchived: false,
			LastMessage: `
**Googleの「Agent Development Kit (ADK-go)」** という正確な名称の公式製品は現在存在しませんが、おそらく **「Firebase Genkit (Go SDK)」** のことを指している可能性が高いです。

Googleは現在、AIエージェント開発のためのフレームワークとして **Genkit** を強く推進しており、これがGo言語を公式にサポートしています（以前はNode.jsのみでしたが、Goのサポートが追加されました）。

もし「ADK」が別の特定のツール（例：Android Open Accessory Development Kitなど）を指している場合はお知らせください。ここでは、Googleの最新の **AIエージェント開発キットである「Genkit for Go」** について解説します。

---

### Firebase Genkit (Go) とは？

Firebase Genkitは、AI機能をアプリケーションに統合するためのオープンソースのフレームワークです。単なるAPIラッパーではなく、 **「エージェント（Agent）」** や「フロー（Flow）」を定義し、デバッグ、デプロイまでを一貫して行える開発キットです。

* **主な特徴:**
* **Go言語ネイティブ:** Goの慣習に沿った書き方が可能。
* **モデルに依存しない:** Gemini, Claude, Llamaなどをプラグインで切り替え可能。
* **開発者UI:** ローカルでエージェントの動きを可視化・テストできるGUIツールが付属。
* **フロー中心:** 入力から出力までの処理の流れ（RAG、ツール呼び出しなど）を定義しやすい。



---

### コード例：シンプルなAIエージェントの作成

以下は、GoogleのGeminiモデルを使用して、入力されたテーマについてジョークを言うシンプルなエージェント（フロー）の作成例です。

#### 1. 事前準備

まず、必要なパッケージを取得します。

` + "```" + `bash
go get github.com/firebase/genkit/go/...

` + "```" + `

#### 2. main.go の実装

` + "```" + `go
package main

import (
	"context"
	"fmt"
	"log"

	// Genkitのコア機能
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	
	// Google AI (Gemini) プラグイン
	"github.com/firebase/genkit/go/plugins/googleai"
)

func main() {
	ctx := context.Background()

	// 1. Google AIプラグインの初期化 (APIキーは環境変数 GOOGLE_GENAI_API_KEY に設定)
	if err := googleai.Init(ctx, nil); err != nil {
		log.Fatalf("Google AIの初期化に失敗しました: %v", err)
	}

	// 2. 使用するモデルの定義
	model := googleai.Model("gemini-1.5-flash")

	// 3. 「Flow（エージェントの処理フロー）」の定義
	// "jokeFlow" という名前でフローを登録します
	jokeFlow := genkit.DefineFlow("jokeFlow", func(ctx context.Context, subject string) (string, error) {
		
		// AIへのプロンプト作成
		prompt := fmt.Sprintf("%s についての面白いジョークを1つ言ってください。", subject)

		// AIモデルの実行
		resp, err := model.Generate(ctx, ai.NewGenerateRequest(
			ai.WithTextPrompt(prompt),
		))
		if err != nil {
			return "", err
		}

		// 結果のテキストを返す
		return resp.Text(), nil
	})

	// 4. Genkitサーバーの起動 (ローカル開発用UIやAPIサーバーとして機能)
	// これにより、CLIやGUIからこのフローを呼び出せるようになります
	if err := genkit.Init(ctx, nil); err != nil {
		log.Fatalf("Genkitの起動に失敗しました: %v", err)
	}

	// 以下はコード内で直接実行する場合の例
	/*
	result, err := jokeFlow.Run(ctx, "バナナ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("AIの回答:", result)
	*/
}

` + "```" + `

#### 3. 実行とテスト

このコードを実行すると、Genkitの開発者UI（Developer UI）を起動してブラウザでテストできます。

` + "```" + `bash
# Genkit CLIを使ってUIを起動 (事前にCLIのインストールが必要)
genkit start -- go run main.go

` + "```" + `

---

### なぜこれが「エージェント開発」に向いているのか？

単にAPIを叩くだけなら標準ライブラリでも可能ですが、Genkit（ADK的な役割）を使うと以下のことが簡単に実装できます。

1. **ツール使用 (Function Calling):**
エージェントに「天気予報API」や「カレンダー検索」などのGo関数を渡し、AIが必要なときにそのGo関数を自律的に実行する仕組みを簡単に書けます。
2. **履歴の管理:**
チャットボットのような会話履歴の保持をサポートします。
3. **構造化出力:**
AIの回答をただのテキストではなく、Goの ` + "`struct`" + `（JSON）として確実に受け取る設定が簡単です。

### 補足：Vertex AI SDK for Go

もしフレームワーク（Genkit）を使わず、もっと低レイヤーでGoogle CloudのAIをGoで扱いたい場合は、 **Vertex AI Go SDK** を使用します。こちらは「キット」というよりは純粋なクライアントライブラリです。

` + "```" + `bash
go get cloud.google.com/go/vertexai/genai

` + "```" + `

### 次のステップ

Genkitについてより深く知りたい、あるいは「ADK-go」が実はIoTデバイス向けのGoogleツール（Android Accessory）のことだった、などあれば教えてください。

**ツールの使い方（Function Calling）を実装するコード例を見たいですか？**
`,
			LastMessageAt: JSTTime(2025, 12, 20, 14, 10),
			CreatedAt:     JSTTime(2025, 12, 20, 14, 0),
			UpdatedAt:     JSTTime(2025, 12, 20, 14, 10),
		},
		Messages: []model.ChatMessage{
			{
				Role:      "user",
				Content:   "GoogleのAgentDevelopmentKit ADK-goについて詳しく教えて。コード例も見せて。",
				Status:    "completed",
				CreatedAt: JSTTime(2025, 12, 20, 14, 0),
			},
			{
				Role: "assistant",
				Content: `
**Googleの「Agent Development Kit (ADK-go)」** という正確な名称の公式製品は現在存在しませんが、おそらく **「Firebase Genkit (Go SDK)」** のことを指している可能性が高いです。

Googleは現在、AIエージェント開発のためのフレームワークとして **Genkit** を強く推進しており、これがGo言語を公式にサポートしています（以前はNode.jsのみでしたが、Goのサポートが追加されました）。

もし「ADK」が別の特定のツール（例：Android Open Accessory Development Kitなど）を指している場合はお知らせください。ここでは、Googleの最新の **AIエージェント開発キットである「Genkit for Go」** について解説します。

---

### Firebase Genkit (Go) とは？

Firebase Genkitは、AI機能をアプリケーションに統合するためのオープンソースのフレームワークです。単なるAPIラッパーではなく、 **「エージェント（Agent）」** や「フロー（Flow）」を定義し、デバッグ、デプロイまでを一貫して行える開発キットです。

* **主な特徴:**
* **Go言語ネイティブ:** Goの慣習に沿った書き方が可能。
* **モデルに依存しない:** Gemini, Claude, Llamaなどをプラグインで切り替え可能。
* **開発者UI:** ローカルでエージェントの動きを可視化・テストできるGUIツールが付属。
* **フロー中心:** 入力から出力までの処理の流れ（RAG、ツール呼び出しなど）を定義しやすい。



---

### コード例：シンプルなAIエージェントの作成

以下は、GoogleのGeminiモデルを使用して、入力されたテーマについてジョークを言うシンプルなエージェント（フロー）の作成例です。

#### 1. 事前準備

まず、必要なパッケージを取得します。

` + "```" + `bash
go get github.com/firebase/genkit/go/...

` + "```" + `

#### 2. main.go の実装

` + "```" + `go
package main

import (
	"context"
	"fmt"
	"log"

	// Genkitのコア機能
	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	
	// Google AI (Gemini) プラグイン
	"github.com/firebase/genkit/go/plugins/googleai"
)

func main() {
	ctx := context.Background()

	// 1. Google AIプラグインの初期化 (APIキーは環境変数 GOOGLE_GENAI_API_KEY に設定)
	if err := googleai.Init(ctx, nil); err != nil {
		log.Fatalf("Google AIの初期化に失敗しました: %v", err)
	}

	// 2. 使用するモデルの定義
	model := googleai.Model("gemini-1.5-flash")

	// 3. 「Flow（エージェントの処理フロー）」の定義
	// "jokeFlow" という名前でフローを登録します
	jokeFlow := genkit.DefineFlow("jokeFlow", func(ctx context.Context, subject string) (string, error) {
		
		// AIへのプロンプト作成
		prompt := fmt.Sprintf("%s についての面白いジョークを1つ言ってください。", subject)

		// AIモデルの実行
		resp, err := model.Generate(ctx, ai.NewGenerateRequest(
			ai.WithTextPrompt(prompt),
		))
		if err != nil {
			return "", err
		}

		// 結果のテキストを返す
		return resp.Text(), nil
	})

	// 4. Genkitサーバーの起動 (ローカル開発用UIやAPIサーバーとして機能)
	// これにより、CLIやGUIからこのフローを呼び出せるようになります
	if err := genkit.Init(ctx, nil); err != nil {
		log.Fatalf("Genkitの起動に失敗しました: %v", err)
	}

	// 以下はコード内で直接実行する場合の例
	/*
	result, err := jokeFlow.Run(ctx, "バナナ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("AIの回答:", result)
	*/
}

` + "```" + `

#### 3. 実行とテスト

このコードを実行すると、Genkitの開発者UI（Developer UI）を起動してブラウザでテストできます。

` + "```" + `bash
# Genkit CLIを使ってUIを起動 (事前にCLIのインストールが必要)
genkit start -- go run main.go

` + "```" + `

---

### なぜこれが「エージェント開発」に向いているのか？

単にAPIを叩くだけなら標準ライブラリでも可能ですが、Genkit（ADK的な役割）を使うと以下のことが簡単に実装できます。

1. **ツール使用 (Function Calling):**
エージェントに「天気予報API」や「カレンダー検索」などのGo関数を渡し、AIが必要なときにそのGo関数を自律的に実行する仕組みを簡単に書けます。
2. **履歴の管理:**
チャットボットのような会話履歴の保持をサポートします。
3. **構造化出力:**
AIの回答をただのテキストではなく、Goの ` + "`struct`" + `（JSON）として確実に受け取る設定が簡単です。

### 補足：Vertex AI SDK for Go

もしフレームワーク（Genkit）を使わず、もっと低レイヤーでGoogle CloudのAIをGoで扱いたい場合は、 **Vertex AI Go SDK** を使用します。こちらは「キット」というよりは純粋なクライアントライブラリです。

` + "```" + `bash
go get cloud.google.com/go/vertexai/genai

` + "```" + `

### 次のステップ

Genkitについてより深く知りたい、あるいは「ADK-go」が実はIoTデバイス向けのGoogleツール（Android Accessory）のことだった、などあれば教えてください。

**ツールの使い方（Function Calling）を実装するコード例を見たいですか？**
`,
				Status:    "received",
				CreatedAt: JSTTime(2025, 12, 20, 14, 1),
			},
		},
	})
}
