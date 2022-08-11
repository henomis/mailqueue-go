package httpserver

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/limiter"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/henomis/mailqueue-go/internal/pkg/app"
)

type HTTPServer struct {
	fiberInstance *fiber.App
	emailQueue    app.EmailQueue
	emailLog      app.EmailLog
	emailTemplate app.EmailTemplate
}

func New(
	emailQueue app.EmailQueue,
	emailLog app.EmailLog,
	emailTemplate app.EmailTemplate,
) *HTTPServer {

	fiberInstance := fiber.New(fiber.Config{
		StrictRouting: true,
	})

	fiberInstance.Use(logger.New())
	fiberInstance.Use(cors.New())
	fiberInstance.Use(limiter.New(limiter.Config{
		Max:        200,
		Expiration: 1 * time.Minute,
	}))

	return &HTTPServer{
		emailQueue:    emailQueue,
		emailLog:      emailLog,
		emailTemplate: emailTemplate,
		fiberInstance: fiberInstance,
	}
}

func (h *HTTPServer) Run(bindAddress string) error {

	h.setupRoutes()

	return h.fiberInstance.Listen(bindAddress)
}

func (h *HTTPServer) setupRoutes() {

	h.fiberInstance.Get("/api/v1/images/mail/:service/:id", h.setEmailAsRead)
	h.fiberInstance.Use("/api/v1", h.authenticationAndAuthorizationMiddleware)

	// LOGS
	h.fiberInstance.Get("/api/v1/logs", h.getLogs)
	h.fiberInstance.Get("/api/v1/logs/:email_id", h.getLog)

	// EMAILS
	h.fiberInstance.Get("/api/v1/emails", h.getEmails)
	h.fiberInstance.Get("/api/v1/emails/:id", h.getEmail)
	h.fiberInstance.Post("/api/v1/emails", h.enqueueEmail)

	// TEMPLATES
	h.fiberInstance.Get("/api/v1/templates", h.getTemplates)
	h.fiberInstance.Get("/api/v1/templates/:id", h.getTemplate)
	h.fiberInstance.Put("/api/v1/templates/:id", h.updateTemplate)
	h.fiberInstance.Post("/api/v1/templates", h.createTemplate)
	h.fiberInstance.Delete("/api/v1/templates/:id", h.deleteTemplate)

}
