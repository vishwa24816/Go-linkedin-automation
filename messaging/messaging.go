package messaging

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"linkedin-automation/stealth" // Import stealth for human-like interactions
	"linkedin-automation/storage" // Import storage for persistence
)

// Messenger handles sending follow-up messages on LinkedIn.
type Messenger struct {
	Browser *rod.Browser
	Page    *rod.Page
	Storage *storage.Storage // Reference to storage for persistence
}

// NewMessenger creates a new Messenger instance.
func NewMessenger(browser *rod.Browser, store *storage.Storage) *Messenger {
	return &Messenger{
		Browser: browser,
		Storage: store,
	}
}

// SendFollowUpMessage sends a personalized message to a connection.
// For simplicity, we assume we have the profile URL of an accepted connection.
func (m *Messenger) SendFollowUpMessage(profileURL, template string, variables map[string]string) error {
	if m.Browser == nil {
		return fmt.Errorf("browser not launched")
	}

	// Check if message already sent
	existingMessage, err := m.Storage.GetMessageRecord(profileURL)
	if err != nil {
		return fmt.Errorf("failed to check existing message: %w", err)
	}
	if existingMessage != nil {
		log.Printf("Follow-up message already sent to %s at %v", profileURL, existingMessage.SentAt)
		return nil // Or return a specific error
	}

	// Substitute variables into the template
	message := applyTemplate(template, variables)

	// Navigate to the connection's profile page
	m.Page = m.Browser.MustPage(profileURL).MustWaitLoad()
	if err := stealth.ApplyPageStealth(m.Page); err != nil {
		log.Printf("Warning: Failed to apply stealth to message page: %v", err)
	}
	stealth.RandomDelay(2*time.Second, 5*time.Second) // Simulate reading profile

	log.Printf("Navigated to connection's profile: %s", profileURL)

	// Click the "Message" button
	messageButton, err := m.Page.Element(`a[data-control-name="overlay.profile_profile_top_card_primary_action_message_button"]`)
	if err != nil {
		messageButton, err = m.Page.Element(`a.pv-top-card-v2__message-button`)
		if err != nil {
			return fmt.Errorf("message button not found for %s: %w", profileURL, err)
		}
	}

	stealth.SimulateHumanClick(messageButton)
	stealth.RandomDelay(1*time.Second, 2*time.Second) // Wait for message modal/panel to appear

	// Find the message input field (often a contenteditable div or textarea)
	messageInput, err := m.Page.Element(`div[contenteditable="true"].msg-form__contenteditable`)
	if err != nil {
		messageInput, err = m.Page.Element(`textarea.msg-form__textarea`)
		if err != nil {
			return fmt.Errorf("message input field not found for %s: %w", profileURL, err)
		}
	}

	// Type the message
	stealth.SimulateHumanTyping(messageInput, message)
	stealth.RandomDelay(1*time.Second, 3*time.Second)

	// Click the "Send" button
	sendButton, err := m.Page.Element(`button.msg-form__send-button`)
	if err != nil {
		return fmt.Errorf("send button not found: %w", err)
	}

	stealth.SimulateHumanClick(sendButton)
	stealth.RandomDelay(1*time.Second, 3*time.Second)

	// Save message record to storage
	msgRecord := &storage.MessageRecord{
		ProfileURL:   profileURL,
		Message:      message,
		SentAt:       time.Now(),
		TemplateUsed: template, // Or a template ID
	}
	if err := m.Storage.SaveMessageRecord(msgRecord); err != nil {
		return fmt.Errorf("failed to save message record to database: %w", err)
	}

	log.Printf("Follow-up message sent to %s: %s", profileURL, message)
	return nil
}

// applyTemplate substitutes variables in a message template.
func applyTemplate(template string, variables map[string]string) string {
	result := template
	for key, value := range variables {
		result = strings.ReplaceAll(result, fmt.Sprintf("{{%s}}", key), value)
	}
	return result
}

// DetectNewConnections uses storage to find profiles with accepted requests that haven't received a message.
func (m *Messenger) DetectNewConnections() ([]string, error) {
	log.Println("Attempting to detect new connections from storage for messaging...")
	profiles, err := m.Storage.GetProfilesWithAcceptedRequestsWithoutMessage()
	if err != nil {
		return nil, fmt.Errorf("failed to get profiles with accepted requests without message: %w", err)
	}
	return profiles, nil
}
