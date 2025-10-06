package ciscofonserver

import (
	"embed"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"

	"github.com/pin/tftp/v3"
)

var connections = make(map[*websocket.Conn]bool)
var connectionsMu sync.Mutex

type logEntry struct {
	Time    time.Time `json:"time"`
	Service string    `json:"service"`
	Method  string    `json:"method"`
	Path    string    `json:"path"`
	Status  string    `json:"status"`
	IP      string    `json:"ip"`
}

type tftpHandler struct {
	s *CiscoFonServer
}

// writeHandler is called when client starts file upload to server
func (s *CiscoFonServer) writeHandler(filename string, wt io.WriterTo) error {
	return nil
}

// readHandler is called when client starts file download from server
func (s *CiscoFonServer) readHandler(filename string, rf io.ReaderFrom) error {
	cleanedFilename := cleanPath(filename)
	filepath := filepath.Join(s.config.String("tftp.dir"), cleanedFilename)
	file, err := os.Open(filepath)
	if err != nil {
		log.Printf("`TFTP` GET %s: %v", filename, err)
		return err
	}
	defer file.Close()
	_, err = rf.ReadFrom(file)
	if err != nil {
		log.Printf("TFTP GET %s: %v", filename, err)
		return err
	}

	return nil
}

func (s *CiscoFonServer) OnSuccess(stats tftp.TransferStats) {
	s.logRequest("TFTP", "READ", stats.Filename, "OK", stats.RemoteAddr.String())
}

func (s *CiscoFonServer) OnFailure(stats tftp.TransferStats, err error) {
	s.logRequest("TFTP", "READ", stats.Filename, "Err: "+err.Error(), stats.RemoteAddr.String())
}

func (s *CiscoFonServer) startTFTPServer() {
	srv := tftp.NewServer(s.readHandler, s.writeHandler)
	srv.SetHook(s)
	srv.SetTimeout(5 * time.Second)  // optional
	err := srv.ListenAndServe(":69") // blocks until s.Shutdown() is called
	if err != nil {
		fmt.Fprintf(os.Stdout, "server: %v\n", err)
		os.Exit(1)
	}
}

//go:embed templates
var templatesFS embed.FS

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

	app.Get("/dashboard", func(c *fiber.Ctx) error {
		return c.Render("templates/dashboard", fiber.Map{})
	})

	app.Get("/dashboard/log-ws", websocket.New(func(c *websocket.Conn) {
		connectionsMu.Lock()
		connections[c] = true
		connectionsMu.Unlock()
		defer func() {
			connectionsMu.Lock()
			delete(connections, c)
			connectionsMu.Unlock()
		}()

		for {
			if _, _, err := c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
		}
	}))

	log.Println("Starting HTTP server on", s.config.String("http.port"))
	err := app.Listen(":" + s.config.String("http.port"))
	if err != nil {
		log.Fatalf("HTTP server error: %v", err)
	}
}

func (s *CiscoFonServer) logRequest(service, method, path string, status string, ip string) {
	entry := logEntry{
		Time:    time.Now(),
		Service: service,
		Method:  method,
		Path:    path,
		Status:  status,
		IP:      ip,
	}
	log.Printf("[%s] %s %s %s %d", service, ip, method, path, status)
	connectionsMu.Lock()
	defer connectionsMu.Unlock()
	for conn := range connections {
		if err := conn.WriteJSON(entry); err != nil {
			log.Println("write:", err)
		}
	}
}
