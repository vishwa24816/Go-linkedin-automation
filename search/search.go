package search

import (
	"fmt"
	"log"
	"net/url"
	"time"

	"github.com/go-rod/rod"
	//"github.com/go-rod/rod/lib/proto" // Removed: not used directly now
	"linkedin-automation/stealth" // Import stealth for human-like interactions
)

// Searcher handles searching for users on LinkedIn.
type Searcher struct {
	Browser *rod.Browser
	Page    *rod.Page
	VisitedProfileURLs map[string]bool // To detect duplicate profiles
}

// NewSearcher creates a new Searcher instance.
func NewSearcher(browser *rod.Browser) *Searcher {
	return &Searcher{
		Browser: browser,
		VisitedProfileURLs: make(map[string]bool),
	}
}

// SearchUserCriteria defines the search parameters.
type SearchUserCriteria struct {
	JobTitle string
	Company  string
	Location string
	Keywords []string
	PageLimit int // Max number of pages to scrape
}

// SearchUsers performs a search on LinkedIn based on the provided criteria.
func (s *Searcher) SearchUsers(criteria SearchUserCriteria) ([]string, error) {
	if s.Browser == nil {
		return nil, fmt.Errorf("browser not launched")
	}

	// Create a new page for searching
	s.Page = s.Browser.MustPage("").MustWindowMaximize()
	if err := stealth.ApplyPageStealth(s.Page); err != nil {
		log.Printf("Warning: Failed to apply stealth to search page: %v", err)
	}


	log.Println("Navigating to LinkedIn search page.")
	// Direct navigation to a search URL can be more efficient if the parameters are known.
	// For now, let's go to the main feed and then to search.
	s.Page.MustNavigate("https://www.linkedin.com/feed/")
	s.Page.MustWaitStable()
	if err := stealth.ApplyPageStealth(s.Page); err != nil { // Re-apply after navigation
		log.Printf("Warning: Failed to apply stealth after feed navigation: %v", err)
	}
	stealth.RandomDelay(1*time.Second, 3*time.Second) // Simulate reading time

	// Navigate to the People search page
	// There isn't always a direct "People" search link, often it's part of a global search.
	// Let's assume we'll use the main search bar and then filter for "People".

	searchURL := s.buildSearchURL(criteria)
	log.Printf("Navigating to generated search URL: %s", searchURL)
	s.Page.MustNavigate(searchURL)
	s.Page.MustWaitStable()
	if err := stealth.ApplyPageStealth(s.Page); err != nil { // Re-apply after navigation
		log.Printf("Warning: Failed to apply stealth after search navigation: %v", err)
	}
	stealth.RandomDelay(2*time.Second, 5*time.Second) // Simulate page load and user thinking

	var profileURLs []string
	pageCount := 0

	for pageCount < criteria.PageLimit {
		log.Printf("Scraping page %d of search results.", pageCount+1)
		// Scroll to load all results on the current page
		// LinkedIn loads results dynamically, so scrolling is often necessary.
		lastHeight := s.Page.MustEval("document.body.scrollHeight").Int()
		for {
			s.Page.Mouse.Scroll(0.0, float64(int(float64(lastHeight)*0.8)), 100) // Changed to float64 for coords and int for speed
			stealth.RandomDelay(500*time.Millisecond, 1*time.Second)
			newHeight := s.Page.MustEval("document.body.scrollHeight").Int()
			if newHeight == lastHeight {
				break // Scrolled to bottom
			}
			lastHeight = newHeight
		}
		stealth.RandomDelay(1*time.Second, 2*time.Second) // Simulate user reviewing results

		// Extract profile URLs
		// This selector might need to be refined based on LinkedIn's dynamic HTML.
		elements := s.Page.MustElements(".reusable-search__result-container a.app-aware-link")
		for _, el := range elements {
			hrefJSON, err := el.Property("href")
			if err != nil {
				log.Printf("Could not get href property for element: %v", err)
				continue
			}
			href := hrefJSON.Str()
			if href == "" {
				log.Printf("Href property is empty for element.")
				continue
			}
			parsedURL, err := url.Parse(href)
			if err != nil {
				log.Printf("Error parsing URL %s: %v", href, err)
				continue
			}
			// Clean up URL to get base profile link
			profileLink := fmt.Sprintf("https://www.linkedin.com%s", parsedURL.Path)

			// Basic duplicate detection
			if !s.VisitedProfileURLs[profileLink] && s.isProfileURL(profileLink) {
				profileURLs = append(profileURLs, profileLink)
				s.VisitedProfileURLs[profileLink] = true
				log.Printf("Found profile: %s", profileLink)
			}
		}

		// Find and click the next page button
		nextButton := s.Page.MustElements(`button[aria-label="Next"]`)
		if len(nextButton) == 0 || !nextButton[0].MustProperty("disabled").Bool() {
			log.Println("No next page button or button is disabled. End of search results.")
			break
		}

		stealth.RandomDelay(1*time.Second, 3*time.Second) // Simulate human hesitation before clicking next
		nextButton[0].MustClick()
		s.Page.MustWaitNavigation()
		if err := stealth.ApplyPageStealth(s.Page); err != nil { // Re-apply after navigation
			log.Printf("Warning: Failed to apply stealth after next page navigation: %v", err)
		}
		s.Page.MustWaitStable()
		pageCount++
	}

	return profileURLs, nil
}

// buildSearchURL constructs a LinkedIn search URL based on criteria.
// This is a simplified example; LinkedIn's search URL parameters can be complex.
func (s *Searcher) buildSearchURL(criteria SearchUserCriteria) string {
	baseURL := "https://www.linkedin.com/search/results/people/?"
	params := url.Values{}

	if criteria.JobTitle != "" {
		params.Add("keywords", criteria.JobTitle) // LinkedIn often uses 'keywords' for job titles too
	}
	if criteria.Company != "" {
		params.Add("currentCompany", criteria.Company) // This parameter might not be directly usable in the URL.
	}
	if criteria.Location != "" {
		params.Add("location", criteria.Location) // Needs to be a valid LinkedIn location
	}
	if len(criteria.Keywords) > 0 {
		// Append keywords to existing 'keywords' or add new ones
		currentKeywords := params.Get("keywords")
		for _, kw := range criteria.Keywords {
			if currentKeywords != "" {
				currentKeywords += " " + kw
			} else {
				currentKeywords = kw
			}
		}
		params.Set("keywords", currentKeywords)
	}

	return baseURL + params.Encode()
}

// isProfileURL checks if the given URL is likely a LinkedIn profile URL.
func (s *Searcher) isProfileURL(u string) bool {
	// A simple check, might need to be more robust.
	return len(u) > 0 && (u == "https://www.linkedin.com/in/" || u == "https://www.linkedin.com/pub/")
}
