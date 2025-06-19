package config

import (
	"os"
	"strconv"
	"strings"
)

type Config struct {
	ServerPort string
	LLMAPIKey     string
	LLMModel      string
	LLMMaxTokens  int
	LLMAPIBaseURL string
	BrowserTimeout   int
	CloudflareWaitMS int
	ProxyList []string
	UserAgents []string
}

func NewConfig() *Config {
	
	parseInt := func(str string, defaultVal int) int {
		if str == "" {
			return defaultVal
		}
		val, err := strconv.Atoi(str)
		if err != nil {
			return defaultVal
		}
		return val
	}

	defaultUserAgents := []string{
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/16.5 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36",
		"Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:109.0) Gecko/20100101 Firefox/114.0",
	}

	userAgents := os.Getenv("USER_AGENTS")
	var userAgentsList []string
	if userAgents != "" {
		userAgentsList = strings.Split(userAgents, ",")
	} else {
		userAgentsList = defaultUserAgents
	}

	proxyList := os.Getenv("PROXY_LIST")
	var proxies []string
	if proxyList != "" {
		proxies = strings.Split(proxyList, ",")
	}

	return &Config{
		ServerPort:       os.Getenv("PORT"),
		LLMAPIKey:        os.Getenv("LLM_API_KEY"),
		LLMModel:         getEnvWithDefault("LLM_MODEL", "gpt-3.5-turbo"),
		LLMMaxTokens:     parseInt(os.Getenv("LLM_MAX_TOKENS"), 4096),
		LLMAPIBaseURL:    getEnvWithDefault("LLM_API_BASE_URL", "https://api.openai.com"),
		BrowserTimeout:   parseInt(os.Getenv("BROWSER_TIMEOUT_SECONDS"), 30),
		CloudflareWaitMS: parseInt(os.Getenv("CLOUDFLARE_WAIT_MS"), 5000),
		ProxyList:        proxies,
		UserAgents:       userAgentsList,
	}
}

func getEnvWithDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}