package repository

import (
	"context"

	"youdoyou-server/model"

	"github.com/jomei/notionapi"
)

type NotionAPIRepository struct {
	client *notionapi.Client
}

func NewNotionRepository(client *notionapi.Client) *NotionAPIRepository {
	return &NotionAPIRepository{client: client}
}

func (r *NotionAPIRepository) QueryDatabase(ctx context.Context, databaseID string, filter map[string]interface{}) ([]model.NotionPage, error) {
	// Convert databaseID to notionapi.DatabaseID
	dbID := notionapi.DatabaseID(databaseID)

	// Build query
	query := &notionapi.DatabaseQueryRequest{
		Filter: buildNotionFilter(filter),
	}

	response, err := r.client.Database.Query(ctx, dbID, query)
	if err != nil {
		return nil, err
	}

	var result []model.NotionPage
	for _, page := range response.Results {
		// Convert notionapi.Properties to map[string]interface{}
		props := make(map[string]interface{})
		for k, v := range page.Properties {
			props[k] = v
		}

		result = append(result, model.NotionPage{
			ID:         page.ID.String(),
			Title:      extractTitle(page),
			Properties: props,
		})
	}

	return result, nil
}

func (r *NotionAPIRepository) CreatePage(ctx context.Context, databaseID string, properties map[string]interface{}) (string, error) {
	dbID := notionapi.DatabaseID(databaseID)

	createRequest := &notionapi.PageCreateRequest{
		Parent: notionapi.Parent{
			DatabaseID: dbID,
		},
		Properties: buildNotionProperties(properties),
	}

	page, err := r.client.Page.Create(ctx, createRequest)
	if err != nil {
		return "", err
	}

	return page.ID.String(), nil
}

func buildNotionFilter(filter map[string]interface{}) notionapi.Filter {
	// Implementation: convert generic filter to notionapi filter
	// Simplification: returning nil for now as spec had placeholder
	return nil
}

func buildNotionProperties(props map[string]interface{}) notionapi.Properties {
	// Implementation: convert generic properties to notionapi properties
	// Simplification: returning empty map for now as spec had placeholder
	return notionapi.Properties{}
}

func extractTitle(page notionapi.Page) string {
	// Extract title from page properties
	// This depends on the specific property name for "Title" in the database
	// Usually "Name" or "Title"
	return ""
}
