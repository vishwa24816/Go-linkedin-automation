package authentication

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	//"github.com/go-rod/rod/lib/proto" // Not used with JS cookie management
	"linkedin-automation/config" // Import the config package
	"linkedin-automation/stealth" // Import the stealth package
)

// Cookie represents a single browser cookie.
type Cookie struct {
	Name       string `json:"name"`
	Value      string `json:"value"`
	Domain     string `json:"domain"`
	Path       string `json:"path"`
	Expires    int64  `json:"expires"`
	Size       int    `json:"size"`
	HTTPOnly   bool   `json:"httpOnly"`
	Secure     bool   `json:"secure"`
	Session    bool   `json:"session"`
	SameSite   string `json:"sameSite"`
	Priority   string `json:"priority"`
	SameParty  bool   `json:"sameParty"`
	SourceScheme string `json:"sourceScheme"`
	SourcePort int    `json:"sourcePort"`
}


// Authenticator handles LinkedIn authentication and session management.
type Authenticator struct {
	Browser *rod.Browser
	Page    *rod.Page
	Config  *config.Config // Add a reference to the configuration
}

// NewAuthenticator creates a new Authenticator instance.
func NewAuthenticator(cfg *config.Config) *Authenticator {
	return &Authenticator{
		Config: cfg,
	}
}

// LaunchBrowser launches a new browser instance.
func (a *Authenticator) LaunchBrowser() error {
	a.Browser = rod.New().
		// .Timeout(10 * time.Minute) // Set a longer timeout for debugging
		MustConnect()

	// stealth.ApplyStealth is a no-op now, as per-page stealth is used.
	log.Println("Browser launched successfully.")
	return nil
}

// CloseBrowser closes the browser instance.
func (a *Authenticator) CloseBrowser() {
	if a.Browser != nil {
		a.Browser.MustClose()
		log.Println("Browser closed.")
	}
}

// Login performs the login operation on LinkedIn.
func (a *Authenticator) Login() error {
	if a.Browser == nil {
		return fmt.Errorf("browser not launched")
	}

	// Try loading cookies first
	loadErr := a.LoadCookies("linkedin_cookies.json")
	if loadErr == nil {
		log.Println("Loaded existing cookies, checking if session is valid...")
		// Create a page and apply stealth
		a.Page = a.Browser.MustPage()
		if err := stealth.ApplyPageStealth(a.Page); err != nil {
			log.Printf("Warning: Failed to apply stealth to page after cookie load: %v", err)
		}
		a.Page.MustNavigate("https://www.linkedin.com/feed/")
		// A more robust check for successful login
		feedModule, err := a.Page.Element(`main#feed-news-module`)
		if err == nil && feedModule.MustVisible() { // Corrected check with MustVisible
			log.Println("Successfully logged in using persistent cookies.")
			return nil
		}
		log.Println("Session invalid, proceeding with new login.")
	} else if os.IsNotExist(loadErr) {
		log.Println("No existing cookies file found, performing fresh login.")
	} else {
		log.Printf("Failed to load cookies: %v, performing fresh login.", loadErr)
	}

	a.Page = a.Browser.MustPage("https://www.linkedin.com/login")
	if err := stealth.ApplyPageStealth(a.Page); err != nil {
		log.Printf("Warning: Failed to apply stealth to login page: %v", err)
	}

	log.Printf("Navigating to LinkedIn login page: %s", a.Page.MustInfo().URL)

	// Wait for the page to load and the elements to be visible
	a.Page.MustWaitStable().
		MustElement("#username").MustInput(a.Config.LinkedIn.Username) // Use Rod's input
	a.Page.MustElement("#password").MustInput(a.Config.LinkedIn.Password) // Use Rod's input

	// Add a random delay before clicking
	stealth.RandomDelay(500*time.Millisecond, 2*time.Second)

	// Click the sign-in button
	a.Page.MustElement(`[type="submit"]`).MustClick()

	// Wait for navigation and potential redirects
	a.Page.MustWaitNavigation()
	// Apply stealth after navigation completes
	if err := stealth.ApplyPageStealth(a.Page); err != nil {
		log.Printf("Warning: Failed to apply stealth after login navigation: %v", err)
	}

	// Check for successful login or error messages
	currentURL := a.Page.MustInfo().URL
	feedModuleAfterLogin, err := a.Page.Element(`main#feed-news-module`)
	if currentURL == "https://www.linkedin.com/feed/" || currentURL == "https://www.linkedin.com/feed/?trk=nav_join" || (err == nil && feedModuleAfterLogin.MustVisible()) { // Corrected check
		log.Println("Successfully logged in to LinkedIn!")
		// Save cookies for future use
		if err := a.SaveCookies("linkedin_cookies.json"); err != nil {
			log.Printf("Warning: Failed to save cookies: %v", err)
		}
		return nil
	}

	// Handle potential login failures or security checkpoints
	// Generic check for common LinkedIn error messages or security challenges
	secVerification, _, err1 := a.Page.Has(`[aria-label*="security verification"]`) // Corrected
	challengeInput, _, err2 := a.Page.Has(`input[name="challengeId"]`)           // Corrected
	if (err1 == nil && secVerification) || (err2 == nil && challengeInput) { // Corrected checks
		return fmt.Errorf("security verification or challenge required (2FA/Captcha detected)")
	}
	// Check for invalid credentials message
	errUsernameEl, _, err3 := a.Page.Has(`[id*="error-for-username"]`) // Corrected
	errPasswordEl, _, err4 := a.Page.Has(`[id*="error-for-password"]`) // Corrected
	formErrorEl, _, err5 := a.Page.Has(`.form__group--error`)         // Corrected
	alertContentEl, _, err6 := a.Page.Has(`.alert-content`)           // Corrected

	if (err3 == nil && errUsernameEl) || (err4 == nil && errPasswordEl) || (err5 == nil && formErrorEl) || (err6 == nil && alertContentEl) { // Corrected checks
		errMsg := ""
		if errEl, err := a.Page.Element(`[id*="error-for-username"]`); err == nil && errEl.MustVisible() {
			errMsg += errEl.MustText() + " "
		}
		if errEl, err := a.Page.Element(`[id*="error-for-password"]`); err == nil && errEl.MustVisible() {
			errMsg += errEl.MustText() + " "
		}
		if errMsg == "" {
			// Fallback for general error messages
			if generalErrEl, err := a.Page.Element(`.form__group--error`); err == nil && generalErrEl.MustVisible() {
				errMsg += generalErrEl.MustText()
			}
		}
		return fmt.Errorf("login failed: %s", errMsg)
	}

	// Generic error if not redirected to feed or an error is detected
	return fmt.Errorf("login failed, unexpected page or state: %s", currentURL)
}

