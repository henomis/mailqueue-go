package httpserver

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/henomis/mailqueue-go/internal/pkg/audit"
	"github.com/henomis/mailqueue-go/internal/pkg/restmodel"
	"github.com/henomis/mailqueue-go/internal/pkg/storagemodel"
)

var whitePixelGIF = []byte{
	71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0,
	255, 255, 255, 33, 249, 4, 1, 0, 0, 0, 0, 44, 0, 0, 0, 0,
	1, 0, 1, 0, 0, 2, 1, 68, 0, 59,
}

const (
	DefaultSkip      = 0
	DefaultLimit     = 10
	DefaultUnlimited = 10000
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

func (h *HTTPServer) authenticationAndAuthorizationMiddleware(c *fiber.Ctx) error {

	//token := c.Get("Authorization")

	return c.Next()
}

// ----
// LOGS
// ----

func (h *HTTPServer) getLogs(c *fiber.Ctx) error {
	limitSkip := &LimitSkip{}
	limitSkip.FromString(c.Query("limit"), c.Query("skip"))
	fields := c.Query("fields")

	storageLogs, count, err := h.emailLog.GetAll(limitSkip.Limit, limitSkip.Skip, fields)
	if err != nil {
		return jsonError(c, "mongoTemplate.ReadAll", err)
	}

	var logs restmodel.LogsCount
	logs.FromStorageModel(storageLogs, count)

	return c.JSON(
		restmodel.Success(
			logs,
		),
	)
}

func (h *HTTPServer) getLog(c *fiber.Ctx) error {

	id := c.Params("email_id")
	if len(id) == 0 {
		return jsonError(c, "validate params", fmt.Errorf("invalid email_id"))
	}

	storageLogItems, err := h.emailLog.Get(id)
	if err != nil {
		return jsonError(c, "emailLog.Items", err)
	}

	var logItems restmodel.Logs
	logItems.FromStorageModel(storageLogItems)

	return c.JSON(
		restmodel.Success(
			&logItems,
		),
	)
}

// ------
// EMAILS
// ------

func (h *HTTPServer) getEmail(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return c.Status(fiber.StatusBadRequest).SendString("id is required")
	}

	storageEmail, err := h.emailQueue.Get(id)
	if err != nil {
		return jsonError(c, "emailQueue.Get", err)
	}

	var email restmodel.Email
	email.FromStorageModel(storageEmail)

	return c.JSON(
		restmodel.Success(
			email,
		),
	)
}

func (h *HTTPServer) getEmails(c *fiber.Ctx) error {

	var storageEmails []storagemodel.Email
	var count int64
	var err error

	limitSkip := &LimitSkip{}
	limitSkip.FromString(c.Query("limit"), c.Query("skip"))
	fields := c.Query("fields")
	mode := c.Query("mode")

	if mode == "with_logs" {
		storageEmails, count, err = h.emailQueue.GetAllWithLogs(limitSkip.Limit, limitSkip.Skip)
		if err != nil {
			return jsonError(c, "emailQueue.GetAll", err)
		}
	} else {
		storageEmails, count, err = h.emailQueue.GetAll(limitSkip.Limit, limitSkip.Skip, fields)
		if err != nil {
			return jsonError(c, "emailQueue.GetAll", err)
		}
	}

	var emails restmodel.EmailsCount
	emails.FromStorageModel(storageEmails, count)

	return c.JSON(
		restmodel.Success(
			emails,
		),
	)

}

