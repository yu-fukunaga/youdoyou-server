package tool

import (
	"context"

	"youdoyou-server/model"
	"youdoyou-server/repository"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
)

type NotionToolInput struct {
	DatabaseID string                 `json:"databaseId" jsonschema_description:"Notion database ID"`
	Filter     map[string]interface{} `json:"filter" jsonschema_description:"Query filter"`
}

func CreateNotionTool(g *genkit.Genkit, notionRepo repository.NotionRepository) ai.Tool {
	return genkit.DefineTool(
		g,
		"getNotion",
		"Queries Notion database and returns pages matching the filter",
		func(ctx *ai.ToolContext, input NotionToolInput) (string, error) {
			// Repository を使って Notion データを取得
			pages, err := notionRepo.QueryDatabase(context.Background(), input.DatabaseID, input.Filter)
			if err != nil {
				return "", err
			}

			result := formatNotionResult(pages)
			return result, nil
		},
	)
}

type NotionCreateInput struct {
	DatabaseID string                 `json:"databaseId" jsonschema_description:"Notion database ID"`
	Properties map[string]interface{} `json:"properties" jsonschema_description:"Page properties to create"`
}

func CreateNotionWriteTool(g *genkit.Genkit, notionRepo repository.NotionRepository) ai.Tool {
	return genkit.DefineTool(
		g,
		"createNotionPage",
		"Creates a new page in Notion database",
		func(ctx *ai.ToolContext, input NotionCreateInput) (string, error) {
			pageID, err := notionRepo.CreatePage(context.Background(), input.DatabaseID, input.Properties)
			if err != nil {
				return "", err
			}

			return "Page created with ID: " + pageID, nil
		},
	)
}

func formatNotionResult(pages []model.NotionPage) string {
	var result string
	for _, page := range pages {
		result += page.Title + " (ID: " + page.ID + ")\n"
	}
	return result
}
