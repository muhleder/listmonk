package core

import (
	"net/http"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// GetEmailByMessageId gets an email by it's message id.
func (c *Core) GetEmailByMessageId(message_id string) (models.Email, error) {
	var res []models.Email
	if err := c.q.GetEmailByMessageId.Select(&res, message_id); err != nil {
		c.log.Printf("error fetching email by message id: %s", pqErrMsg(err))
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

func (c *Core) GetEmailByCampaignSubscriberUUID(campaign_uuid string, subscriber_uuid string) (models.Email, error) {
	var res []models.Email
	if err := c.q.GetEmailByCampaignSubscriberUUID.Get(&res, campaign_uuid, subscriber_uuid); err != nil {
		c.log.Printf("error fetching email by campaign and subscriber id: %s", pqErrMsg(err))
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

// StoreEmail stores an email in the database.
func (c *Core) StoreEmail(e models.Email) error {
	if _, err := c.q.StoreEmail.Exec(e.MessageID, e.CampaignUUID, e.SubscriberUUID, e.MessageID, e.Recipient, e.Source, e.Subject, e.Status, e.SentAt); err != nil {
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
			c.i18n.Ts("globals.messages.errorUpdating", "status", "email", "error", pqErrMsg(err)))
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return echo.NewHTTPError(http.StatusBadRequest,
			c.i18n.Ts("globals.messages.notFound", "status", "email"))
	}
	return nil
}
