package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	calendar "google.golang.org/api/calendar/v3"
)

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
		log.Fatalf("Unable to read calendar")
	}

	var calIds []string

	if len(cal.Items) == 0 {
		log.Fatalf("retrieved data length is equal to 0")
	} else {
		for _, item := range cal.Items {
			calIds = append(calIds, item.Id)
		}

		for _, id := range calIds {

			events, err := service.Events.List(id).TimeMin(startTime.Format("2006-01-02T15:04:05-07:00")).TimeMax(endTime.Format("2006-01-02T15:04:05-07:00")).Do()
			if err != nil {
				log.Fatalf("Unable to get events %v", err)
			}
			for _, item := range events.Items {
				var startDateStr string
				var endDateStr string
				if len(item.Start.Date) != 0 {
					startDateStr = item.Start.Date
				} else {
					startDateStr = item.Start.DateTime
				}
				if len(item.End.Date) != 0 {
					endDateStr = item.End.Date
				} else {
					endDateStr = item.End.DateTime
				}
				fmt.Println(item.Summary + " " + item.Location + " " + item.Description + " " + startDateStr + " " + endDateStr)
			}
			if len(events.Items) != 0 {
				fmt.Println("")
			}

		}
	}

}
