package scraper

import (
	"context"
	"fmt"

	"github.com/Sagn1k/scarab/config"
	"github.com/Sagn1k/scarab/llm"
)

type ScraperService struct {
	config      *config.Config
	renderer    *BrowserRenderer
	llmClient   *llm.Client
	proxyRotator *ProxyRotator
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
	
	waitTime := s.config.CloudflareWaitMS
	if waitParam, ok := params["waitTime"].(float64); ok {
		waitTime = int(waitParam)
	}
	

	var targetSelectors []string
	if selectors, ok := params["selectors"].([]interface{}); ok {
		for _, selector := range selectors {
			if selectorStr, ok := selector.(string); ok {
				targetSelectors = append(targetSelectors, selectorStr)
			}
		}
	}
	

	options := &RenderOptions{
		WaitTime:    waitTime,
		Selectors:   targetSelectors,
	}
	
	html, err := s.renderer.RenderPage(ctx, url, options)
	if err != nil {
		return "", fmt.Errorf("failed to render page: %w", err)
	}
	
	markdown, err := s.llmClient.HTMLToMarkdown(ctx, html, url)
	if err != nil {
		return "", fmt.Errorf("failed to convert to markdown: %w", err)
	}
	
	return markdown, nil
}