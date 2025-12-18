# Go-based LinkedIn Automation Tool

This project is a proof-of-concept for a Go-based LinkedIn automation tool using the `Rod` library, showcasing advanced browser automation capabilities, human-like behavior simulation, and sophisticated anti-bot detection techniques.

## Features

*   **Authentication System**:
    *   Login using credentials from environment variables or `config.yaml`.
    *   Graceful handling of login failures and security checkpoints.
    *   Persistence and reuse of session cookies.
*   **Search & Targeting**:
    *   Search users by job title, company, location, and keywords.
    *   Efficient parsing and collection of profile URLs.
    *   Handling pagination across search results.
    *   Basic duplicate profile detection.
*   **Connection Requests**:
    *   Navigation to user profiles.
    *   Targeted clicking of the "Connect" button.
    *   Sending personalized notes within character limits.
    *   Tracking sent requests and enforcing daily limits using SQLite.
*   **Messaging System**:
    *   Sending follow-up messages automatically to "accepted" connections.
    *   Support for templates with dynamic variables.
    *   Comprehensive message tracking using SQLite.
*   **Anti-Bot Detection Strategy**:
    *   **Human-like Mouse Movement**: Basic simulation (more advanced implementation is a TODO).
    *   **Randomized Timing Patterns**: Random delays between actions.
    *   **Browser Fingerprint Masking**: Applied on a *per-page basis* by injecting JavaScript to modify user agent strings, adjust viewport dimensions, and disable automation flags (`navigator.webdriver`).
    *   **Random Scrolling Behavior**: Variable scroll speeds and micro-pauses.
    *   **Realistic Typing Simulation**: Varying keystroke intervals.
*   **Code Quality Standards**:
    *   Modular Architecture (organized into packages: `authentication`, `search`, `messaging`, `stealth`, `config`, `storage`).
    *   Robust Error Handling.
    *   Structured Logging (using standard `log` package).
    *   State Persistence using SQLite.

## Project Structure

```
.
├── main.go
├── config.yaml
├── go.mod
├── go.sum
├── linkedin_automation.db (generated after first run)
├── authentication/
│   └── authentication.go
├── config/
│   └── config.go
├── connection/
│   └── connection.go
├── messaging/
│   └── messaging.go
├── search/
│   └── search.go
├── stealth/
│   └── stealth.go
└── storage/
    └── storage.go
```

## Setup and Running

### Prerequisites

*   Go (version 1.18 or higher)
*   Chrome or Chromium browser installed on your system (Rod uses it for automation)

### Installation

1.  **Clone the repository:**
    ```bash
    git clone [repository_url]
    cd [repository_name]
    ```
    (Note: This step is conceptual as you are operating within a managed environment.)

2.  **Download Go modules:**
    ```bash
    go mod tidy
    ```

### Configuration

The application can be configured using a `config.yaml` file or environment variables.

#### `config.yaml` Example:

Create a `config.yaml` file in the project root:

```yaml
linkedin:
  username: "your_linkedin_username"
  password: "your_linkedin_password"
```

#### Environment Variables:

Alternatively, you can set environment variables with the prefix `LINKEDIN_AUTOMATION_`.

Example:
```bash
$env:LINKEDIN_AUTOMATION_LINKEDIN_USERNAME="your_linkedin_username" # For PowerShell
$env:LINKEDIN_AUTOMATION_LINKEDIN_PASSWORD="your_linkedin_password" # For PowerShell
```
or
```bash
export LINKEDIN_AUTOMATION_LINKEDIN_USERNAME="your_linkedin_username" # For Linux/macOS
export LINKEDIN_AUTOMATION_LINKEDIN_PASSWORD="your_linkedin_password" # For Linux/macOS
```

### Running the Tool

To run the tool, execute:

```bash
go run main.go
```

The tool will:
1.  Load configuration.
2.  Launch a browser.
3.  Attempt to log in to LinkedIn (using saved cookies if available), applying per-page stealth.
4.  Perform a sample search for "Software Engineer" in "San Francisco Bay Area" with "Go" and "Golang" keywords, applying per-page stealth.
5.  Send connection requests to found profiles (up to a daily limit, and avoiding duplicates), applying per-page stealth.
6.  Simulate accepted connections and send follow-up messages, applying per-page stealth.

## Current Limitations / TODOs

*   **Rod API Challenges for Global Stealth**: Due to unexpected behavior and undefined methods (`browser.SetUserAgent`, `browser.SetViewport`, `browser.EvalOnNewDocument`) on the `rod.Browser` object in this specific environment, browser fingerprinting is currently implemented on a *per-page basis* by injecting JavaScript. A more robust, global solution would be preferred if environment constraints allow.
*   **Human-like Mouse Movement**: This feature is currently disabled due to persistent Rod API limitations in this environment regarding precise mouse control. Ideally, it would involve sophisticated Bézier curve implementations with overshoot and micro-corrections.
*   **LinkedIn UI Changes**: LinkedIn's UI changes frequently, which can break element selectors. The current selectors are illustrative and may need updates.
*   **Advanced Anti-Bot Techniques**: Further enhancements for anti-bot detection could include WebGL fingerprinting, canvas fingerprinting, and more dynamic manipulation of browser properties.
*   **Error Handling**: While basic error handling is in place, more graceful degradation and retry mechanisms could be added for specific scenarios (e.g., network issues, temporary LinkedIn errors).
*   **Dynamic Data Extraction**: For sending personalized messages, extracting the recipient's name from their profile page is a crucial step that is currently simulated.
*   **"New Connection" Detection**: The `DetectNewConnections` function is a placeholder; real-world implementation requires robust scraping of LinkedIn notifications or the "My Network" page, which is highly fragile.

## License

This project is licensed under the MIT License.