func (h *HTTPServer) enqueueEmail(c *fiber.Ctx) error {

	var emailToEnqueue restmodel.Email
	if err := c.BodyParser(&emailToEnqueue); err != nil {
		return jsonError(c, "bodyParser", err)
	}

	id, err := h.emailQueue.Enqueue(emailToEnqueue.ToStorageModel())
	if err != nil {
		return jsonError(c, "emailQueue.Enqueue", err)
	}
	audit.Log(audit.Info, "emailQueue.Enqueue: %s", id)

	err = h.emailQueue.SetStatus(id, storagemodel.StatusQueued)
	if err != nil {
		return jsonError(c, "emailQueue.SetStatus", err)
	}
	_, err = h.emailLog.Create(
		&storagemodel.Log{
			Timestamp: time.Now().UTC(),
			Service:   emailToEnqueue.Service,
			EmailID:   id,
			Status:    storagemodel.StatusQueued,
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

func (h *HTTPServer) setEmailAsRead(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "validate params", fmt.Errorf("invalid id"))
	}

	service := c.Params("service")
	if len(service) == 0 {
		return jsonError(c, "validate params", fmt.Errorf("invalid service"))
	}

	err := h.emailQueue.SetStatus(id, storagemodel.StatusRead)
	if err != nil {
		return jsonError(c, "emailQueue.SetStatus", err)
	}

	_, err = h.emailLog.Create(
		&storagemodel.Log{
			Timestamp: time.Now().UTC(),
			Service:   service,
			EmailID:   id,
			Status:    storagemodel.StatusRead,
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

// ---------
// TEMPLATES
// ---------

func (h *HTTPServer) getTemplate(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "getTemplate", fmt.Errorf("id is required"))
	}

	storageTemplate, err := h.emailTemplate.Get(id)
	if err != nil {
		return jsonError(c, "mongoTemplate.Read", err)
	}

	template := &restmodel.Template{}
	template.FromStorageModel(storageTemplate)

	return c.JSON(
		restmodel.Success(
			template,
		),
	)

}

func (h *HTTPServer) getTemplates(c *fiber.Ctx) error {

	limitSkip := &LimitSkip{}
	limitSkip.FromString(c.Query("limit"), c.Query("skip"))
	fields := c.Query("fields")

	storageTemplates, count, err := h.emailTemplate.GetAll(limitSkip.Limit, limitSkip.Skip, fields)
	if err != nil {
		return jsonError(c, "mongoTemplate.ReadAll", err)
	}

	var templates restmodel.TemplatesCount
	templates.FromStorageModel(storageTemplates, count)

	return c.JSON(
		restmodel.Success(
			templates,
		),
	)

}

func (h *HTTPServer) deleteTemplate(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "deleteTemplate", fmt.Errorf("id is required"))
	}

	err := h.emailTemplate.Delete(id)
	if err != nil {
		return jsonError(c, "mongoTemplate.Delete", err)
	}

	return c.JSON(
		restmodel.Success(
			nil,
		),
	)

}

func (h *HTTPServer) createTemplate(c *fiber.Ctx) error {

	var template restmodel.Template
	if err := c.BodyParser(&template); err != nil {
		return jsonError(c, "bodyParser", err)
	}

	if err := validateTemplate(&template); err != nil {
		return jsonError(c, "validateTemplate", err)
	}

	id, err := h.emailTemplate.Create(template.ToStorageModel())
	if err != nil {
		return jsonError(c, "mongoTemplate.Create", err)
	}

	return c.JSON(
		restmodel.Success(
			&restmodel.TemplateID{ID: id},
		),
	)

}

func (h *HTTPServer) updateTemplate(c *fiber.Ctx) error {

	id := c.Params("id")
	if len(id) == 0 {
		return jsonError(c, "getTemplate", fmt.Errorf("id is required"))
	}

	var template restmodel.Template
	if err := c.BodyParser(&template); err != nil {
		return jsonError(c, "bodyParser", err)
	}

	if err := validateTemplate(&template); err != nil {
		return jsonError(c, "validateTemplate", err)
	}

	err := h.emailTemplate.Update(id, template.ToStorageModel())
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

func validateTemplate(template *restmodel.Template) error {
	if len(template.Name) == 0 || len(template.Name) > 50 {
		return fmt.Errorf("invalid name length")
	}
	if len(template.Template) == 0 || len(template.Template) > 5000 {
		return fmt.Errorf("invalid template length")
	}
	return nil
}
