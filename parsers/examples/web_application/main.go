package main

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/fsvxavier/nexs-lib/parsers"
	"github.com/fsvxavier/nexs-lib/parsers/datetime"
)

// Event represents an event with flexible date parsing
type Event struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	Location    string    `json:"location"`
	CreatedAt   time.Time `json:"created_at"`
}

// EventService handles event operations with advanced date parsing
type EventService struct {
	parser *datetime.Parser
	events []Event
	nextID int
}

func NewEventService() *EventService {
	// Create a parser optimized for web application use
	parser := datetime.NewParser(
		parsers.WithLocation(time.UTC),              // Use UTC for storage
		parsers.WithDateOrder(parsers.DateOrderDMY), // European format preference
		parsers.WithStrictMode(false),               // Allow flexible parsing
		parsers.WithCustomFormats(
			"02/01/2006 15:04",    // DD/MM/YYYY HH:MM
			"2006-01-02T15:04",    // ISO without timezone
			"Jan 2, 2006 3:04 PM", // Friendly text format
		),
	)

	return &EventService{
		parser: parser,
		events: []Event{},
		nextID: 1,
	}
}

func main() {
	fmt.Println("=== Web Application Example: Event Management System ===")
	fmt.Println("This example shows practical usage in a real web application\n")

	service := NewEventService()

	// Demo the functionality
	fmt.Println("🌐 Event Management Web Server Demo")
	fmt.Println("📊 Demonstrating real-world date parsing scenarios:\n")

	// Demo various date parsing scenarios
	demoWebApplicationUsage(service)

	fmt.Println("\n🚀 To run as actual web server, uncomment:")
	fmt.Println("   // setupRoutes(service)")
	fmt.Println("   // log.Fatal(http.ListenAndServe(\":8080\", nil))")
	fmt.Println("\n📚 Available endpoints would be:")
	fmt.Println("   GET  /                    - Home page")
	fmt.Println("   GET  /events              - List all events")
	fmt.Println("   POST /events/create       - Create new event")
	fmt.Println("   GET  /events/search       - Search events by date range")
	fmt.Println("   POST /api/parse-date      - Parse date API endpoint")
}

func demoWebApplicationUsage(service *EventService) {
	// Simulate various user inputs that would come from web forms
	fmt.Println("1. Creating Events with Various Date Formats:")

	testEvents := []struct {
		title       string
		startDate   string
		endDate     string
		description string
	}{
		{
			title:       "Conference 2024",
			startDate:   "15/03/2024 09:00",
			endDate:     "17/03/2024 17:00",
			description: "Annual tech conference",
		},
		{
			title:       "Team Meeting",
			startDate:   "March 20, 2024 2:00 PM",
			endDate:     "March 20, 2024 3:30 PM",
			description: "Weekly team sync",
		},
		{
			title:       "Product Launch",
			startDate:   "2024-04-01T10:00:00Z",
			endDate:     "2024-04-01T12:00:00Z",
			description: "New product announcement",
		},
		{
			title:       "Holiday Party",
			startDate:   "25/12/2024",
			endDate:     "25/12/2024 23:59",
			description: "Year-end celebration",
		},
		{
			title:       "Training Session",
			startDate:   "1651161600", // Unix timestamp
			endDate:     "1651168800",
			description: "New employee training",
		},
	}

	for i, eventData := range testEvents {
		event, err := service.createEvent(eventData.title, eventData.description,
			eventData.startDate, eventData.endDate, "Conference Room A")

		if err == nil {
			fmt.Printf("   ✓ Event %d: %s\n", i+1, event.Title)
			fmt.Printf("     Start: %s (parsed from: %s)\n",
				event.StartDate.Format("January 2, 2006 15:04"), eventData.startDate)
			fmt.Printf("     End:   %s (parsed from: %s)\n",
				event.EndDate.Format("January 2, 2006 15:04"), eventData.endDate)
		} else {
			fmt.Printf("   ✗ Event %d failed: %v\n", i+1, err)
		}
	}

	fmt.Println("\n2. Date Format Detection:")
	dateInputs := []string{
		"2024-03-15T10:30:45Z",
		"March 15, 2024 10:30 AM",
		"15/03/2024 10:30",
		"15-Mar-2024",
		"1710498645",
		"today",
		"tomorrow at 2pm",
	}

	for _, input := range dateInputs {
		format, err := service.parser.ParseFormat(context.Background(), input)
		if err == nil {
			fmt.Printf("   %-25s -> Format: %s\n", input, format)
		} else {
			fmt.Printf("   %-25s -> Detection failed: %v\n", input, err)
		}
	}

	fmt.Println("\n3. Search Events by Date Range:")
	// Search for events in March 2024
	startSearch := "01/03/2024"
	endSearch := "31/03/2024"
	results := service.searchEventsByDateRange(startSearch, endSearch)

	fmt.Printf("   Events between %s and %s:\n", startSearch, endSearch)
	for _, event := range results {
		fmt.Printf("     • %s (%s)\n", event.Title, event.StartDate.Format("Jan 2, 2006"))
	}

	fmt.Println("\n4. Error Handling Examples:")
	invalidDates := []string{
		"invalid-date",
		"32/13/2024",
		"February 30, 2024",
		"25:99:99",
	}

	for _, invalid := range invalidDates {
		_, err := service.parser.Parse(context.Background(), invalid)
		if err != nil {
			if parseErr, ok := err.(*parsers.ParseError); ok {
				fmt.Printf("   Input: '%s'\n", invalid)
				fmt.Printf("   Error: %s\n", parseErr.Message)
				if len(parseErr.Suggestions) > 0 {
					fmt.Printf("   Suggestions: %v\n", parseErr.Suggestions[0])
				}
			}
		}
	}

	fmt.Println("\n5. API Response Demonstration:")
	demonstrateAPIResponse(service)
}

