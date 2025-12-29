package tool

import (
	"youdoyou-server/repository"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type ToolFactory struct {
	g            *genkit.Genkit
	chatRepo     repository.ChatRepository
	calendarRepo repository.CalendarRepository
	notionRepo   repository.NotionRepository
}

func NewToolFactory(
	g *genkit.Genkit,
	chatRepo repository.ChatRepository,
	calendarRepo repository.CalendarRepository,
	notionRepo repository.NotionRepository,
) *ToolFactory {
	return &ToolFactory{
		g:            g,
		chatRepo:     chatRepo,
		calendarRepo: calendarRepo,
		notionRepo:   notionRepo,
	}
}

// 複数の Tool を一度に返す
func (f *ToolFactory) CreateAllTools() []ai.Tool {
	return []ai.Tool{
		// CreateCalendarTool(f.g, f.calendarRepo),
		CreateNotionTool(f.g, f.notionRepo),
		CreateNotionWriteTool(f.g, f.notionRepo),
	}
}

// 特定の Tool だけ返す
// Not used in current service but keeping it compliant
func (f *ToolFactory) CreateToolsByDependencies(deps []string) []ai.Tool {
	var tools []ai.Tool

	for _, dep := range deps {
		switch dep {
		case "calendar":
			tools = append(tools, CreateCalendarTool(f.g, f.calendarRepo))
		case "notion":
			tools = append(tools,
				CreateNotionTool(f.g, f.notionRepo),
				CreateNotionWriteTool(f.g, f.notionRepo),
			)
		}
	}

	return tools
}
