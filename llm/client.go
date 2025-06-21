package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/Sagn1k/scarab/config"
)

type Client struct {
	config     *config.Config
	httpClient *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{
		config:     cfg,
		httpClient: &http.Client{},
	}
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenAIRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
	Temperature float64   `json:"temperature,omitempty"`
}

type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Choices []struct {
		Index        int     `json:"index"`
		Message      Message `json:"message"`
		FinishReason string  `json:"finish_reason"`
	} `json:"choices"`
}

func (c *Client) HTMLToMarkdown(ctx context.Context, html string, url string) (string, error) {
	if len(html) > 100000 {
		html = html[:100000] + "..."
	}

	systemPrompt := fmt.Sprintf(`You are an expert web content extractor. 
Your task is to analyze the given HTML content from the URL: %s
and convert it to clean, well-formatted markdown. 

Focus on extracting the meaningful content:
1. Extract the main article content, title, and key information
2. Preserve important text formatting (headers, bold, italic, lists, links)
3. Include relevant images by noting [Image: description] in markdown
4. Organize the content in a logical structure
5. Remove advertisements, navigation menus, footers, and other boilerplate content
6. Preserve tables if they contain important data
7. Do not include any JavaScript or HTML tags in your response
8. Format code blocks properly if present

Return ONLY the markdown content with no additional explanations or notes.`, url)

	userMessage := fmt.Sprintf("Here is the HTML content to convert to markdown:\n\n%s", html)

	request := OpenAIRequest{
		Model:       c.config.LLMModel,
		MaxTokens:   c.config.LLMMaxTokens,
		Temperature: 0.1,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
	}

	response, err := c.callAPI(ctx, request)
	if err != nil {
		return "", fmt.Errorf("LLM API error: %w", err)
	}

	if len(response.Choices) == 0 || response.Choices[0].Message.Content == "" {
		return "", errors.New("LLM returned empty response")
	}

	return response.Choices[0].Message.Content, nil
}

func (c *Client) callAPI(ctx context.Context, req OpenAIRequest) (*OpenAIResponse, error) {
	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	apiURL := strings.TrimSuffix(c.config.LLMAPIBaseURL, "/") + "/v1/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.config.LLMAPIKey)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned non-200 status: %d, body: %s", resp.StatusCode, string(body))
	}

	var apiResp OpenAIResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("error parsing response: %w", err)
	}

	return &apiResp, nil
}
