package scraper

import (
	"context"
	"fmt"
	"github.com/Sagn1k/scarab/config"
	"github.com/Sagn1k/scarab/llm"
	"strings"
)

type ScraperService struct {
	config        *config.Config
	renderer      *BrowserRenderer
	llmClient     *llm.Client
	proxyRotator  *ProxyRotator
	headerRotator *HeaderRotator
}

func NewScraperService(cfg *config.Config) *ScraperService {

	proxyRotator := NewProxyRotator(cfg.ProxyList)
	headerRotator := NewHeaderRotator(cfg.UserAgents)
	browserRenderer := NewBrowserRenderer(cfg, proxyRotator, headerRotator)
	llmClient := llm.NewClient(cfg)

	return &ScraperService{
		config:        cfg,
		renderer:      browserRenderer,
		llmClient:     llmClient,
		proxyRotator:  proxyRotator,
		headerRotator: headerRotator,
	}
}

func (s *ScraperService) Scrape(ctx context.Context, url string, params map[string]interface{}) (string, error) {
	// Extract wait time parameter if provided, otherwise use default
	waitTime := s.config.CloudflareWaitMS
	if waitParam, ok := params["waitTime"].(float64); ok {
		waitTime = int(waitParam)
	}

	// Extract specific elements if provided
	var targetSelectors []string
	if selectors, ok := params["selectors"].([]interface{}); ok {
		for _, selector := range selectors {
			if selectorStr, ok := selector.(string); ok {
				targetSelectors = append(targetSelectors, selectorStr)
			}
		}
	}

	// Check if Cloudflare bypass is enabled/disabled
	bypassCF := true // Default to true
	if bypassValue, ok := params["bypassCloudflare"].(bool); ok {
		bypassCF = bypassValue
	}

	// Configure renderer options
	options := &RenderOptions{
		WaitTime:  waitTime,
		Selectors: targetSelectors,
		BypassCF:  bypassCF,
	}

	// Try multiple strategies if Cloudflare bypass is enabled
	maxRetries := 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		// Attempt to render the page
		html, err := s.renderer.RenderPage(ctx, url, options)
		if err != nil {
			if attempt < maxRetries-1 {
				// Retry with a different proxy if available
				s.proxyRotator.GetNextProxy()
				fmt.Printf("Render failed on attempt %d, retrying with new proxy...\n", attempt+1)
				continue
			}
			return "", fmt.Errorf("failed to render page after %d attempts: %w", maxRetries, err)
		}

		// Check if we're still on the Cloudflare challenge page
		if bypassCF && strings.Contains(html, "Just a moment") &&
			strings.Contains(strings.ToLower(html), "cloudflare") {

			if attempt < maxRetries-1 {
				// Adjust strategy for next attempt
				fmt.Printf("Still hitting Cloudflare on attempt %d, adjusting strategy...\n", attempt+1)
				options.WaitTime = options.WaitTime * 2 // Double the wait time
				continue
			}

			// If we're on the last attempt and still hitting Cloudflare, just use what we have
			fmt.Println("Warning: Could not bypass Cloudflare after all attempts")
		}

		// Process the HTML with LLM to generate markdown
		markdown, err := s.llmClient.HTMLToMarkdown(ctx, html, url)
		if err != nil {
			return "", fmt.Errorf("failed to convert to markdown: %w", err)
		}

		return markdown, nil
	}

	return "", fmt.Errorf("failed to render page after multiple attempts")
}
