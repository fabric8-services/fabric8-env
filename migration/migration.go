package migration

import "database/sql"
import "github.com/fabric8-services/fabric8-common/migration"

type migrateData struct {
}

func Migrate(db *sql.DB, catalog string) error {
	return migration.Migrate(db, catalog, migrateData{})
}

func (d migrateData) Asset(name string) ([]byte, error) {
	return Asset(name)
}

func (d migrateData) AssetNameWithArgs() [][]string {
	names := [][]string{
		{"000-bootstrap.sql"},
		{"001-environments.sql"},
	}
	return names
}
