package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V5_0_2 performs the DB migrations.
func V5_0_2(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {

	if _, err := db.Exec(`

		-- Add meta column to subscriptions.
		ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS subheading TEXT NULL;

	`); err != nil {
		return err
	}

	return nil
}
