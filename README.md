# Web Scraper with LLM Markdown Generation

A scalable web scraper built in Go that uses Large Language Models (LLMs) to convert web page content into clean, well-structured markdown. It features proxy rotation, header rotation, and can handle JavaScript-rendered pages using the Rod framework.

## Features

- **LLM-Powered Content Extraction**: Automatically converts web page content to markdown using LLMs (e.g., OpenAI GPT models)
- **Dynamic Content Rendering**: Uses [Rod](https://github.com/go-rod/rod) to render JavaScript-based pages
- **Cloudflare Bypass**: Configurable wait times to bypass Cloudflare and similar protection mechanisms
- **Proxy Rotation**: Supports random and sequential rotation of proxies to avoid IP bans
- **Header Rotation**: Rotates User-Agent headers to appear as different browsers
- **REST API**: Built with [Fiber](https://github.com/gofiber/fiber) for high-performance endpoints
- **Modular Design**: Well-organized components for easy maintenance and extension

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/webscraper.git
   cd webscraper
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Configure environment variables:
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

4. Build the application:
   ```bash
   go build -o webscraper
   ```

## Usage

### Running the API Server

```bash
./webscraper
```

By default, the server runs on port 3000. You can change this in the `.env` file.

### Making API Requests

To scrape a webpage and convert it to markdown:

```bash
curl -X POST http://localhost:3000/scrape \
  -H "Content-Type: application/json" \
  -d '{
    "url": "https://example.com/page-to-scrape",
    "params": {
      "waitTime": 5000,
      "selectors": ["#main-content", ".article-body"]
    }
  }'
```

## Configuration

Configure the application using environment variables or the `.env` file:

| Variable | Description | Default |
|----------|-------------|---------|
| PORT | Server port | 3000 |
| LLM_API_KEY | API key for the LLM service | - |
| LLM_MODEL | Model to use for markdown generation | gpt-3.5-turbo |
| LLM_MAX_TOKENS | Maximum number of tokens for LLM response | 4096 |
| LLM_API_BASE_URL | Base URL for LLM API | https://api.openai.com |
| BROWSER_TIMEOUT_SECONDS | Maximum time to wait for browser operations | 30 |
| CLOUDFLARE_WAIT_MS | Wait time for Cloudflare bypass | 5000 |
| PROXY_LIST | Comma-separated list of proxies | - |
| USER_AGENTS | Comma-separated list of user agents | (predefined list) |

## Project Structure

```
webscraper/
├── main.go           # Entry point
├── api/              # API server and routes
├── config/           # Configuration handling
├── errors/           # Error definitions
├── llm/              # LLM client for markdown conversion
├── renderer/         # Browser renderer using Rod
├── scraper/          # Core scraping logic
│   └── rotator.go    # Proxy and header rotation
└── .env.example      # Example environment configuration
```

## License

MIT License