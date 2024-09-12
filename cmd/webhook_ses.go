package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/knadh/listmonk/models"
	"github.com/labstack/echo/v4"
	sns "github.com/robbiet480/go.sns"
)

type sesMail struct {
	EventType string `json:"eventType"`
	NotifType string `json:"notificationType"`
	Bounce    struct {
		BounceType        string `json:"bounceType"`
		BounceSubType     string `json:"bounceSubType"`
		BouncedRecipients []struct {
			EmailAddress   string `json:"emailAddress"`
			Action         string `json:"action"`
			Status         string `json:"status"`
			DiagnosticCode string `json:"diagnosticCode"`
		} `json:"bouncedRecipients"`
		Timestamp    time.Time `json:"timestamp"`
		FeedbackID   string    `json:"feedbackId"`
		ReportingMTA string    `json:"reportingMTA"`
	} `json:"bounce"`
	Complaint struct {
		ComplainedRecipients []struct {
			EmailAddress string `json:"emailAddress"`
		} `json:"complainedRecipients"`
		Timestamp             time.Time `json:"timestamp"`
		FeedbackID            string    `json:"feedbackId"`
		UserAgent             string    `json:"userAgent"`
		ComplaintFeedbackType string    `json:"complaintFeedbackType"`
		ArrivalDate           time.Time `json:"arrivalDate"`
	} `json:"complaint"`
	Delivery struct {
		Timestamp    time.Time `json:"timestamp"`
		Recipients   []string  `json:"recipients"`
		SmtpResponse string    `json:"smtpResponse"`
	} `json:"delivery"`
	Send struct {
	} `json:"send"`
	Reject struct {
		Reason string `json:"reason"`
	} `json:"reject"`
	Open struct {
		IpAddress string    `json:"ipAddress"`
		Timestamp time.Time `json:"timestamp"`
		UserAgent string    `json:"userAgent"`
	} `json:"open"`
	Click struct {
		IpAddress string              `json:"ipAddress"`
		Timestamp time.Time           `json:"timestamp"`
		UserAgent string              `json:"userAgent"`
		Link      string              `json:"link"`
		LinkTags  map[string][]string `json:"linkTags"`
	} `json:"click"`
	Mail struct {
		Timestamp        time.Time           `json:"timestamp"`
		MessageID        string              `json:"messageId"`
		Source           string              `json:"source"`
		HeadersTruncated bool                `json:"headersTruncated"`
		Destination      []string            `json:"destination"`
		Headers          []map[string]string `json:"headers"`
		CommonHeaders    struct {
			From      []string `json:"from"`
			To        []string `json:"to"`
			MessageID string   `json:"messageId"`
			Subject   string   `json:"subject"`
		}
	} `json:"mail"`
}

