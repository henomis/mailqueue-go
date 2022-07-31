package app

import (
	"fmt"
	"strconv"
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

const (
	DefaultSkip  = 0
	DefaultLimit = 10
)

type LimitSkip struct {
	Limit int64
	Skip  int64
}

func (ls *LimitSkip) FromString(limit, skip string) {

	if limit == "" {
		ls.Limit = DefaultLimit
	} else {
		ls.Limit, _ = strconv.ParseInt(limit, 10, 64)
	}
	if skip == "" {
		ls.Skip = DefaultSkip
	} else {
		ls.Skip, _ = strconv.ParseInt(skip, 10, 64)
	}

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
		return jsonError(c, "validate params", fmt.Errorf("invalid id"))
	}

	service := c.Params("service")
	if len(service) == 0 {
		return jsonError(c, "validate params", fmt.Errorf("invalid service"))
	}

	err := a.emailQueue.SetStatus(id, email.StatusRead)
	if err != nil {
		return jsonError(c, "emailQueue.SetStatus", err)
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
		return jsonError(c, "emailLog.Log", err)
	}

	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Content-Type", "image/gif")

	audit.Log(audit.Info, "readEmail: %s", id)

	return c.Send(whitePixelGIF)

}

func (a *App) enqueueEmail(c *fiber.Ctx) error {

	var emailToEnqueue restmodel.Email
	if err := c.BodyParser(&emailToEnqueue); err != nil {
		return jsonError(c, "bodyParser", err)
	}

	id, err := a.emailQueue.Enqueue(emailToEnqueue.ToStorageEmail())
	if err != nil {
		return jsonError(c, "emailQueue.Enqueue", err)
	}
	audit.Log(audit.Info, "emailQueue.Enqueue: %s", id)

	err = a.emailQueue.SetStatus(id, email.StatusQueued)
	if err != nil {
		return jsonError(c, "emailQueue.SetStatus", err)
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
		return jsonError(c, "emailLog.Log", err)
	}

	return c.JSON(
		restmodel.Success(
			&restmodel.EmailID{ID: id},
		),
	)

}

func (a *App) getLog(c *fiber.Ctx) error {

	id := c.Params("email_id")
	if len(id) == 0 {
		return jsonError(c, "validate params", fmt.Errorf("invalid email_id"))
	}

	storageLogItems, err := a.emailLog.Items(id)
	if err != nil {
		return jsonError(c, "emailLog.Items", err)
	}

	var logItems restmodel.LogItems
	logItems.FromStorage(storageLogItems)

	return c.JSON(
		restmodel.Success(
			&logItems,
		),
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

func (a *App) getTemplate(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "getTemplate", fmt.Errorf("id is required"))
	}

	storageTemplate, err := a.mongoTemplate.Read(id)
	if err != nil {
		return jsonError(c, "mongoTemplate.Read", err)
	}

	template := &restmodel.Template{}
	template.FromStorageTemplate(storageTemplate)

	return c.JSON(
		restmodel.Success(
			template,
		),
	)

}

func (a *App) getTemplates(c *fiber.Ctx) error {

	limitSkip := &LimitSkip{}
	limitSkip.FromString(c.Query("limit"), c.Query("skip"))

	storageTemplates, count, err := a.mongoTemplate.ReadAll(limitSkip.Limit, limitSkip.Skip)
	if err != nil {
		return jsonError(c, "mongoTemplate.ReadAll", err)
	}

	var templates restmodel.Templates
	templates.FromStorage(storageTemplates, count)

	return c.JSON(
		restmodel.Success(
			templates,
		),
	)

}

func (a *App) deleteTemplate(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "deleteTemplate", fmt.Errorf("id is required"))
	}

	err := a.mongoTemplate.Delete(id)
	if err != nil {
		return jsonError(c, "mongoTemplate.Delete", err)
	}

	return c.JSON(
		restmodel.Success(
			nil,
		),
	)

}

func (a *App) addTemplate(c *fiber.Ctx) error {

	var template restmodel.Template
	if err := c.BodyParser(&template); err != nil {
		return jsonError(c, "bodyParser", err)
	}

	if len(template.Name) > 100 {
		return jsonError(c, "validate params", fmt.Errorf("invalid name legth"))
	} else if len(template.Template) > 5000 {
		return jsonError(c, "validate params", fmt.Errorf("invalid template length"))
	}

	id, err := a.mongoTemplate.Create(template.ToStorageTemplate())
	if err != nil {
		return jsonError(c, "mongoTemplate.Create", err)
	}

	return c.JSON(
		restmodel.Success(
			&restmodel.TemplateID{ID: id},
		),
	)

}

func (a *App) updateTemplate(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "getTemplate", fmt.Errorf("id is required"))
	}

	var template restmodel.Template
	if err := c.BodyParser(&template); err != nil {
		return jsonError(c, "bodyParser", err)
	}

	if len(template.Name) > 100 {
		return jsonError(c, "validate params", fmt.Errorf("invalid name legth"))
	} else if len(template.Template) > 5000 {
		return jsonError(c, "validate params", fmt.Errorf("invalid template length"))
	}

	err := a.mongoTemplate.Update(id, template.ToStorageTemplate())
	if err != nil {
		return jsonError(c, "mongoTemplate.Create", err)
	}

	return c.JSON(
		restmodel.Success(
			&restmodel.TemplateID{ID: id},
		),
	)

}

func jsonError(c *fiber.Ctx, message string, err error) error {
	audit.Log(audit.Error, "%s: %s", message, err.Error())
	return c.JSON(
		restmodel.Error(
			err.Error(),
		),
	)
}