func (s *EventService) createEvent(title, description, startDateStr, endDateStr, location string) (*Event, error) {
	ctx := context.Background()

	// Parse start date with flexible parsing
	startDate, err := s.parser.Parse(ctx, startDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid start date '%s': %w", startDateStr, err)
	}

	// Parse end date
	endDate, err := s.parser.Parse(ctx, endDateStr)
	if err != nil {
		return nil, fmt.Errorf("invalid end date '%s': %w", endDateStr, err)
	}

	// Validate date logic
	if endDate.Before(startDate) {
		return nil, fmt.Errorf("end date cannot be before start date")
	}

	event := Event{
		ID:          s.nextID,
		Title:       title,
		Description: description,
		StartDate:   startDate,
		EndDate:     endDate,
		Location:    location,
		CreatedAt:   time.Now(),
	}

	s.events = append(s.events, event)
	s.nextID++

	return &event, nil
}

func (s *EventService) searchEventsByDateRange(startStr, endStr string) []Event {
	ctx := context.Background()

	start, err := s.parser.Parse(ctx, startStr)
	if err != nil {
		return nil
	}

	end, err := s.parser.Parse(ctx, endStr)
	if err != nil {
		return nil
	}

	var results []Event
	for _, event := range s.events {
		if (event.StartDate.After(start) || event.StartDate.Equal(start)) &&
			(event.StartDate.Before(end) || event.StartDate.Equal(end)) {
			results = append(results, event)
		}
	}

	return results
}

func demonstrateAPIResponse(service *EventService) {
	// Simulate date parsing API response
	testInput := "15/03/2024 14:30"

	ctx := context.Background()
	parsedDate, err := service.parser.Parse(ctx, testInput)

	response := struct {
		Input       string     `json:"input"`
		Success     bool       `json:"success"`
		ParsedDate  *time.Time `json:"parsed_date,omitempty"`
		Format      string     `json:"format,omitempty"`
		Error       string     `json:"error,omitempty"`
		Suggestions []string   `json:"suggestions,omitempty"`
	}{
		Input: testInput,
	}

	if err != nil {
		response.Success = false
		response.Error = err.Error()

		if parseErr, ok := err.(*parsers.ParseError); ok {
			response.Suggestions = parseErr.Suggestions
		}
	} else {
		response.Success = true
		response.ParsedDate = &parsedDate

		// Detect the format used
		if format, formatErr := service.parser.ParseFormat(ctx, testInput); formatErr == nil {
			response.Format = format
		}
	}

	// Show JSON response
	jsonBytes, _ := json.MarshalIndent(response, "   ", "  ")
	fmt.Printf("   API Response for '%s':\n   %s\n", testInput, jsonBytes)
}

// Commented out HTTP handlers for demonstration - uncomment to use as actual web server

