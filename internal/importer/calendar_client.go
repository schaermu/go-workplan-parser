package importer

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"google.golang.org/api/option"
)

type CalendarClient struct {
	client     *calendar.Service
	calendarId string
}

func NewCalendarClient(calendarId string) CalendarClient {
	ctx := context.Background()
	b, err := ioutil.ReadFile("credentials.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	config, err := google.ConfigFromJSON(b, calendar.CalendarEventsScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	httpClient := getClient(config)
	srv, err := calendar.NewService(ctx, option.WithHTTPClient(httpClient))
	if err != nil {
		log.Fatalf("Unable to retrieve Calendar client: %v", err)
	}

	return CalendarClient{
		client:     srv,
		calendarId: calendarId,
	}
}

func (c *CalendarClient) CreateEvent(start time.Time, end time.Time, description string, isAllDayEvent bool) {
	startDate := &calendar.EventDateTime{
		DateTime: start.Format("2006-01-02T15:04:05"),
		TimeZone: "Europe/Zurich",
	}
	endDate := &calendar.EventDateTime{
		DateTime: end.Format("2006-01-02T15:04:05"),
		TimeZone: "Europe/Zurich",
	}

	if isAllDayEvent {
		fullDay, _ := time.ParseDuration("24h")
		startDate = &calendar.EventDateTime{
			Date:     start.Format("2006-01-02"),
			TimeZone: "Europe/Zurich",
		}
		endDate = &calendar.EventDateTime{
			Date:     start.Add(fullDay).Format("2006-01-02"),
			TimeZone: "Europe/Zurich",
		}
	}

	event := &calendar.Event{
		Summary:      description,
		Start:        startDate,
		End:          endDate,
		Transparency: "transparent",
	}

	event, err := c.client.Events.Insert(c.calendarId, event).Do()
	if err != nil {
		log.Fatalf("    Unable to create event %v\n", err)
	}
	log.Printf("    Event created: %s\n", event.Id)
}

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	// The file token.json stores the user's access and refresh tokens, and is
	// created automatically when the authorization flow completes for the first
	// time.
	tokFile := "token.json"
	tok, err := tokenFromFile(tokFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		log.Fatalf("Unable to read authorization code: %v", err)
	}

	tok, err := config.Exchange(context.TODO(), authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web: %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}
