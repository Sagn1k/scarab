package scraper

import (
	"context"
	"fmt"
	// "math/rand"
	// "strings"
	"time"

	"github.com/Sagn1k/scarab/config"
	"github.com/go-rod/rod"
	// "github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type RenderOptions struct {
	WaitTime    int
	Selectors   []string
}

type BrowserRenderer struct {
	config        *config.Config
	browser       *rod.Browser
	proxyRotator  *ProxyRotator
	headerRotator *HeaderRotator
}

func NewBrowserRenderer(cfg *config.Config, pr *ProxyRotator, hr *HeaderRotator) *BrowserRenderer {
	return &BrowserRenderer{
		config:        cfg,
		proxyRotator:  pr,
		headerRotator: hr,
	}
}

func (r *BrowserRenderer) initBrowser(ctx context.Context) error {
	if r.browser != nil {
		return nil
	}

	l := launcher.New().
		Headless(true).
		Set("--window-size", "1280,720")

	// proxy := r.proxyRotator.GetRandomProxy()
	// if proxy != "" {
	// 	l = l.Proxy(proxy)
	// }

	browserURL := l.MustLaunch()
	
	browser := rod.New().Context(ctx).ControlURL(browserURL)
	err := browser.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to browser: %w", err)
	}
	
	r.browser = browser
	return nil
}

func (r *BrowserRenderer) RenderPage(ctx context.Context, url string, options *RenderOptions) (string, error) {
	if err := r.initBrowser(ctx); err != nil {
		return "", err
	}

	timeoutDuration := time.Duration(r.config.BrowserTimeout) * time.Second
	_, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	page, err := r.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	headers := r.headerRotator.GetHeaders()
	for key, value := range headers {
		_, headerErr := page.SetExtraHeaders([]string{key, value})
		if headerErr != nil {
			return "", fmt.Errorf("failed to set headers: %w", headerErr)
		}
	}

	err = page.Navigate(url)
	if err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	page.WaitNavigation(proto.PageLifecycleEventNameLoad)()

	if options != nil && options.WaitTime > 0 {
		time.Sleep(time.Duration(options.WaitTime) * time.Millisecond)
	}

	if options != nil && len(options.Selectors) > 0 {
		for _, selector := range options.Selectors {
			err := rod.Try(func() {
				page.Timeout(5 * time.Second).MustElement(selector)
			})
			if err != nil {
				fmt.Printf("Warning: selector %s not found: %v\n", selector, err)
			}
		}
	}

	html, err := page.HTML()
	if err != nil {
		return "", fmt.Errorf("failed to get HTML: %w", err)
	}

	return html, nil
}

// func (r *BrowserRenderer) handleCloudflare(page *rod.Page, maxWaitTime int) error {
// 	isCloudflare := false
	
// 	_ = rod.Try(func() {
// 		text, err := page.Element("body").Text()
// 		if err == nil && 
// 		   (strings.Contains(strings.ToLower(text), "cloudflare") || 
// 			strings.Contains(text, "Just a moment...") ||
// 			strings.Contains(text, "Checking your browser") ||
// 			strings.Contains(text, "verify you are human")) {
// 			isCloudflare = true
// 		}
// 	})

// 	if !isCloudflare {
// 		return nil // Not a Cloudflare page, no need to bypass
// 	}
	
// 	fmt.Println("Detected Cloudflare challenge, attempting to solve...")

// 	// Wait for the challenge iframe to load completely (this is important)
// 	time.Sleep(3 * time.Second)
	
// 	// Look for common Cloudflare checkbox selectors
// 	checkboxSelectors := []string{
// 		"input[type=checkbox]",
// 		".recaptcha-checkbox",
// 		"#checkbox",
// 		"#recaptcha-anchor",
// 		"#cf-checkbox",
// 		"[role=checkbox]", 
// 		"div.checkbox",
// 		"span.checkbox",
// 		"iframe[src*='cloudflare']", // Sometimes the checkbox is in an iframe
// 	}
	