/*
func setupRoutes(service *EventService) {
	http.HandleFunc("/", homeHandler)
	http.HandleFunc("/events", service.eventsHandler)
	http.HandleFunc("/events/create", service.createEventHandler)
	http.HandleFunc("/events/search", service.searchEventsHandler)
	http.HandleFunc("/api/parse-date", service.parseDateAPIHandler)
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	html := `
<!DOCTYPE html>
<html>
<head>
    <title>Event Management System</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 40px; }
        .container { max-width: 800px; margin: 0 auto; }
        .form-group { margin: 15px 0; }
        label { display: block; margin-bottom: 5px; font-weight: bold; }
        input, textarea { width: 100%; padding: 8px; margin-bottom: 10px; }
        button { background: #007cba; color: white; padding: 10px 20px; border: none; cursor: pointer; }
        .examples { background: #f5f5f5; padding: 15px; margin: 20px 0; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🗓️ Event Management System</h1>
        <p>This demo shows advanced date parsing in a real web application.</p>

        <h2>Create New Event</h2>
        <form action="/events/create" method="POST">
            <div class="form-group">
                <label>Event Title:</label>
                <input type="text" name="title" required>
            </div>

            <div class="form-group">
                <label>Description:</label>
                <textarea name="description" rows="3"></textarea>
            </div>

            <div class="form-group">
                <label>Start Date (flexible format):</label>
                <input type="text" name="start_date" placeholder="e.g., 15/03/2024 10:30, March 15 2024, today, tomorrow" required>
            </div>

            <div class="form-group">
                <label>End Date (flexible format):</label>
                <input type="text" name="end_date" placeholder="e.g., 15/03/2024 12:00, March 15 2024 2pm" required>
            </div>

            <div class="form-group">
                <label>Location:</label>
                <input type="text" name="location">
            </div>

            <button type="submit">Create Event</button>
        </form>

        <div class="examples">
            <h3>📝 Supported Date Formats:</h3>
            <ul>
                <li><strong>ISO Format:</strong> 2024-03-15T10:30:45Z</li>
                <li><strong>European:</strong> 15/03/2024, 15/03/2024 10:30</li>
                <li><strong>American:</strong> 03/15/2024, 03/15/2024 10:30 AM</li>
                <li><strong>Text:</strong> March 15, 2024, Mar 15 2024 2:30 PM</li>
                <li><strong>Relative:</strong> today, tomorrow, yesterday</li>
                <li><strong>Unix Timestamps:</strong> 1710498645, 1710498645.123</li>
            </ul>
        </div>

        <h2>📋 Actions</h2>
        <p><a href="/events">View All Events</a></p>
        <p><a href="/events/search">Search Events</a></p>
        <p><a href="/api/parse-date">Date Parser API</a></p>
    </div>
</body>
</html>
`
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprint(w, html)
}

func (s *EventService) eventsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(s.events)
	}
}

func (s *EventService) createEventHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		err := r.ParseForm()
		if err != nil {
			http.Error(w, "Failed to parse form", http.StatusBadRequest)
			return
		}

		title := r.FormValue("title")
		description := r.FormValue("description")
		startDate := r.FormValue("start_date")
		endDate := r.FormValue("end_date")
		location := r.FormValue("location")

		event, err := s.createEvent(title, description, startDate, endDate, location)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create event: %v", err), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(event)
	}
}

func (s *EventService) searchEventsHandler(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start")
	endDate := r.URL.Query().Get("end")

	if startDate == "" || endDate == "" {
		http.Error(w, "start and end date parameters required", http.StatusBadRequest)
		return
	}

	results := s.searchEventsByDateRange(startDate, endDate)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

func (s *EventService) parseDateAPIHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == "POST" {
		var request struct {
			DateString string `json:"date_string"`
		}

		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		ctx := context.Background()

		// Parse the date
		parsedDate, err := s.parser.Parse(ctx, request.DateString)

		response := struct {
			Input       string    `json:"input"`
			Success     bool      `json:"success"`
			ParsedDate  *time.Time `json:"parsed_date,omitempty"`
			Format      string    `json:"format,omitempty"`
			Error       string    `json:"error,omitempty"`
			Suggestions []string  `json:"suggestions,omitempty"`
		}{
			Input: request.DateString,
		}

		if err != nil {
			response.Success = false
			response.Error = err.Error()

			if parseErr, ok := err.(*parsers.ParseError); ok {
				response.Suggestions = parseErr.Suggestions
			}
		} else {
			response.Success = true
			response.ParsedDate = &parsedDate

			// Detect the format used
			if format, formatErr := s.parser.ParseFormat(ctx, request.DateString); formatErr == nil {
				response.Format = format
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	} else {
		// Show API documentation
		w.Header().Set("Content-Type", "text/html")
		html := `
<h1>Date Parser API</h1>
<p>POST to this endpoint with JSON:</p>
<pre>
{
  "date_string": "15/03/2024 10:30"
}
</pre>
<p>Returns:</p>
<pre>
{
  "input": "15/03/2024 10:30",
  "success": true,
  "parsed_date": "2024-03-15T10:30:00Z",
  "format": "02/01/2006 15:04"
}
</pre>
`
		fmt.Fprint(w, html)
	}
}
*/

func init() {
	fmt.Println("=== Web Application Features ===")
	fmt.Println("✅ Flexible date input forms")
	fmt.Println("✅ Real-time format detection")
	fmt.Println("✅ Comprehensive error handling")
	fmt.Println("✅ Date range search capabilities")
	fmt.Println("✅ RESTful API endpoints")
	fmt.Println("✅ JSON response formatting")
	fmt.Println("✅ User-friendly error messages")
	fmt.Println("✅ Multiple timezone support")
	fmt.Println("==============================\n")
}
