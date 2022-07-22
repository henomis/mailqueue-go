package app

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/henomis/mailqueue-go/internal/pkg/auditlogger"
	"github.com/henomis/mailqueue-go/internal/pkg/email"
)

var whitePixelGIF = []byte{
	71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0,
	255, 255, 255, 33, 249, 4, 1, 0, 0, 0, 0, 44, 0, 0, 0, 0,
	1, 0, 1, 0, 0, 2, 1, 68, 0, 59,
}

type queryParameters struct {
	Sort   string
	Offset string
	Limit  string
	Filter string
}

func getSortSkipLimitAndFilter(c *fiber.Ctx) *queryParameters {
	return &queryParameters{
		Sort:   c.Query("sort"),
		Offset: c.Query("offset"),
		Limit:  c.Query("limit"),
		Filter: c.Query("filter"),
	}
}

func (a *App) authenticationAndAuthorizationMiddleware(c *fiber.Ctx) error {

	//token := c.Get("Authorization")

	return c.Next()
}

func (a *App) setEmailAsRead(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("id is required")
	}

	service := c.Params("service")
	if len(service) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("service is required")
	}

	err := a.Queue.SetStatus(id, email.StatusRead)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	_, err = a.Log.Log(
		&email.Log{
			Timestmap: time.Now().UTC(),
			Service:   service,
			EmailID:   id,
			Status:    email.StatusRead,
		},
	)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Content-Type", "image/gif")

	a.AuditLogger.Log(auditlogger.Info, "readEmail: %s", id)

	return c.Send(whitePixelGIF)

}

func (a *App) enqueueEmail(c *fiber.Ctx) error {

	e := email.Email{}
	if err := c.BodyParser(&e); err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	id, err := a.Queue.Enqueue(&e)
	if err != nil {
		a.AuditLogger.Log(auditlogger.Error, "enqueueEmail: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}
	a.AuditLogger.Log(auditlogger.Info, "enqueueEmail: %s", id)

	err = a.Queue.SetStatus(id, email.StatusQueued)
	if err != nil {
		a.AuditLogger.Log(auditlogger.Error, "enqueueEmail: %s", err.Error())
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(id)
}

func (a *App) getLog(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("id is required")
	}

	l, err := a.Log.Items(id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).SendString(err.Error())
	}

	return c.JSON(l)
}

func (a *App) getEmail(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("id is required")
	}

	// uuid := c.Params("uuid")

	// e, err := a.Queue.GetByUUID(email.UniqueID(uuid))
	// if err != nil {
	// 	a.AuditLogger.Log(auditlogger.Error, "getEmail: %s", err.Error())
	// 	return c.Status(400).SendString(err.Error())
	// }

	// return c.JSON(e)
	return nil
}

func (a *App) template(c *fiber.Ctx) error {

	return nil
}

func (a *App) getTemplate(c *fiber.Ctx) error {

	name := c.Params("uuid")
	if len(name) > 0 {

	}

	return nil
}