// 	// First look for and switch to any security iframe
// 	var iframeHandled bool
// 	_ = rod.Try(func() {
// 		iframes := page.MustElements("iframe")
// 		for _, iframe := range iframes {
// 			src, err := iframe.Attribute("src")
// 			if err == nil && src != nil && 
// 			   (strings.Contains(*src, "cloudflare") || 
// 				strings.Contains(*src, "recaptcha") || 
// 				strings.Contains(*src, "captcha") || 
// 				strings.Contains(*src, "challenge")) {
				
// 				// Switch to this iframe
// 				frameObj := iframe.MustFrame()
				
// 				// Now look for the checkbox inside iframe
// 				for _, selector := range checkboxSelectors {
// 					err := rod.Try(func() {
// 						checkbox := frameObj.MustElement(selector)
						
// 						checkbox.Hover()
// 						time.Sleep(time.Duration(300+rand.Intn(500)) * time.Millisecond)
						
// 						checkbox.Click(proto.InputMouseButtonLeft)
// 						fmt.Println("Clicked checkbox in iframe!")
						
// 						time.Sleep(time.Duration(2000+rand.Intn(1000)) * time.Millisecond)
// 						iframeHandled = true
// 					})
// 					if err == nil {
// 						break
// 					}
// 				}
				
// 				if iframeHandled {
// 					break
// 				}
// 			}
// 		}
// 	})

// 	if !iframeHandled {
// 		for _, selector := range checkboxSelectors {
// 			err := rod.Try(func() {
// 				checkbox := page.MustElement(selector)
				
// 				if checkbox.MustVisible() {
// 					checkbox.Hover()
// 					time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)
					
// 					checkbox.Click(proto.InputMouseButtonLeft)
// 					fmt.Println("Clicked checkbox on main page!")
					
// 					time.Sleep(time.Duration(2000+rand.Intn(1000)) * time.Millisecond)
// 				}
// 			})
			
// 			if err == nil {
// 				break
// 			}
// 		}
// 	}
// 	_ = rod.Try(func() {
// 		buttons := page.MustElements("button")
// 		for _, btn := range buttons {
// 			if txt, err := btn.Text(); err == nil {
// 				txtLower := strings.ToLower(txt)
// 				if strings.Contains(txtLower, "verify") || 
// 				   strings.Contains(txtLower, "continue") || 
// 				   strings.Contains(txtLower, "submit") ||
// 				   strings.Contains(txtLower, "i'm human") {
// 					btn.Click(proto.InputMouseButtonLeft)
// 					fmt.Println("Clicked verification button!")
// 					time.Sleep(2 * time.Second)
// 					break
// 				}
// 			}
// 		}
// 	})
	
// 	// Sometimes we need to press keyboard keys
// 	_ = rod.Try(func() {
// 		// Press Enter key as some forms submit with Enter
// 		page.Keyboard.Press(input.Enter)
// 		time.Sleep(time.Second)
// 	})
	
// 	// Wait for the challenge to complete
// 	// This is crucial - Cloudflare often takes several seconds to verify
// 	fmt.Println("Waiting for Cloudflare verification to complete...")
// 	waitDuration := time.Duration(maxWaitTime) * time.Millisecond
// 	if waitDuration < 10*time.Second {
// 		waitDuration = 10 * time.Second // Minimum 10 seconds wait
// 	}
// 	time.Sleep(waitDuration)

// 	// Check if we're still on the Cloudflare page
// 	stillOnCloudflare := false
// 	_ = rod.Try(func() {
// 		text, err := page.Element("body").Text()
// 		if err == nil && 
// 		   (strings.Contains(strings.ToLower(text), "cloudflare") && 
// 			strings.Contains(text, "Just a moment...")) {
// 			stillOnCloudflare = true
// 		}
// 	})

// 	if stillOnCloudflare {
// 		return fmt.Errorf("failed to bypass Cloudflare challenge")
// 	}

// 	fmt.Println("Successfully bypassed Cloudflare!")
// 	return nil
// }

func (r *BrowserRenderer) Close() error {
	if r.browser == nil {
		return nil
	}

	return r.browser.Close()
}