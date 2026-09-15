package calendar

import (
	"encoding/json"
	"fmt"
	"time"

	"timeful/server/models"
	"timeful/server/services"
	"timeful/server/services/providerconfig"
	"timeful/server/utils"
)

type OutlookCalendar struct {
	models.OAuth2CalendarAuth
}

func (calendar *OutlookCalendar) GetCalendarList() (map[string]models.SubCalendar, error) {
	response := services.CallApi(nil, &calendar.OAuth2CalendarAuth, "GET", fmt.Sprintf("%s/me/calendars?$select=id,name", providerconfig.MicrosoftGraphAPIBaseURL()), nil)
	defer response.Body.Close()

	responseBody := struct {
		Value []struct {
			Id   string `json:"id"`
			Name string `json:"name"`
		} `json:"value"`
		Error map[string]any `json:"error"`
	}{}

	err := json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("error fetching Outlook calendars: %v", responseBody.Error)
	}

	calendars := make(map[string]models.SubCalendar)
	for _, calendar := range responseBody.Value {
		calendars[calendar.Id] = models.SubCalendar{
			Name:    calendar.Name,
			Enabled: utils.TruePtr(),
		}
	}

	return calendars, nil
}

func (calendar *OutlookCalendar) GetCalendarEvents(calendarId string, timeMin time.Time, timeMax time.Time) ([]models.CalendarEvent, error) {
	url := fmt.Sprintf("%s/me/calendars/%s/calendarview?startdatetime=%s&enddatetime=%s&$select=id,subject,start,end,showAs",
		providerconfig.MicrosoftGraphAPIBaseURL(),
		calendarId,
		timeMin.Format(time.RFC3339),
		timeMax.Format(time.RFC3339))
	response := services.CallApi(nil, &calendar.OAuth2CalendarAuth, "GET", url, nil)
	defer response.Body.Close()

	responseBody := struct {
		Value []struct {
			Id      string `json:"id"`
			Subject string `json:"subject"`
			Start   struct {
				DateTime string `json:"dateTime"`
			} `json:"start"`
			End struct {
				DateTime string `json:"dateTime"`
			} `json:"end"`
			ShowAs string `json:"showAs"`
		} `json:"value"`
		Error map[string]any `json:"error"`
	}{}
	err := json.NewDecoder(response.Body).Decode(&responseBody)
	if err != nil {
		return nil, err
	}

	if responseBody.Error != nil {
		return nil, fmt.Errorf("error fetching Outlook events: %v", responseBody.Error)
	}

	calendarEvents := make([]models.CalendarEvent, 0)
	for _, event := range responseBody.Value {
		// Custom time format for Outlook date-time strings
		const outlookTimeFormat = "2006-01-02T15:04:05.0000000"

		startTime, err := time.Parse(outlookTimeFormat, event.Start.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse start time: %w", err)
		}
		endTime, err := time.Parse(outlookTimeFormat, event.End.DateTime)
		if err != nil {
			return nil, fmt.Errorf("failed to parse end time: %w", err)
		}

		calendarEvents = append(calendarEvents, models.CalendarEvent{
			Id:         event.Id,
			CalendarId: calendarId,
			Summary:    event.Subject,
			StartDate:  models.NewDateTimeFromTime(startTime),
			EndDate:    models.NewDateTimeFromTime(endTime),
			Free:       event.ShowAs == "free",
		})
	}

	return calendarEvents, nil
}
