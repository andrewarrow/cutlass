package browser

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/launcher"
)

// BrowserSession represents a browser automation session
type BrowserSession struct {
	Launcher *launcher.Launcher
	Browser  *rod.Browser
	Page     *rod.Page
}

// NewBrowserSession creates a new browser session with common setup
func NewBrowserSession() (*BrowserSession, error) {
	// Launch browser
	l := launcher.New().Headless(true)
	url, err := l.Launch()
	if err != nil {
		return nil, fmt.Errorf("error launching browser: %v", err)
	}

	browser := rod.New().ControlURL(url)
	if err := browser.Connect(); err != nil {
		l.Cleanup()
		return nil, fmt.Errorf("error connecting to browser: %v", err)
	}

	// Create page with panic recovery
	var page *rod.Page
	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Fprintf(os.Stderr, "Error creating page: %v\n", r)
				return
			}
		}()
		page = browser.MustPage()
	}()

	if page == nil {
		browser.Close()
		l.Cleanup()
		return nil, fmt.Errorf("failed to create page")
	}

	// Set timeout
	page = page.Timeout(60 * time.Second)

	return &BrowserSession{
		Launcher: l,
		Browser:  browser,
		Page:     page,
	}, nil
}

// Close cleans up the browser session
func (bs *BrowserSession) Close() {
	if bs.Page != nil {
		bs.Page.Close()
	}
	if bs.Browser != nil {
		bs.Browser.Close()
	}
	if bs.Launcher != nil {
		bs.Launcher.Cleanup()
	}
}

// NavigateAndWait navigates to a URL and waits for it to load
func (bs *BrowserSession) NavigateAndWait(url string) error {
	return bs.NavigateAndWaitWithTimeout(url, 60*time.Second)
}

// NavigateAndWaitWithTimeout navigates to a URL with a custom timeout
func (bs *BrowserSession) NavigateAndWaitWithTimeout(url string, timeout time.Duration) error {
	// Set page timeout for this navigation
	page := bs.Page.Timeout(timeout)
	
	err := page.Navigate(url)
	if err != nil {
		return fmt.Errorf("error navigating to %s: %v", url, err)
	}

	err = page.WaitLoad()
	if err != nil {
		return fmt.Errorf("error waiting for page load: %v", err)
	}

	// Wait for dynamic content - use shorter timeout for Google pages
	waitTime := 1 * time.Second
	if strings.Contains(url, "google.com") {
		waitTime = 500 * time.Millisecond // Even shorter for Google to avoid bot detection
	}
	page.WaitRequestIdle(waitTime, []string{}, []string{}, nil)

	return nil
}

// EnsureDataDir creates the data directory if it doesn't exist
func EnsureDataDir() error {
	dataDir := "./data"
	return os.MkdirAll(dataDir, 0755)
}