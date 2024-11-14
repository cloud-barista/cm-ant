package db

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/cloud-barista/cm-ant/internal/utils"
	"github.com/rs/zerolog/log"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type zerologGormLogger struct{}

func (z zerologGormLogger) Printf(format string, v ...interface{}) {
    log.Printf((fmt.Sprintf(format, v...)))
}


func migrateDB(defaultDb *gorm.DB) error {
	err := defaultDb.AutoMigrate(
		&load.MonitoringAgentInfo{},
		&load.LoadGeneratorServer{},
		&load.LoadGeneratorInstallInfo{},
		&load.LoadTestExecutionInfo{},
		&load.LoadTestExecutionHttpInfo{},
		&load.LoadTestExecutionState{},

		&cost.EstimateCostInfo{},
		&cost.EstimateForecastCostInfo{},
	)

	if err != nil {
		log.Error().Msgf("Failed to auto migrate database tables: %v", err)
		return err
	}

	log.Info().Msg("Database tables auto migration completed successfully")
	return nil
}

func connectSqliteDB(dbPath string) (*gorm.DB, error) {
	log.Info().Msgf("SQLite configuration: meta SQLite DB path is %s", dbPath)

	newLogger := logger.New(
		zerologGormLogger{},
		logger.Config{
			SlowThreshold: time.Second,
			// LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)

	sqliteDb, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Error().Msgf("Failed to connect to SQLite database: %v", err)
		return nil, err
	}

	log.Info().Msg("Connected to SQLite database successfully")
	return sqliteDb, nil
}

func sqliteFilePath(host string) string {
	dbFilePath := host
	rp := utils.RootPath()

	if dbFilePath != "" {
		dbFilePath = strings.Replace(dbFilePath, "${ROOT}", rp, 1)
	} else {
		dbFilePath = rp + "/meta/ant_meta.db"
	}
	return dbFilePath
}

func NewDBConnection() (*gorm.DB, error) {
	d := config.AppConfig.Database

	var db *gorm.DB
	driver := strings.ToLower(d.Driver)
	// if driver == "sqlite" || driver == "sqlite3" {
	// 	sqlFilePath := sqliteFilePath(d.Host)
	// 	sqliteDB, err := connectSqliteDB(sqlFilePath)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	db = sqliteDB
	// } else
	if driver == "postgres" {
		postgresDb, err := connectPostgresDB(d.Host, d.Port, d.User, d.Password, d.Name)
		if err != nil {
			log.Error().Msgf("Failed to establish Postgres DB connection: %v", err)
			return nil, err
		}

		db = postgresDb
		log.Info().Msgf("Initialized Postgres database successfully [%s]", d.Driver)
	} else {
		return nil, errors.New("unsuppored database driver")
	}

	err := migrateDB(db)
	if err != nil {
		log.Info().Msgf("Failed to migrate database: %v", err)
		return nil, err
	}

	dbConfig, _ := db.DB()
	dbConfig.SetMaxOpenConns(25)
	dbConfig.SetMaxIdleConns(125)
	dbConfig.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func connectPostgresDB(host, port, user, password, name string) (*gorm.DB, error) {
	if host == "" || port == "" || user == "" || password == "" || name == "" {
		return nil, errors.New("database connection info is incorrect")
	}
	newLogger := logger.New(
		zerologGormLogger{},
		logger.Config{
			SlowThreshold: time.Second,
			// LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Info().Msgf("Failed to connect to Postgresql database: %v", err)
		return nil, err
	}

	log.Info().Msg("Connected to Postgresql database successfully")
	return db, nil
}
