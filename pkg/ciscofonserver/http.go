package ciscofonserver

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
)

func (s *CiscoFonServer) startHTTPServer() {
	engine := html.NewFileSystem(http.FS(templatesFS), ".html")
	app := fiber.New(fiber.Config{
		Views: engine,

		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Status code defaults to 500
			code := fiber.StatusInternalServerError

			// Retrieve the custom status code if it's a *fiber.Error
			var e *fiber.Error
			if errors.As(err, &e) {
				code = e.Code
			}
			s.logRequest("HTTP", c.Method(), c.Path(), fmt.Sprintf("%d", code), c.IP())

			c.Set(fiber.HeaderContentType, fiber.MIMETextPlainCharsetUTF8)

			// Return from handler
			return c.Status(code).SendString(err.Error())
		},
	})

	app.Use(func(c *fiber.Ctx) error {
		err := c.Next()
		if !strings.HasPrefix(c.Path(), "/dashboard") && err == nil {
			s.logRequest("HTTP", c.Method(), c.Path(), "200", c.IP())
		}
		return err
	})

	app.All("*.xml", func(c *fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "text/xml")
		return c.Next()
	})

	app.Static("/", s.config.String("http.dir"))

	s.registerDashboardRoutes(app)

	log.Println("Starting HTTP server on", s.config.String("http.port"))
	err := app.Listen(":" + s.config.String("http.port"))
	if err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}
