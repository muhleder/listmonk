package core

import (
	"net/http"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetEmail gets an email by it's message id.
func (c *Core) GetEmailByMessageId(message_id string) (models.Email, error) {
	var res []models.Email
	if err := c.q.GetEmailByMessageId.Select(&res, message_id); err != nil {
		c.log.Printf("error fetching emails for opt-in: %s", pqErrMsg(err))
		return models.Email{}, echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorFetching", "name", "email", "error", pqErrMsg(err)))
	}
	if len(res) == 0 {
		return models.Email{}, echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "email"))
	}
	out := res[0]
	return out, nil
}

// CreateEmail stores an email in the database.
func (c *Core) CreateEmail(e models.Email) error {
	var newID int
	if err := c.q.CreateEmail.Get(&newID, e.CampaignID, e.MessageID, e.Recipient, e.Source, e.Subject, e.Status, e.SentAt); err != nil {
		c.log.Printf("error creating email: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "email", "error", pqErrMsg(err)))
	}
	return nil
}

// UpdateEmail updates an email matching a specific message id.
func (c *Core) UpdateEmailStatus(message_id string, status string) error {
	res, err := c.q.UpdateEmail.Exec(message_id, status)
	if err != nil {
		c.log.Printf("error updating email: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorUpdating", "name", "email", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "name", "email"))
	}
	return nil
}
