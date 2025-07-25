package robots

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// RobotsChecker handles robots.txt parsing and URL validation
type RobotsChecker struct {
	userAgent string
	logger    *slog.Logger
	cache     map[string]*RobotsData
}

// RobotsData represents parsed robots.txt data for a domain
type RobotsData struct {
	rules      []Rule
	crawlDelay time.Duration
	sitemaps   []string
	fetchTime  time.Time
}

// Rule represents a robots.txt rule
type Rule struct {
	UserAgent string
	Directive string // "Allow" or "Disallow"
	Path      string
}

// NewRobotsChecker creates a new robots.txt checker
func NewRobotsChecker(userAgent string, logger *slog.Logger) *RobotsChecker {
	if logger == nil {
		logger = slog.Default()
	}

	return &RobotsChecker{
		userAgent: userAgent,
		logger:    logger,
		cache:     make(map[string]*RobotsData),
	}
}

// IsAllowed checks if a URL is allowed according to robots.txt
func (rc *RobotsChecker) IsAllowed(targetURL string) (bool, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false, fmt.Errorf("invalid URL: %w", err)
	}

	// Validate URL scheme
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return false, fmt.Errorf("invalid URL: missing scheme or host")
	}

	// Get domain key for caching
	domain := parsedURL.Scheme + "://" + parsedURL.Host

	// Check cache first
	robotsData, exists := rc.cache[domain]
	if !exists {
		// Fetch and parse robots.txt
		robotsData, err = rc.fetchRobots(domain)
		if err != nil {
			rc.logger.Warn("Failed to fetch robots.txt, allowing by default",
				"domain", domain, "error", err)
			return true, nil // Allow by default if robots.txt is unavailable
		}
		rc.cache[domain] = robotsData
	}

	// Check if URL is allowed based on rules
	return rc.checkRules(robotsData.rules, parsedURL.Path), nil
}

// GetCrawlDelay returns the crawl delay for a domain
func (rc *RobotsChecker) GetCrawlDelay(targetURL string) (time.Duration, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return 0, fmt.Errorf("invalid URL: %w", err)
	}

	// Validate URL scheme
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return 0, fmt.Errorf("invalid URL: missing scheme or host")
	}

	domain := parsedURL.Scheme + "://" + parsedURL.Host
	robotsData, exists := rc.cache[domain]
	if !exists {
		robotsData, err = rc.fetchRobots(domain)
		if err != nil {
			return 0, nil // No delay if robots.txt is unavailable
		}
		rc.cache[domain] = robotsData
	}

	return robotsData.crawlDelay, nil
}

// fetchRobots fetches and parses robots.txt from a domain
func (rc *RobotsChecker) fetchRobots(domain string) (*RobotsData, error) {
	robotsURL := domain + "/robots.txt"

	rc.logger.Debug("Fetching robots.txt", "url", robotsURL)

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	req, err := http.NewRequest("GET", robotsURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", rc.userAgent)

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch robots.txt: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("robots.txt returned status %d", resp.StatusCode)
	}

	// Parse robots.txt content
	robotsData := &RobotsData{
		rules:     make([]Rule, 0),
		fetchTime: time.Now(),
	}

	scanner := bufio.NewScanner(resp.Body)
	currentUserAgent := ""

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// Split directive and value
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}

		directive := strings.TrimSpace(strings.ToLower(parts[0]))
		value := strings.TrimSpace(parts[1])

		switch directive {
		case "user-agent":
			currentUserAgent = value
		case "disallow", "allow":
			if currentUserAgent != "" && rc.matchesUserAgent(currentUserAgent) {
				robotsData.rules = append(robotsData.rules, Rule{
					UserAgent: currentUserAgent,
					Directive: strings.Title(directive),
					Path:      value,
				})
			}
		case "crawl-delay":
			if currentUserAgent != "" && rc.matchesUserAgent(currentUserAgent) {
				if delay, err := time.ParseDuration(value + "s"); err == nil {
					robotsData.crawlDelay = delay
				}
			}
		case "sitemap":
			robotsData.sitemaps = append(robotsData.sitemaps, value)
		}
	}

	rc.logger.Debug("Parsed robots.txt",
		"domain", domain,
		"rules_count", len(robotsData.rules),
		"crawl_delay", robotsData.crawlDelay)

	return robotsData, scanner.Err()
}

// matchesUserAgent checks if a user-agent pattern matches our user agent
func (rc *RobotsChecker) matchesUserAgent(pattern string) bool {
	pattern = strings.ToLower(pattern)
	userAgent := strings.ToLower(rc.userAgent)

	// Handle empty pattern
	if pattern == "" {
		return false
	}

	// Handle wildcard
	if pattern == "*" {
		return true
	}

	// Check for exact match or prefix match
	return strings.Contains(userAgent, pattern)
}

// checkRules evaluates robots.txt rules for a given path
func (rc *RobotsChecker) checkRules(rules []Rule, urlPath string) bool {
	// Default is allowed
	allowed := true
	bestMatch := ""

	// Find the most specific matching rule (longest path match)
	for _, rule := range rules {
		if rc.pathMatches(rule.Path, urlPath) {
			// Use the longest matching path (most specific)
			if len(rule.Path) > len(bestMatch) {
				bestMatch = rule.Path
				allowed = rule.Directive == "Allow"
			}
		}
	}

	return allowed
}

// pathMatches checks if a robots.txt path pattern matches a URL path
func (rc *RobotsChecker) pathMatches(pattern, urlPath string) bool {
	// Empty pattern matches nothing
	if pattern == "" {
		return false
	}

	// Exact match
	if pattern == urlPath {
		return true
	}

	// Prefix match with wildcard
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(urlPath, prefix)
	}

	// Prefix match - pattern must end with /
	if strings.HasSuffix(pattern, "/") {
		return strings.HasPrefix(urlPath, pattern)
	}

	// For patterns without trailing slash, use prefix matching (robots.txt standard)
	return strings.HasPrefix(urlPath, pattern)
}

// ClearCache clears the robots.txt cache
func (rc *RobotsChecker) ClearCache() {
	rc.cache = make(map[string]*RobotsData)
}

// GetCacheSize returns the number of cached robots.txt entries
func (rc *RobotsChecker) GetCacheSize() int {
	return len(rc.cache)
}
