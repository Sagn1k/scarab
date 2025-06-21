package scraper

import (
	"math/rand"
	"sync"
	"time"
)

type HeaderRotator struct {
	userAgents []string
	mu         sync.Mutex
	rng        *rand.Rand
}

func NewHeaderRotator(userAgents []string) *HeaderRotator {
	return &HeaderRotator{
		userAgents: userAgents,
		rng:        rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *HeaderRotator) GetRandomUserAgent() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.userAgents) == 0 {
		return "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36"
	}

	return r.userAgents[r.rng.Intn(len(r.userAgents))]
}

func (r *HeaderRotator) GetHeaders() map[string]string {
	ua := r.GetRandomUserAgent()

	headers := map[string]string{
		"User-Agent":                ua,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"Accept-Language":           "en-US,en;q=0.5",
		"Accept-Encoding":           "gzip, deflate, br",
		"DNT":                       "1",
		"Connection":                "keep-alive",
		"Upgrade-Insecure-Requests": "1",
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
	}

	return headers
}

type ProxyRotator struct {
	proxies []string
	index   int
	mu      sync.Mutex
	rng     *rand.Rand
}

func NewProxyRotator(proxies []string) *ProxyRotator {
	return &ProxyRotator{
		proxies: proxies,
		index:   0,
		rng:     rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (r *ProxyRotator) GetNextProxy() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.proxies) == 0 {
		return ""
	}

	proxy := r.proxies[r.index]
	r.index = (r.index + 1) % len(r.proxies)

	return proxy
}

func (r *ProxyRotator) GetRandomProxy() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.proxies) == 0 {
		return ""
	}

	return r.proxies[r.rng.Intn(len(r.proxies))]
}
