package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"github.com/PuloV/ics-golang"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

type ICalendar struct {
	Name  string
	URL   string
	Color string
}

type CommonEvent struct {
	Summary     string
	Description string
	Location    string
	Color       string
	Start       time.Time
	End         time.Time
}

func getIcalendarFromJSONArray(fileName string) ([]ICalendar, error) {
	data, err := ioutil.ReadFile(fileName)
	if err != nil {
		return nil, err
	}
	var calendars []ICalendar
	err = json.Unmarshal(data, &calendars)

	if err != nil {
		return nil, err
	} else {
		return calendars, nil
	}
}

func getTokenFromJSON(fileName string) (*oauth2.Token, error) {
	fp, err := os.Open(fileName)
	defer fp.Close()
	if err != nil {
		return nil, err
	}
	token := &oauth2.Token{}
	err = json.NewDecoder(fp).Decode(token)
	return token, err
}

func main() {
	b, err := ioutil.ReadFile("./credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}
	startTime := time.Now()
	endTime := startTime.Add(7 * 24 * time.Hour)

	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	token, err := getTokenFromJSON("./token.json")
	if err != nil {
		log.Fatalf("unable to read token data: %v", err)
	}
	httpClient := config.Client(context.Background(), token)
	service, err := calendar.New(httpClient)
	cal, err := service.CalendarList.List().MaxResults(20).Do()

	if err != nil {
		log.Fatalf("Unable to read calendar: %v", err)
	}

	colors, err := service.Colors.Get().Do()

	if err != nil {
		log.Fatalf("cannot read color data from google calendar. : %v", err)
	}

	var commonEvents []CommonEvent

	if len(cal.Items) == 0 {
		log.Fatalf("retrieved data length is equal to 0")
	} else {
		for _, item := range cal.Items {
			id := item.Id
			color := colors.Calendar[item.ColorId].Background
			events, err := service.Events.List(id).TimeMin(startTime.Format("2006-01-02T15:04:05-07:00")).TimeMax(endTime.Format("2006-01-02T15:04:05-07:00")).Do()
			if err != nil {
				log.Fatalf("Unable to get events: %v", err)
			}
			for _, event := range events.Items {
				var startDateStr string
				var endDateStr string
				var format string

				if len(event.Start.Date) != 0 {
					startDateStr = event.Start.Date
					format = "2006-01-02"
				} else {
					startDateStr = event.Start.DateTime
					format = "2006-01-02T15:04:05-07:00"
				}
				if len(event.End.Date) != 0 {
					endDateStr = event.End.Date
					format = "2006-01-02"
				} else {
					endDateStr = event.End.DateTime
					format = "2006-01-02T15:04:05-07:00"
				}
				startDate, _ := time.Parse(format, startDateStr)
				endDate, _ := time.Parse(format, endDateStr)
				commonEvents = append(commonEvents, CommonEvent{Summary: event.Summary, Location: event.Location, Description: event.Description, Color: color, Start: startDate, End: endDate})
			}
		}
	}

	icsCalendars, err := getIcalendarFromJSONArray("./ics.json")
	if err != nil {
		log.Fatalf("Unable to load ics json file: %v", err)
	}

	icsParser := ics.New()
	inputChannel := icsParser.GetInputChan()

	urlIcalMap := make(map[string]ICalendar)

	for _, ical := range icsCalendars {
		inputChannel <- ical.URL
		urlIcalMap[ical.URL] = ical
	}
	icsParser.Wait()

	calendars, err := icsParser.GetCalendars()
	if err != nil {
		log.Fatalf("cannnot parse ics data: %s", err)
	}

	for _, cal := range calendars {
		for date := startTime; date.Before(endTime); date = date.Add(24 * time.Hour) {
			eventList, err := cal.GetEventsByDate(date)
			if err != nil {
				// log.Fatalf("cannnot load event list")
			} else {
				for _, event := range eventList {
					color := urlIcalMap[event.GetCalendar().GetUrl()].Color
					commonEvents = append(commonEvents, CommonEvent{Summary: event.GetSummary(), Location: event.GetLocation(), Description: event.GetDescription(), Color: color, Start: event.GetStart(), End: event.GetEnd()})
				}
			}
		}
	}

	for _, event := range commonEvents {
		fmt.Println(event)
	}
}
