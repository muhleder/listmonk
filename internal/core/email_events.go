package core

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
)

// StoreEmailEvent stores an email event in the database.
func (c *Core) StoreEmailEvent(e models.EmailEvent) error {
	// We should have either a message id or a campaign/subscriber id pair.
	var email models.Email
	switch true {
	case e.MessageID != "":
		email, _ = c.GetEmailByMessageId(e.MessageID)
	case e.CampaignUUID != "" && e.SubscriberUUID != "":
		email, _ = c.GetEmailByCampaignSubscriberUUID(e.CampaignUUID, e.SubscriberUUID)
	default:
		c.log.Printf("Missing message id or campaign/subscriber ids when saving EmailEvent. Timestamp: %v", e.Timestamp)
	}

	if _, err := c.q.StoreEmailEvent.Exec(email.ID, e.MessageID, e.CampaignUUID, e.SubscriberUUID, e.Event, e.EventData, e.Timestamp); err != nil {
		c.log.Printf("error creating email_event: %v", err)
		return echo.NewHTTPError(http.StatusInternalServerError,
			c.i18n.Ts("globals.messages.errorCreating", "name", "email_event", "error", pqErrMsg(err)))
	}
	return nil
}

func (c *Core) RegisterOpenEmailEvent(campUUID, subUUID string, context echo.Context) {
	eventData := struct {
		IpAddress string
		UserAgent string
		Referer   string
	}{
		IpAddress: context.RealIP(),
		UserAgent: context.Request().UserAgent(),
		Referer:   context.Request().Referer(),
	}
	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		c.log.Printf("error encoding JSON: %v", err)
	}
	email, err := c.GetEmailByCampaignSubscriberUUID(campUUID, subUUID)
	if err != nil {
		c.log.Printf("error getting email in RegisterOpenEmailEvent: %v %v %s", campUUID, subUUID, err)
	}
	event := models.EmailEvent{
		EmailID:        email.ID,
		CampaignUUID:   campUUID,
		SubscriberUUID: subUUID,
		Event:          "open",
		EventData:      json.RawMessage(jsonBytes),
		Timestamp:      time.Now(),
	}
	c.StoreEmailEvent(event)
}

func (c *Core) RegisterClickEmailEvent(campUUID, subUUID, linkUUID string, context echo.Context) {
	eventData := struct {
		IpAddress string
		UserAgent string
		Referer   string
		LinkUUID  string
	}{
		IpAddress: context.RealIP(),
		UserAgent: context.Request().UserAgent(),
		Referer:   context.Request().Referer(),
		LinkUUID:  linkUUID,
	}
	jsonBytes, err := json.Marshal(eventData)
	if err != nil {
		c.log.Printf("error encoding JSON: %v", err)
	}
	email, err := c.GetEmailByCampaignSubscriberUUID(campUUID, subUUID)
	if err != nil {
		c.log.Printf("error getting email in RegisterClickEmailEvent: %v %v %s", campUUID, subUUID, err)
	}
	event := models.EmailEvent{
		EmailID:        email.ID,
		CampaignUUID:   campUUID,
		SubscriberUUID: subUUID,
		Event:          "click",
		EventData:      json.RawMessage(jsonBytes),
		Timestamp:      time.Now(),
	}
	c.StoreEmailEvent(event)
}
