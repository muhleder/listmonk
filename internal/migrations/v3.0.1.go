package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V3_0_1 performs the DB migrations.
func V3_0_1(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {

	if _, err := db.Exec(`

		-- Add meta column to subscriptions.
		ALTER TABLE campaigns ADD COLUMN IF NOT EXISTS preview_text TEXT NULL;

	`); err != nil {
		return err
	}

	return nil
}
