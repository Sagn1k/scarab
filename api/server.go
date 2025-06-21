package api

import (
	"log"

	"github.com/Sagn1k/scarab/config"
	"github.com/Sagn1k/scarab/scraper"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

type Server struct {
	app    *fiber.App
	config *config.Config
}

func NewServer(cfg *config.Config) *Server {
	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(logger.New())
	app.Use(recover.New())

	server := &Server{
		app:    app,
		config: cfg,
	}

	server.registerRoutes()

	return server
}

func (s *Server) Start() {
	port := s.config.ServerPort
	if port == "" {
		port = "3000"
	}

	log.Printf("Server starting on port %s", port)
	log.Fatal(s.app.Listen(":" + port))
}

func (s *Server) registerRoutes() {
	s.app.Get("/", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status":  "ok",
			"message": "Scarab API is running",
		})
	})

	s.setupScraperRoutes()
}

func (s *Server) setupScraperRoutes() {
	scraperService := scraper.NewScraperService(s.config)

	s.app.Post("/scrape", func(c *fiber.Ctx) error {
		var req ScrapeRequest
		if err := c.BodyParser(&req); err != nil {
			return err
		}

		if req.URL == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": "URL is required",
			})
		}

		result, err := scraperService.Scrape(c.Context(), req.URL, req.Params)
		if err != nil {
			return err
		}

		return c.JSON(ScrapeResponse{
			Success:  true,
			Markdown: result,
		})
	})
}

type ScrapeRequest struct {
	URL    string                 `json:"url"`
	Params map[string]interface{} `json:"params"`
}

type ScrapeResponse struct {
	Success  bool   `json:"success"`
	Markdown string `json:"markdown"`
}
