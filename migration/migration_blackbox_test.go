package migration_test

import (
	"database/sql"
	"fmt"
	"os"
	"testing"

	"github.com/fabric8-services/fabric8-env/migration"

	migrationsupport "github.com/fabric8-services/fabric8-common/migration"

	"github.com/fabric8-services/fabric8-common/gormsupport"
	"github.com/fabric8-services/fabric8-common/resource"

	"github.com/jinzhu/gorm"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	dbName      = "test"
	defaultHost = "localhost"
	defaultPort = "5436"
)

type MigrationTestSuite struct {
	suite.Suite
}

const (
	databaseName = "test"
)

var (
	sqlDB *sql.DB
	host  string
	port  string
)

func TestMigration(t *testing.T) {
	suite.Run(t, new(MigrationTestSuite))
}

func (s *MigrationTestSuite) SetupTest() {
	resource.Require(s.T(), resource.Database)

	host = os.Getenv("F8_POSTGRES_HOST")
	if host == "" {
		host = defaultHost
	}
	port = os.Getenv("F8_POSTGRES_PORT")
	if port == "" {
		port = defaultPort
	}

	dbConfig := fmt.Sprintf("host=%s port=%s user=postgres password=mysecretpassword sslmode=disable connect_timeout=5", host, port)

	db, err := sql.Open("postgres", dbConfig)
	require.NoError(s.T(), err, "cannot connect to database: %s", dbName)
	defer db.Close()

	_, err = db.Exec("DROP DATABASE " + dbName)
	if err != nil && !gormsupport.IsInvalidCatalogName(err) {
		require.NoError(s.T(), err, "failed to drop database '%s'", dbName)
	}

	_, err = db.Exec("CREATE DATABASE " + dbName)
	require.NoError(s.T(), err, "failed to create database '%s'", dbName)
}

func (s *MigrationTestSuite) TestMigrate() {
	t := s.T()
	dbConfig := fmt.Sprintf("host=%s port=%s user=postgres password=mysecretpassword dbname=%s sslmode=disable connect_timeout=5",
		host, port, dbName)
	var err error
	sqlDB, err = sql.Open("postgres", dbConfig)
	require.NoError(t, err, "cannot connect to DB '%s'", dbName)
	defer sqlDB.Close()

	gormDB, err := gorm.Open("postgres", dbConfig)
	require.NoError(t, err, "cannot connect to DB '%s'", dbName)
	defer gormDB.Close()

	t.Run("checkMigration001", checkMigration001)
	t.Run("checkMigration002", checkMigration002)
}

func checkMigration001(t *testing.T) {
	err := migrationsupport.Migrate(sqlDB, databaseName, migration.Steps()[:2])
	require.NoError(t, err)

	t.Run("insert_ok", func(t *testing.T) {
		_, err := sqlDB.Exec(`INSERT INTO environments (id, name, type, space_id, namespace_name, cluster_url)
			VALUES (uuid_generate_v4(), 'osio-stage', 'stage', uuid_generate_v4(), '', 'cluster1.com')`)
		require.NoError(t, err)
	})
}

func checkMigration002(t *testing.T) {
	err := migrationsupport.Migrate(sqlDB, databaseName, migration.Steps()[:3])
	require.NoError(t, err)

	t.Run("insert_ok", func(t *testing.T) {
		_, err := sqlDB.Exec(`INSERT INTO environments (id, name, type, space_id, namespace_name, cluster_url)
			VALUES (uuid_generate_v4(), 'osio-stage', 'stage', uuid_generate_v4(), '', 'cluster1.com')`)
		require.NoError(t, err)
	})

	t.Run("insert_null_failed", func(t *testing.T) {
		_, err := sqlDB.Exec(`INSERT INTO environments (id, space_id, namespace_name, cluster_url)
			VALUES (uuid_generate_v4(), uuid_generate_v4(), '', 'cluster1.com')`)
		require.Error(t, err)
	})
}
