package migrations

import (
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/knadh/koanf/v2"
	"github.com/knadh/stuffbin"
)

// V5_0_0 performs the DB migrations.
func V5_0_1(db *sqlx.DB, fs stuffbin.FileSystem, ko *koanf.Koanf, lo *log.Logger) error {
	lo.Println("IMPORTANT: this upgrade might take a while if you have a large database. Please be patient ...")
	if _, err := db.Exec(`
		-- Add another option to campaign status to allow using campaigns for external automation.
		ALTER TYPE campaign_status ADD VALUE 'api';
	`); err != nil {
		return err
	}

	return nil
}
