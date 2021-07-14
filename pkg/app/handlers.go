package app

import (
	"github.com/gofiber/fiber/v2"
	"github.com/henomis/mailqueue-go/pkg/email"
	"github.com/henomis/mailqueue-go/pkg/trace"
)

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

func (a *App) readEmail(c *fiber.Ctx) error {

	var GIF = []byte{
		71, 73, 70, 56, 57, 97, 1, 0, 1, 0, 128, 0, 0, 0, 0, 0,
		255, 255, 255, 33, 249, 4, 1, 0, 0, 0, 0, 44, 0, 0, 0, 0,
		1, 0, 1, 0, 0, 2, 1, 68, 0, 59,
	}

	uuid := c.Params("uuid")

	a.Queue.SetStatus(&email.Email{UUID: email.UniqueID(uuid)}, email.StatusRead)

	c.Set("Cache-Control", "no-cache, no-store, must-revalidate")
	c.Set("Content-Type", "image/gif")

	a.Tracer.Trace(trace.Info, "readEmail: %s", uuid)

	return c.Send(GIF)

}

func (a *App) enqueueEmail(c *fiber.Ctx) error {

	e := email.Email{}
	if err := c.BodyParser(&e); err != nil {
		return err
	}

	uuid, err := a.Queue.Enqueue(&e)
	if err != nil {
		a.Tracer.Trace(trace.Error, "enqueueEmail: %s", err.Error())
		return c.Status(400).SendString(err.Error())
	}

	a.Tracer.Trace(trace.Info, "enqueueEmail: %s", string(uuid))
	a.Queue.SetStatus(&email.Email{UUID: email.UniqueID(uuid)}, email.StatusQueued)

	return c.JSON(uuid)
}

func (a *App) getLog(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	if len(uuid) > 0 {
		l, err := a.Log.GetByUUID(email.UniqueID(uuid))
		if err != nil {
			return c.Status(400).SendString(err.Error())
		}
		return c.JSON(l)
	}

	return nil
}

func (a *App) getEmail(c *fiber.Ctx) error {
	uuid := c.Params("uuid")

	e, err := a.Queue.GetByUUID(email.UniqueID(uuid))
	if err != nil {
		a.Tracer.Trace(trace.Error, "getEmail: %s", err.Error())
		return c.Status(400).SendString(err.Error())
	}

	return c.JSON(e)
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
