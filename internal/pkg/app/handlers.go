package app

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/henomis/mailqueue-go/internal/pkg/audit"
	"github.com/henomis/mailqueue-go/internal/pkg/email"
	"github.com/henomis/mailqueue-go/internal/pkg/restmodel"
)

var whitePixelGIF = []byte{
	71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0,
	255, 255, 255, 33, 249, 4, 1, 0, 0, 0, 0, 44, 0, 0, 0, 0,
	1, 0, 1, 0, 0, 2, 1, 68, 0, 59,
}

// type queryParameters struct {
// 	Sort   string
// 	Offset string
// 	Limit  string
// 	Filter string
// }

// func getSortSkipLimitAndFilter(c *fiber.Ctx) *queryParameters {
// 	return &queryParameters{
// 		Sort:   c.Query("sort"),
// 		Offset: c.Query("offset"),
// 		Limit:  c.Query("limit"),
// 		Filter: c.Query("filter"),
// 	}
// }

func (a *App) authenticationAndAuthorizationMiddleware(c *fiber.Ctx) error {

	//token := c.Get("Authorization")

	return c.Next()
}

func (a *App) setEmailAsRead(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "validate params", fiber.StatusBadRequest, fmt.Errorf("invalid id"))
	}

	service := c.Params("service")
	if len(service) == 0 {
		return jsonError(c, "validate params", fiber.StatusBadRequest, fmt.Errorf("invalid service"))
	}

	err := a.emailQueue.SetStatus(id, email.StatusRead)
	if err != nil {
		return jsonError(c, "emailQueue.SetStatus", fiber.StatusInternalServerError, err)
	}

	_, err = a.emailLog.Log(
		&email.Log{
			Timestamp: time.Now().UTC(),
			Service:   service,
			EmailID:   id,
			Status:    email.StatusRead,
		},
	)
	if err != nil {
		return jsonError(c, "emailLog.Log", fiber.StatusInternalServerError, err)
	}

	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Content-Type", "image/gif")

	audit.Log(audit.Info, "readEmail: %s", id)

	return c.Send(whitePixelGIF)

}

func (a *App) enqueueEmail(c *fiber.Ctx) error {

	var emailToEnqueue restmodel.Email
	if err := c.BodyParser(&emailToEnqueue); err != nil {
		return jsonError(c, "bodyParser", fiber.StatusBadRequest, err)
	}

	id, err := a.emailQueue.Enqueue(emailToEnqueue.ToStorageEmail())
	if err != nil {
		return jsonError(c, "emailQueue.Enqueue", fiber.StatusInternalServerError, err)
	}
	audit.Log(audit.Info, "emailQueue.Enqueue: %s", id)

	err = a.emailQueue.SetStatus(id, email.StatusQueued)
	if err != nil {
		return jsonError(c, "emailQueue.SetStatus", fiber.StatusInternalServerError, err)
	}
	_, err = a.emailLog.Log(
		&email.Log{
			Timestamp: time.Now().UTC(),
			Service:   emailToEnqueue.Service,
			EmailID:   id,
			Status:    email.StatusQueued,
		},
	)
	if err != nil {
		return jsonError(c, "emailLog.Log", fiber.StatusInternalServerError, err)
	}

	return c.JSON(
		&restmodel.Response{
			Status: fiber.StatusOK,
			Data:   &restmodel.EmailID{ID: id},
		},
	)
}

func (a *App) getLog(c *fiber.Ctx) error {

	id := c.Params("email_id")
	if len(id) == 0 {
		return jsonError(c, "validate params", fiber.StatusBadRequest, fmt.Errorf("invalid email_id"))
	}

	storageLogItems, err := a.emailLog.Items(id)
	if err != nil {
		return jsonError(c, "emailLog.Items", fiber.StatusInternalServerError, err)
	}

	var logItems restmodel.LogItems
	logItems.FromStorage(storageLogItems)

	return c.JSON(
		&restmodel.Response{
			Status: fiber.StatusOK,
			Data:   &logItems,
		},
	)
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

// func (a *App) getTemplate(c *fiber.Ctx) error {

// 	name := c.Params("uuid")
// 	if len(name) > 0 {

// 	}

// 	return nil
// }

func jsonError(c *fiber.Ctx, message string, status int, err error) error {
	audit.Log(audit.Error, "%s: %s", message, err.Error())
	return c.JSON(
		&restmodel.Response{
			Status: status,
			Error:  err.Error(),
		},
	)
}
