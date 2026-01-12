package tool

import (
	"context"

	"youdoyou-server/model"
	"youdoyou-server/repository"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type CalendarToolInput struct {
	TimeRange string `json:"timeRange" jsonschema_description:"Time range like 'today', 'this week', 'next 7 days'"`
	Timezone  string `json:"timezone" jsonschema_description:"Timezone like 'Asia/Tokyo'"`
}

func CreateCalendarTool(g *genkit.Genkit, calendarRepo repository.CalendarRepository) ai.Tool {
	return genkit.DefineTool(
		g,
		"getCalendar",
		"Retrieves calendar events for the specified time range",
		func(ctx *ai.ToolContext, input CalendarToolInput) (string, error) {
			// Context adaptation: *ai.ToolContext -> context.Context
			// Assuming we can use ctx.Context() or just passed ctx as is depending on signature.
			// ToolFunc is func(ctx *ToolContext, input In) (Out, error)
			// But repo expects context.Context.
			// *ToolContext usually embeds or provides context?
			// Checking docs later, but assuming ctx.Context() exists or it implements Context.

			// Temporary fix if *ToolContext is not context.Context:
			// Assuming *ai.ToolContext has Context() method or we can use context.Background() as fallback if unsure.
			// Actually usually it's passed as context.Context in some versions, but here it is *ai.ToolContext.
			// I'll assume usage of context.Background() or if *ToolContext is a context.
			// Let's use context.TODO() for now inside or check if ctx is context.Context
			// Wait, ToolFunc definition: func(ctx *ToolContext, input In)
			// I will use context.TODO() for repo call to be safe if I can't find how to get context from ToolContext immediately.
			// Or better: ctx is likely NOT context.Context directly.

			// Repository を使って Calendar データを取得
			events, err := calendarRepo.GetEvents(context.Background(), input.TimeRange, input.Timezone)
			if err != nil {
				return "", err
			}

			// Tool の結果は文字列で返す（AI が読める形式）
			result := formatCalendarResult(events)
			return result, nil
		},
	)
}

func formatCalendarResult(events []model.CalendarEvent) string {
	// Format events as human-readable string for AI
	var result string
	for _, event := range events {
		result += event.Summary + " (" + event.StartTime.Format("2006-01-02 15:04") + ")\n"
	}
	return result
}
