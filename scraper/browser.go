package scraper

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/Sagn1k/scarab/config"
	"github.com/go-rod/rod"
	"github.com/go-rod/rod/lib/input"
	"github.com/go-rod/rod/lib/launcher"
	"github.com/go-rod/rod/lib/proto"
)

type RenderOptions struct {
	WaitTime  int
	Selectors []string
	BypassCF  bool
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

	// Uncomment if you want to use proxies
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
	if options != nil && options.BypassCF {
		timeoutDuration = time.Duration(r.config.BrowserTimeout*2) * time.Second
	}
	_, cancel := context.WithTimeout(ctx, timeoutDuration)
	defer cancel()

	page, err := r.browser.Page(proto.TargetCreateTarget{URL: "about:blank"})
	if err != nil {
		return "", fmt.Errorf("failed to create page: %w", err)
	}
	defer page.Close()

	_ = (&proto.EmulationSetDeviceMetricsOverride{
		Width:             1920,
		Height:            1080,
		DeviceScaleFactor: 1,
		Mobile:            false,
	}).Call(page)

	_ = rod.Try(func() {
		page.Eval(`
			// Override navigator properties to avoid detection
			Object.defineProperty(navigator, 'webdriver', {
				get: () => false,
			});
			
			// Override Chrome properties
			if (window.chrome === undefined) {
				window.chrome = {
					runtime: {},
				};
			}
			
			// Add plugins to look like a real browser
			const pluginArray = [
				{ name: 'Chrome PDF Plugin', filename: 'internal-pdf-viewer' },
				{ name: 'Chrome PDF Viewer', filename: 'mhjfbmdgcfjbbpaeojofohoefgiehjai' },
				{ name: 'Native Client', filename: 'internal-nacl-plugin' }
			];
			
			// Create a plugins property
			Object.defineProperty(navigator, 'plugins', {
				get: () => pluginArray,
				enumerable: true,
				configurable: true
			});
			
			// Override languages
			Object.defineProperty(navigator, 'languages', {
				get: () => ['en-US', 'en'],
				enumerable: true,
				configurable: true
			});
		`)
	})

	headers := r.headerRotator.GetHeaders()
	headerPairs := []string{}
	for key, value := range headers {
		headerPairs = append(headerPairs, key, value)
	}
	_, headerErr := page.SetExtraHeaders(headerPairs)
	if headerErr != nil {
		return "", fmt.Errorf("failed to set headers: %w", headerErr)
	}

	_ = proto.EmulationSetUserAgentOverride{
		UserAgent:      headers["User-Agent"],
		AcceptLanguage: "en-US,en;q=0.9",
		Platform:       "Windows",
	}.Call(page)

	err = page.Navigate(url)
	if err != nil {
		return "", fmt.Errorf("failed to navigate to URL: %w", err)
	}

	page.WaitNavigation(proto.PageLifecycleEventNameLoad)()

	_ = rod.Try(func() {
		page.Mouse.Scroll(0, float64(100+rand.Intn(300)), 5)
		time.Sleep(time.Duration(300+rand.Intn(500)) * time.Millisecond)
	})

	cfWaitTime := r.config.CloudflareWaitMS
	if options != nil && options.WaitTime > 0 {
		cfWaitTime = options.WaitTime
	}

	if options == nil || options.BypassCF {
		if err := r.handleCloudflare(page, cfWaitTime); err != nil {
			fmt.Printf("Warning: Cloudflare bypass failed: %v\n", err)
		}
	} else {
		time.Sleep(time.Duration(cfWaitTime) * time.Millisecond)
	}

	_ = rod.Try(func() {
		buttons := page.MustElements("button")
		for _, btn := range buttons {
			if txt, err := btn.Text(); err == nil {
				txtLower := strings.ToLower(txt)
				if strings.Contains(txtLower, "accept") ||
					strings.Contains(txtLower, "continue") ||
					strings.Contains(txtLower, "agree") {
					btn.Click(proto.InputMouseButtonLeft, 1)
					time.Sleep(500 * time.Millisecond)
					break
				}
			}
		}
	})

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

	var html string
	var htmlErr error
	for attempts := 0; attempts < 3; attempts++ {
		html, htmlErr = page.HTML()
		if htmlErr == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if htmlErr != nil {
		return "", fmt.Errorf("failed to get HTML after multiple attempts: %w", htmlErr)
	}

	if (strings.Contains(html, "Just a moment") || strings.Contains(html, "checking your browser")) &&
		strings.Contains(strings.ToLower(html), "cloudflare") {
		fmt.Println("Warning: Still on Cloudflare challenge page after bypass attempt")
	}

	return html, nil
}

func (r *BrowserRenderer) handleCloudflare(page *rod.Page, maxWaitTime int) error {
	isCloudflare := false

	_ = rod.Try(func() {
		bodyElem, err := page.Element("body")
		if err == nil {
			text, textErr := bodyElem.Text()
			if textErr == nil &&
				(strings.Contains(strings.ToLower(text), "cloudflare") ||
					strings.Contains(text, "Just a moment...") ||
					strings.Contains(text, "Checking your browser") ||
					strings.Contains(text, "verify you are human")) {
				isCloudflare = true
			}
		}
	})

	if !isCloudflare {
		return nil
	}

	fmt.Println("Detected Cloudflare challenge, attempting to solve...")

	time.Sleep(3 * time.Second)

	checkboxSelectors := []string{
		"input[type=checkbox]",
		".recaptcha-checkbox",
		"#checkbox",
		"#recaptcha-anchor",
		"#cf-checkbox",
		"[role=checkbox]",
		"div.checkbox",
		"span.checkbox",
		"iframe[src*='cloudflare']",
	}

	var iframeHandled bool
	_ = rod.Try(func() {
		iframes := page.MustElements("iframe")
		for _, iframe := range iframes {
			src, err := iframe.Attribute("src")
			if err == nil && src != nil &&
				(strings.Contains(*src, "cloudflare") ||
					strings.Contains(*src, "recaptcha") ||
					strings.Contains(*src, "captcha") ||
					strings.Contains(*src, "challenge")) {

				frameObj := iframe.MustFrame()

				for _, selector := range checkboxSelectors {
					err := rod.Try(func() {
						checkbox := frameObj.MustElement(selector)

						checkbox.Hover()
						time.Sleep(time.Duration(300+rand.Intn(500)) * time.Millisecond)

						checkbox.Click(proto.InputMouseButtonLeft, 1)
						fmt.Println("Clicked checkbox in iframe!")

						time.Sleep(time.Duration(2000+rand.Intn(1000)) * time.Millisecond)
						iframeHandled = true
					})
					if err == nil {
						break
					}
				}

				if iframeHandled {
					break
				}
			}
		}
	})

	if !iframeHandled {
		for _, selector := range checkboxSelectors {
			err := rod.Try(func() {
				checkbox := page.MustElement(selector)

				if checkbox.MustVisible() {
					checkbox.Hover()
					time.Sleep(time.Duration(200+rand.Intn(300)) * time.Millisecond)

					checkbox.Click(proto.InputMouseButtonLeft, 1)
					fmt.Println("Clicked checkbox on main page!")

					time.Sleep(time.Duration(2000+rand.Intn(1000)) * time.Millisecond)
				}
			})

			if err == nil {
				break
			}
		}
	}
	_ = rod.Try(func() {
		buttons := page.MustElements("button")
		for _, btn := range buttons {
			if txt, err := btn.Text(); err == nil {
				txtLower := strings.ToLower(txt)
				if strings.Contains(txtLower, "verify") ||
					strings.Contains(txtLower, "continue") ||
					strings.Contains(txtLower, "submit") ||
					strings.Contains(txtLower, "i'm human") {
					btn.Click(proto.InputMouseButtonLeft, 1)
					fmt.Println("Clicked verification button!")
					time.Sleep(2 * time.Second)
					break
				}
			}
		}
	})

	_ = rod.Try(func() {
		page.Keyboard.Press(input.Enter)
		time.Sleep(time.Second)
	})

	fmt.Println("Waiting for Cloudflare verification to complete...")
	waitDuration := time.Duration(maxWaitTime) * time.Millisecond
	if waitDuration < 10*time.Second {
		waitDuration = 10 * time.Second
	}
	time.Sleep(waitDuration)

	stillOnCloudflare := false
	_ = rod.Try(func() {
		bodyElem, err := page.Element("body")
		if err == nil {
			text, textErr := bodyElem.Text()
			if textErr == nil {
				if strings.Contains(strings.ToLower(text), "cloudflare") &&
					strings.Contains(text, "Just a moment...") {
					stillOnCloudflare = true
				}

				if strings.Contains(text, "Just a moment...") {
					stillOnCloudflare = true
				}
			}
		}
	})

	if stillOnCloudflare {
		return fmt.Errorf("failed to bypass Cloudflare challenge")
	}

	fmt.Println("Successfully bypassed Cloudflare!")
	return nil
}

func (r *BrowserRenderer) Close() error {
	if r.browser == nil {
		return nil
	}

	return r.browser.Close()
}
