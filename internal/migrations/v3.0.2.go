package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V3_0_2 performs the DB migrations.
func V3_0_2(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {

	if _, err := db.Exec(`

CREATE TABLE IF NOT EXISTS emails (
    id serial PRIMARY KEY,
    campaign_id int,
    message_id character varying(255),
    recipient character varying(255) NOT NULL,
    source character varying(255) NOT NULL,
    subject character varying(255) DEFAULT NULL::character varying,
    status character varying(255) NOT NULL,
    sent_at timestamp(0) without time zone NOT NULL
);

-- Indices -------------------------------------------------------

CREATE UNIQUE INDEX IF NOT EXISTS emails_pkey ON emails(id int4_ops);
CREATE INDEX IF NOT EXISTS emails_campaign_id_index ON emails(campaign_id int4_ops);
CREATE INDEX IF NOT EXISTS emails_recipient_index ON emails(recipient text_ops);
CREATE INDEX IF NOT EXISTS emails_message_id_index ON emails(message_id text_ops);

CREATE TABLE IF NOT EXISTS email_events (
    id serial PRIMARY KEY,
    message_id character varying(255) NOT NULL,
    event character varying(255) NOT NULL,
    event_data json,
    timestamp timestamp(0) without time zone
);

-- Indices -------------------------------------------------------

CREATE UNIQUE INDEX IF NOT EXISTS email_events_pkey ON email_events(id int4_ops);
CREATE INDEX IF NOT EXISTS email_events_message_id_index ON email_events(message_id text_ops);

	`); err != nil {
		return err
	}

	return nil
}
