package ciscofonserver

import (
	"embed"
	"log"
	"sync"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
)

//go:embed templates
var templatesFS embed.FS

var connections = make(map[*websocket.Conn]bool)
var connectionsMu sync.Mutex

var logBuffer = make([]logEntry, 0, 100)
var logBufferMu sync.Mutex

type logEntry struct {
	Time    time.Time `json:"time"`
	Service string    `json:"service"`
	Method  string    `json:"method"`
	Path    string    `json:"path"`
	Status  string    `json:"status"`
	IP      string    `json:"ip"`
}

func (s *CiscoFonServer) registerDashboardRoutes(app *fiber.App) {
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

	app.Get("/dashboard/api/logs-buffer", func(c *fiber.Ctx) error {
		logBufferMu.Lock()
		defer logBufferMu.Unlock()
		return c.JSON(logBuffer)
	})
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
	log.Printf("[%s] %s %s %s %s", service, ip, method, path, status)

	logBufferMu.Lock()
	if len(logBuffer) >= 100 {
		logBuffer = logBuffer[1:]
	}
	logBuffer = append(logBuffer, entry)
	logBufferMu.Unlock()

	connectionsMu.Lock()
	defer connectionsMu.Unlock()
	for conn := range connections {
		if err := conn.WriteJSON(entry); err != nil {
			log.Println("write:", err)
		}
	}
}
