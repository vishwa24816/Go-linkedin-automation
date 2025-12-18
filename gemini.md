Build a Go-based LinkedIn automation tool using the Rod library that showcases advanced browser automation capabilities, human-like behavior simulation, and sophisticated anti-bot detection techniques.

This proof-of-concept evaluates your understanding of browser automation patterns, stealth mechanisms, and your ability to architect clean, maintainable Go code that mimics authentic user behavior.

Core Functional Requirements

Authentication System

Login using credentials from environment variables

Detect and handle login failures gracefully

Identify security checkpoints (2FA, captcha)

Persist session cookies for seamless reuse

Search & Targeting

Search users by job title, company, location, keywords

Parse and collect profile URLs efficiently

Handle pagination across search results

Implement duplicate profile detection

Connection Requests

Navigate to user profiles programmatically

Click Connect button with precise targeting

Send personalized notes within character limits

Track sent requests and enforce daily limits

Messaging System

Detect newly accepted connections

Send follow-up messages automatically

Support templates with dynamic variables

Maintain comprehensive message tracking

Anti-Bot Detection Strategy

Implementing robust anti-detection mechanisms is critical to this assignment. You must implement at least 8 stealth techniques, including all 3 mandatory requirements listed below. These techniques simulate authentic human behavior patterns and mask automation signatures.

Human-like Mouse Movement

Implement Bézier curves with variable speed, natural overshoot, and micro-corrections. Avoid straight-line trajectories that indicate bot behavior.

Randomized Timing Patterns

Add realistic, randomized delays between actions. Vary think time, scroll speed, and interaction intervals to mimic human cognitive processing.

Browser Fingerprint Masking

Modify user agent strings, adjust viewport dimensions, disable automation flags (navigator.webdriver), and randomize browser properties to avoid detection.

Select at least 5 additional techniques from this list to enhance anti-detection capabilities. Combining multiple methods creates more convincing human-like behavior patterns and reduces detection probability.

Random Scrolling Behavior

Implement variable scroll speeds, natural acceleration/deceleration, occasional scroll-back movements, and viewport-aware scrolling patterns.

Realistic Typing Simulation

Vary keystroke intervals, introduce occasional typos with corrections, implement backspace patterns, and simulate human typing rhythm variations.

Mouse Hovering & Movement

Add random hover events over elements, implement natural cursor wandering, and create realistic movement patterns during page interactions.

Activity Scheduling

Operate only during business hours, implement realistic break patterns, vary daily activity windows, and simulate human work schedules.

Rate Limiting & Throttling

Enforce connection request quotas, space out messaging intervals, implement cooldown periods, and track daily/hourly action limits.

Code Quality Standards

Modular Architecture

Organize code into logical packages: authentication, search, messaging, stealth, config. Separate concerns clearly with well-defined interfaces.

Robust Error Handling

Implement comprehensive error detection, graceful degradation, retry mechanisms with exponential backoff, and detailed error logging.

Structured Logging

Use leveled logging (debug, info, warn, error), include contextual information, timestamp all events, and support configurable output formats.

Configuration Management

Support YAML or JSON config files, environment variable overrides, validation of config values, and sensible defaults for all settings.

State Persistence

Track sent requests, accepted connections, and message history using SQLite or JSON storage. Enable resumption after interruptions.

Documentation & Comments

Write clear inline comments explaining complex logic, document public functions, include usage examples, and maintain a comprehensive README.

