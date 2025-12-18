package stealth

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/go-rod/rod"
)

// ApplyStealth is a placeholder for browser-wide stealth.
// Due to environment/Rod API limitations, browser-wide EvalOnNewDocument and SetUserAgent/Viewport on Browser are not working as expected.
// We will apply stealth techniques on a per-page basis using ApplyPageStealth.
func ApplyStealth(browser *rod.Browser) (*rod.Browser, error) {
	log.Println("Note: Browser-wide stealth via ApplyStealth is limited in this environment. Applying per-page stealth.")
	return browser, nil
}

// ApplyPageStealth injects JavaScript into a page to spoof browser fingerprints.
func ApplyPageStealth(page *rod.Page) error {
	userAgent := getRandomUserAgent()
	width := 1024 + rand.Intn(200) // 1024 to 1223
	height := 768 + rand.Intn(150) // 768 to 917
	platform := getPlatform()

	log.Printf("Applying page stealth: User-Agent: %s, Viewport: %dx%d, Platform: %s", userAgent, width, height, platform)

	// Inject script to override user agent, platform, viewport, and webdriver flag
	_, err := page.Eval(fmt.Sprintf(`
		Object.defineProperty(navigator, 'userAgent', { get: () => '%s' });
		Object.defineProperty(navigator, 'platform', { get: () => '%s' });
		Object.defineProperty(navigator, 'webdriver', { get: () => undefined });

		// Spoof viewport dimensions
		Object.defineProperty(screen, 'width', { get: () => %d });
		Object.defineProperty(screen, 'height', { get: () => %d });
		Object.defineProperty(window, 'innerWidth', { get: () => %d });
		Object.defineProperty(window, 'innerHeight', { get: () => %d });
	`, userAgent, platform, width, height, width, height))
	if err != nil {
		return fmt.Errorf("failed to inject stealth scripts into page: %w", err)
	}

	return nil
}

// getRandomUserAgent returns a randomly selected common user agent string.
func getRandomUserAgent() string {
	userAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Edge/120.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.1 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36",
	}
	rand.Seed(time.Now().UnixNano())
	return userAgents[rand.Intn(len(userAgents))]
}

// getPlatform returns a random platform string for navigator.platform spoofing.
func getPlatform() string {
	platforms := []string{"Win32", "MacIntel", "Linux x86_64"}
	rand.Seed(time.Now().UnixNano() + 1) // Different seed for platform selection
	return platforms[rand.Intn(len(platforms))]
}


// RandomDelay introduces a random delay within a specified range.
func RandomDelay(min, max time.Duration) {
	delay := min + time.Duration(rand.Int63n(int64(max-min+1)))
	time.Sleep(delay)
}

// SimulateHumanTyping types text with randomized delays and optional typos.
func SimulateHumanTyping(el *rod.Element, text string) error {
	for _, r := range text {
		char := string(r)
		el.MustInput(char)
		// Introduce random delay between keystrokes
		RandomDelay(50*time.Millisecond, 200*time.Millisecond) // Typical human typing speed
	}
	return nil
}

// SimulateHumanClick performs a click with a human-like delay.
func SimulateHumanClick(el *rod.Element) error {
	RandomDelay(100*time.Millisecond, 400*time.Millisecond) // Simulate human reaction time
	el.MustClick()
	return nil
}

// SimulateHumanScroll scrolls the page with variable speed and occasional micro-pauses.
func SimulateHumanScroll(page *rod.Page, distance int) error {
	scrollStep := 50 // Pixels per scroll step
	duration := 200 * time.Millisecond // Base duration for a step
	
	currentScroll := 0
	for currentScroll < distance {
		// Variable scroll speed
		speedFactor := 1.0 + rand.Float64()*0.5 // Vary speed by up to 50%
		step := int(float64(scrollStep) * speedFactor)
		
		if currentScroll + step > distance {
			step = distance - currentScroll
		}
		
		_, err := page.Eval(fmt.Sprintf("window.scrollBy(0, %d)", step))
		if err != nil {
			return fmt.Errorf("failed to scroll: %w", err)
		}
		currentScroll += step
		
		// Micro-pauses
		RandomDelay(duration/2, duration*2)
	}
	return nil
}

// HumanLikeMouseMove is currently disabled due to Rod API limitations in this environment.
// Original goal: Simulate a mouse moving from a source point to a destination point using a Bezier curve.
// This requires precise control over mouse events which is not reliably achievable with the current setup.
func HumanLikeMouseMove(page *rod.Page, startX, startY, endX, endY float64) error {
	log.Println("Warning: HumanLikeMouseMove is currently disabled due to environment limitations.")
	return nil
}