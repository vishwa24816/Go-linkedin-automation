package connection

import (
	"fmt"
	"log"
	"time"

	"github.com/go-rod/rod"
	"linkedin-automation/stealth" // Import stealth for human-like interactions
	"linkedin-automation/storage" // Import storage for persistence
)

// ConnectionRequester handles sending connection requests on LinkedIn.
type ConnectionRequester struct {
	Browser *rod.Browser
	Page    *rod.Page
	Storage *storage.Storage // Reference to storage for persistence
	DailyLimit int // Example daily limit
}

// NewConnectionRequester creates a new ConnectionRequester instance.
func NewConnectionRequester(browser *rod.Browser, store *storage.Storage) *ConnectionRequester {
	return &ConnectionRequester{
		Browser: browser,
		Storage: store,
		DailyLimit: 100, // Default daily limit, can be configured
	}
}

// SendConnectionRequest navigates to a profile, clicks connect, and sends a personalized note.
func (cr *ConnectionRequester) SendConnectionRequest(profileURL, note string) error {
	if cr.Browser == nil {
		return fmt.Errorf("browser not launched")
	}

	// Check if already sent
	existingRequest, err := cr.Storage.GetSentRequestByProfileURL(profileURL)
	if err != nil {
		return fmt.Errorf("failed to check existing request: %w", err)
	}
	if existingRequest != nil {
		log.Printf("Connection request already processed for %s (status: %s)", profileURL, existingRequest.Status)
		return nil // Or return a specific error if you want to differentiate
	}

	// Check daily limit
	requestsToday, err := cr.Storage.GetCountOfSentRequestsToday()
	if err != nil {
		return fmt.Errorf("failed to get count of sent requests today: %w", err)
	}
	if requestsToday >= cr.DailyLimit {
		return fmt.Errorf("daily connection request limit (%d) reached. Sent %d today.", cr.DailyLimit, requestsToday)
	}

	cr.Page = cr.Browser.MustPage(profileURL).MustWaitLoad()
	if err := stealth.ApplyPageStealth(cr.Page); err != nil {
		log.Printf("Warning: Failed to apply stealth to connection page: %v", err)
	}
	stealth.RandomDelay(2*time.Second, 5*time.Second) // Simulate reading profile

	log.Printf("Navigated to profile: %s", profileURL)

	connectButton, err := cr.Page.Element(`button[aria-label^="Invite"]`)
	if err != nil {
		connectButton, err = cr.Page.Element(`button[data-control-name="connect"]`)
		if err != nil {
			return fmt.Errorf("connect button not found for %s: %w", profileURL, err)
		}
	}

	stealth.SimulateHumanClick(connectButton)
	stealth.RandomDelay(1*time.Second, 2*time.Second) // Wait for modal to appear

	addNoteButton, err := cr.Page.Element(`button.artdeco-button--secondary.mr1[aria-label="Add a note"]`)
	if err == nil {
		stealth.SimulateHumanClick(addNoteButton)
		stealth.RandomDelay(500*time.Millisecond, 1*time.Second) // Wait for textarea to appear

		noteTextArea := cr.Page.MustElement(`textarea#custom-message`)
		if len(note) > 300 {
			note = note[:300]
			log.Printf("Note truncated to 300 characters for %s", profileURL)
		}
		stealth.SimulateHumanTyping(noteTextArea, note)
		stealth.RandomDelay(1*time.Second, 3*time.Second)

		sendButton := cr.Page.MustElement(`button[aria-label="Send now"]`)
		stealth.SimulateHumanClick(sendButton)
		stealth.RandomDelay(1*time.Second, 3*time.Second)
	} else {
		log.Println("No 'Add a note' option, sending direct connection request.")
		sendButton := cr.Page.MustElement(`button[aria-label="Send now"]`)
		stealth.SimulateHumanClick(sendButton)
		stealth.RandomDelay(1*time.Second, 3*time.Second)
	}

	// Save the sent request to storage
	sentReq := &storage.SentRequest{
		ProfileURL: profileURL,
		Note:       note,
		SentAt:     time.Now(),
		Status:     storage.StatusSent,
	}
	if err := cr.Storage.SaveSentRequest(sentReq); err != nil {
		return fmt.Errorf("failed to save sent request to database: %w", err)
	}

	log.Printf("Connection request sent to %s with note: '%s'", profileURL, note)
	return nil
}