// SaveCookies saves the browser session cookies to a file using JavaScript.
func (a *Authenticator) SaveCookies(filename string) error {
	if a.Page == nil {
		return fmt.Errorf("no page available to save cookies from")
	}

	// Execute JavaScript to get all cookies for the current domain
	js := `
		function getCookies() {
			const cookies = document.cookie.split('; ').map(c => {
				const [name, value] = c.split('=');
				return { Name: name, Value: value };
			});
			return JSON.stringify(cookies);
		}
		getCookies();
	`
	res, err := a.Page.Evaluate(js).Str()
	if err != nil {
		return fmt.Errorf("failed to get cookies via JS: %w", err)
	}

	// Rod's Evaluate returns a string, so res is already the JSON string.
	// We might need to unmarshal and re-marshal if we want pretty print, but for now, save as is.
	err = os.WriteFile(filename, []byte(res), 0644)
	if err != nil {
		return fmt.Errorf("failed to write cookies to file: %w", err)
	}

	log.Printf("Cookies saved to %s", filename)
	return nil
}

// LoadCookies loads browser session cookies from a file and injects them using JavaScript.
func (a *Authenticator) LoadCookies(filename string) error {
	if a.Browser == nil {
		return fmt.Errorf("browser not launched")
	}
	if a.Page == nil {
		// A page is needed to set cookies via JS. If no page exists, defer.
		// Login flow ensures a page is created, so this should eventually be fine.
		return fmt.Errorf("no page available to load cookies into")
	}

	data, err := os.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			return err // Return os.ErrNotExist so caller can check
		}
		return fmt.Errorf("failed to read cookies file: %w", err)
	}

	var cookies []Cookie // Use our custom Cookie struct
	if err := json.Unmarshal(data, &cookies); err != nil {
		return fmt.Errorf("failed to unmarshal cookies: %w", err)
	}

	for _, cookie := range cookies {
		// Construct the cookie string for document.cookie
		cookieStr := fmt.Sprintf("%s=%s; domain=%s; path=%s; expires=%s;",
			cookie.Name, cookie.Value, cookie.Domain, cookie.Path, time.Unix(cookie.Expires, 0).UTC().Format(time.RFC1123))

		if cookie.Secure {
			cookieStr += " Secure;"
		}
		if cookie.HTTPOnly {
			cookieStr += " HttpOnly;"
		}
		// Note: SameSite, SameParty, etc. might need more complex JS to set or are not directly settable via document.cookie

		js := fmt.Sprintf("document.cookie = `%s`;", cookieStr)
		_, err := a.Page.Evaluate(js).Str()
		if err != nil {
			log.Printf("Warning: Failed to set cookie %s via JS: %v", cookie.Name, err)
		}
	}

	log.Printf("Cookies loaded from %s", filename)
	return nil
}