func handleSesNotificationWebhook(c echo.Context) error {
	var (
		app = c.Get("app").(*App)

		m                   sesMail
		event               models.EmailEvent
		notificationPayload sns.Payload
	)

	// Read the request body instead of using c.Bind() to read to save the entire raw request as meta.
	rawReq, err := io.ReadAll(c.Request().Body)
	if err != nil {
		app.log.Printf("error reading ses notification body: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.Ts("globals.messages.internalError"))
	}

	if err := json.Unmarshal([]byte(rawReq), &notificationPayload); err != nil {
		app.log.Printf("could not unmarshall verify for SES notification: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
	}

	if err := notificationPayload.VerifyPayload(); err != nil {
		app.log.Printf("could not verify SES notification: %v", err)
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
	}

	messageTypeHeader := c.Request().Header.Get("X-Amz-Sns-Message-Type")
	if messageTypeHeader == "SubscriptionConfirmation" {
		if _, err := notificationPayload.Subscribe(); err != nil {
			app.log.Printf("error processing SNS (SES) subscription: %v", err)
			fmt.Println("Error when subscribing!", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
		return c.JSON(http.StatusOK, okResp{true})
	}
	if messageTypeHeader == "UnsubscribeConfirmation" {
		if _, err := notificationPayload.Unsubscribe(); err != nil {
			app.log.Printf("error processing SNS (SES) subscription: %v", err)
			fmt.Println("Error when subscribing!", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
		return c.JSON(http.StatusOK, okResp{true})
	}

	if err := json.Unmarshal([]byte(notificationPayload.Message), &m); err != nil {
		app.log.Printf("could not unmarshal SES notification: %v", err)
		app.log.Printf("notificationPayload.Message: %v", notificationPayload.Message)
		return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
	}

	sesType := getType(m)

	switch sesType {
	case "Delivery":
		event, err = recordDeliveryNotification(m)
		if err != nil {
			app.log.Printf("error processing SES notification: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
	case "Send":
		event, err = recordSendNotification(m)
		if err != nil {
			app.log.Printf("error processing SES notification: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
	case "Reject":
		event, err = recordRejectNotification(m)
		if err != nil {
			app.log.Printf("error processing SES notification: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
	case "Bounce":
		event, err = recordBounceNotification(m)
		if err != nil {
			app.log.Printf("error processing SES notification: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
	case "Complaint":
		event, err = recordComplaintNotification(m)
		if err != nil {
			app.log.Printf("error processing SES notification: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
	case "Open":
		event, err = recordOpenNotification(m)
		if err != nil {
			app.log.Printf("error processing SES notification: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
	case "Click":
		event, err = recordClickNotification(m)
		if err != nil {
			app.log.Printf("error processing SES notification: %v", err)
			return echo.NewHTTPError(http.StatusBadRequest, app.i18n.T("globals.messages.invalidData"))
		}
	default:
		app.log.Printf("unhandled SES event/notification type: %v", sesType)
		return c.JSON(http.StatusOK, okResp{true})
	}

	app.ensureEmailExists(m)
	app.setEmailStatus(m)

	_, err = app.queries.CreateEmailEvent.Exec(event.MessageID, event.Event, event.EventData, event.Timestamp)
	if err != nil {
		app.log.Printf("error recording email event: %v", err)
	}

	return c.JSON(http.StatusOK, okResp{true})
}

func (app *App) ensureEmailExists(m sesMail) {
	var emailCount int
	if err := app.queries.CountEmailsByMessageId.Get(&emailCount, m.Mail.MessageID); err != nil {
		app.log.Printf("error ensuring email: %v", err)
		return
	}
	if emailCount > 0 {
		return
	}
	e := models.Email{
		MessageID: m.Mail.MessageID,
		Recipient: m.Mail.Destination[0],
		Source:    m.Mail.Source,
		Subject:   m.Mail.CommonHeaders.Subject,
		Status:    "sent",
		SentAt:    m.Mail.Timestamp,
	}
	if _, err := app.queries.CreateEmail.Exec(nil, e.MessageID, e.Recipient, e.Source, e.Subject, e.Status, e.SentAt); err != nil {
		app.log.Printf("error saving email: %v", err)
	}
}

func (app *App) setEmailStatus(m sesMail) {
	var newStatus string
	switch getType(m) {
	case "Delivery":
		newStatus = "delivered"
	case "Reject":
		newStatus = "rejected"
	case "Bounce":
		newStatus = "bounced"
	case "Complaint":
		newStatus = "complained"
	default:
		return
	}
	if _, err := app.queries.UpdateEmail.Exec(newStatus); err != nil {
		app.log.Printf("error updating email status: %v", err)
	}
}

func recordDeliveryNotification(m sesMail) (models.EmailEvent, error) {
	var event models.EmailEvent
	jsonBytes, err := json.Marshal(m.Delivery)
	if err != nil {
		return event, fmt.Errorf("error encoding JSON: %v", err)
	}
	event = models.EmailEvent{
		MessageID: m.Mail.MessageID,
		Event:     "delivery",
		EventData: json.RawMessage(jsonBytes),
		Timestamp: time.Time(m.Delivery.Timestamp),
	}
	return event, nil
}

func recordRejectNotification(m sesMail) (models.EmailEvent, error) {
	var event models.EmailEvent
	jsonBytes, err := json.Marshal(m.Reject)
	if err != nil {
		return event, fmt.Errorf("error encoding JSON: %v", err)
	}
	event = models.EmailEvent{
		MessageID: m.Mail.MessageID,
		Event:     "reject",
		EventData: json.RawMessage(jsonBytes),
		Timestamp: time.Time(m.Mail.Timestamp),
	}
	return event, nil
}

func recordBounceNotification(m sesMail) (models.EmailEvent, error) {
	var event models.EmailEvent
	jsonBytes, err := json.Marshal(m.Bounce)
	if err != nil {
		return event, fmt.Errorf("error encoding JSON: %v", err)
	}
	event = models.EmailEvent{
		MessageID: m.Mail.MessageID,
		Event:     "bounce",
		EventData: json.RawMessage(jsonBytes),
		Timestamp: time.Time(m.Bounce.Timestamp),
	}
	return event, nil
}

func recordComplaintNotification(m sesMail) (models.EmailEvent, error) {
	var event models.EmailEvent
	jsonBytes, err := json.Marshal(m.Complaint)
	if err != nil {
		return event, fmt.Errorf("error encoding JSON: %v", err)
	}
	event = models.EmailEvent{
		MessageID: m.Mail.MessageID,
		Event:     "complaint",
		EventData: json.RawMessage(jsonBytes),
		Timestamp: time.Time(m.Complaint.Timestamp),
	}
	return event, nil
}

func recordSendNotification(m sesMail) (models.EmailEvent, error) {
	var event models.EmailEvent
	jsonBytes, err := json.Marshal(m.Send)
	if err != nil {
		return event, fmt.Errorf("error encoding JSON: %v", err)
	}
	event = models.EmailEvent{
		MessageID: m.Mail.MessageID,
		Event:     "send",
		EventData: json.RawMessage(jsonBytes),
		Timestamp: time.Time(m.Mail.Timestamp),
	}
	return event, nil
}

func recordOpenNotification(m sesMail) (models.EmailEvent, error) {
	var event models.EmailEvent
	jsonBytes, err := json.Marshal(m.Open)
	if err != nil {
		return event, fmt.Errorf("error encoding JSON: %v", err)
	}
	event = models.EmailEvent{
		MessageID: m.Mail.MessageID,
		Event:     "open",
		EventData: json.RawMessage(jsonBytes),
		Timestamp: time.Time(m.Open.Timestamp),
	}
	return event, nil
}

func recordClickNotification(m sesMail) (models.EmailEvent, error) {
	var event models.EmailEvent
	jsonBytes, err := json.Marshal(m.Click)
	if err != nil {
		return event, fmt.Errorf("error encoding JSON: %v", err)
	}
	event = models.EmailEvent{
		MessageID: m.Mail.MessageID,
		Event:     "click",
		EventData: json.RawMessage(jsonBytes),
		Timestamp: time.Time(m.Click.Timestamp),
	}
	return event, nil
}

func getType(m sesMail) string {
	if m.NotifType != "" {
		return m.NotifType
	}
	if m.EventType != "" {
		return m.EventType
	}
	return ""
}
