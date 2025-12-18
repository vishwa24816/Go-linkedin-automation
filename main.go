package main

import (
	"log"
	"time"

	"linkedin-automation/authentication"
	"linkedin-automation/config"
	"linkedin-automation/connection"
	"linkedin-automation/messaging"
	"linkedin-automation/search"
	"linkedin-automation/stealth"
	"linkedin-automation/storage" // Import the storage package
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	log.Printf("Configuration loaded successfully. LinkedIn Username: %s", cfg.LinkedIn.Username)

	// Initialize Storage
	dbPath := "linkedin_automation.db" // Define your database path
	store, err := storage.NewStorage(dbPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer store.Close() // Ensure database connection is closed

	auth := authentication.NewAuthenticator(cfg)
	if err := auth.LaunchBrowser(); err != nil {
		log.Fatalf("Failed to launch browser: %v", err)
	}
	defer auth.CloseBrowser() // Ensure browser is closed when main exits

	if err := auth.Login(); err != nil {
		log.Fatalf("Failed to login to LinkedIn: %v", err)
	}

	log.Println("Successfully authenticated and logged in to LinkedIn.")

	// Initialize Searcher
	searcher := search.NewSearcher(auth.Browser) // Pass the authenticated browser instance

	// Define sample search criteria
	searchCriteria := search.SearchUserCriteria{
		JobTitle:  "Software Engineer",
		Location:  "San Francisco Bay Area",
		Keywords:  []string{"Go", "Golang"},
		PageLimit: 1, // Search across 1 page for demonstration
	}

	log.Printf("Starting user search with criteria: %+v", searchCriteria)
	profileURLs, err := searcher.SearchUsers(searchCriteria)
	if err != nil {
		log.Fatalf("Error during user search: %v", err)
	}

	log.Printf("Found %d unique profile URLs:", len(profileURLs))
	for _, url := range profileURLs {
		log.Println(url)
	}

	// Initialize ConnectionRequester with storage
	connRequester := connection.NewConnectionRequester(auth.Browser, store)

	// Send connection requests
	log.Println("Sending connection requests...")
	for i, profileURL := range profileURLs {
		// Example personalized note
		note := "Hi, I came across your profile and was impressed by your work in Go. I'd love to connect!"
		if err := connRequester.SendConnectionRequest(profileURL, note); err != nil {
			log.Printf("Failed to send connection request to %s: %v", profileURL, err)
		}
		// Add a longer delay between connection requests to avoid rate limits and detection
		if i < len(profileURLs)-1 {
			stealth.RandomDelay(5*time.Second, 15*time.Second) // Human-like delay between requests
		}
	}

	// Initialize Messenger with storage
	messenger := messaging.NewMessenger(auth.Browser, store)

	// Simulate accepted connections for demonstration purposes
	// In a real scenario, you would use messenger.DetectNewConnections()
	// and then filter for profiles where a connection request was sent by this tool.
	simulatedAcceptedConnections := []string{}
	// Use the first found profile as a simulated accepted connection for demonstration
	// In a real scenario, you would get this from DB after a successful connection.
	if len(profileURLs) > 0 {
		// For demonstration, let's assume the first profile request was accepted
		// Update status in DB
		err := store.UpdateRequestStatus(profileURLs[0], storage.StatusAccepted)
		if err != nil {
			log.Printf("Failed to update status for %s: %v", profileURLs[0], err)
		}
		simulatedAcceptedConnections = append(simulatedAcceptedConnections, profileURLs[0])
	}


	log.Println("Sending follow-up messages to simulated accepted connections...")
	for _, profileURL := range simulatedAcceptedConnections {
		template := "Hello {{Name}}, thanks for connecting! I'm {{MyName}}, a {{MyTitle}}. I was particularly interested in your work on {{Interest}}. Let's chat more about it sometime."
		variables := map[string]string{
			"Name":     "Connection Name", // This would be dynamically extracted
			"MyName":   "Your Name",
			"MyTitle":  "Your Job Title",
			"Interest": "Go-based automation tools",
		}

		if err := messenger.SendFollowUpMessage(profileURL, template, variables); err != nil {
			log.Printf("Failed to send follow-up message to %s: %v", profileURL, err)
		}
		stealth.RandomDelay(10*time.Second, 30*time.Second) // Human-like delay between messages
	}

	log.Println("Automation task completed.")
}
