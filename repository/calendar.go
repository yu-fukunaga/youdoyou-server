package repository

import (
	"context"
	"time"

	"youdoyou-server/model"

	"google.golang.org/api/calendar/v3"
)

type GoogleCalendarRepository struct {
	service *calendar.Service
}

func NewGoogleCalendarRepository(service *calendar.Service) CalendarRepository {
	return &GoogleCalendarRepository{service: service}
}

func (r *GoogleCalendarRepository) GetEvents(ctx context.Context, timeRange string, timezone string) ([]model.CalendarEvent, error) {
	// timeRange: "this week", "next 7 days", "today" etc
	startTime, endTime := parseTimeRange(timeRange, timezone)

	events, err := r.service.Events.List("primary").
		TimeMin(startTime.Format(time.RFC3339)).
		TimeMax(endTime.Format(time.RFC3339)).
		Context(ctx).
		Do()

	if err != nil {
		return nil, err
	}

	var result []model.CalendarEvent
	for _, item := range events.Items {
		// Parse time with error handling fallback
		start, _ := time.Parse(time.RFC3339, item.Start.DateTime)
		end, _ := time.Parse(time.RFC3339, item.End.DateTime)

		// If DateTime is empty, it might be an all-day event (Date field)
		if item.Start.DateTime == "" && item.Start.Date != "" {
			start, _ = time.Parse("2006-01-02", item.Start.Date)
			end, _ = time.Parse("2006-01-02", item.End.Date)
		}

		result = append(result, model.CalendarEvent{
			ID:        item.Id,
			Summary:   item.Summary,
			StartTime: start,
			EndTime:   end,
			Location:  item.Location,
		})
	}

	return result, nil
}

func parseTimeRange(timeRange string, timezone string) (time.Time, time.Time) {
	// Implementation of time range parsing
	// e.g., "this week", "next 7 days" → start, end
	loc, _ := time.LoadLocation(timezone)
	if loc == nil {
		loc = time.UTC
	}
	now := time.Now().In(loc)

	switch timeRange {
	case "today":
		start := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
		end := start.AddDate(0, 0, 1)
		return start, end
	case "this week":
		start := now.AddDate(0, 0, -int(now.Weekday()))
		end := start.AddDate(0, 0, 7)
		return start, end
	case "next week":
		start := now.AddDate(0, 0, -int(now.Weekday())+7)
		end := start.AddDate(0, 0, 7)
		return start, end
	case "next 7 days":
		start := now
		end := start.AddDate(0, 0, 7)
		return start, end
	default:
		return now, now.AddDate(0, 0, 7)
	}
}
