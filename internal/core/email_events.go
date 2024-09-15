package core

import (
	"net/http"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// StoreEmailEvent stores an email event in the database.
func (c *Core) StoreEmailEvent(e models.EmailEvent) error {
	var newID int
	if err := c.q.StoreEmailEvent.Get(&newID, e.MessageID, e.Event, e.EventData, e.Timestamp); err != nil {
		c.log.Printf("error creating email_event: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "email_event", "error", pqErrMsg(err)))
	}
	return nil
}
