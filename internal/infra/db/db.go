package db

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func migrateDB(defaultDb *gorm.DB) error {
	err := defaultDb.AutoMigrate(
		&load.MonitoringAgentInfo{},
		&load.LoadGeneratorServer{},
		&load.LoadGeneratorInstallInfo{},
		&load.LoadTestExecutionState{},
		&load.LoadTestExecutionInfo{},
		&load.LoadTestExecutionHttpInfo{},
	)

	if err != nil {
		log.Printf("[ERROR] Failed to auto migrate database tables: %v\n", err)
		return err
	}

	log.Println("[INFO] Database tables auto migration completed successfully")
	return nil
}

func connectSqliteDB(dbPath string) (*gorm.DB, error) {
	log.Printf("[INFO] SQLite configuration: meta SQLite DB path is %s\n", dbPath)

	newLogger := logger.New(
		log.New(os.Stdout, "\r", log.LstdFlags),
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
		log.Printf("[ERROR] Failed to connect to SQLite database: %v\n", err)
		return nil, err
	}

	log.Println("[INFO] Connected to SQLite database successfully")
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
	if d.Driver == "sqlite" || d.Driver == "sqlite3" {
		sqlFilePath := sqliteFilePath(d.Host)
		sqliteDB, err := connectSqliteDB(sqlFilePath)
		if err != nil {
			log.Fatalf("[ERROR] Failed to establish SQLite DB connection: %v\n", err)
		}

		err = migrateDB(sqliteDB)
		if err != nil {
			log.Fatalf("[ERROR] Failed to migrate SQLite database: %v\n", err)
		}
		db = sqliteDB
		log.Printf("[INFO] Initialized SQLite database successfully [%s]\n", d.Driver)
	} else if d.Driver == "postgres" {
		postgresDb, err := connectPostgresDB(d.Host, d.Port, d.User, d.Password, d.Name)
		if err != nil {
			log.Fatalf("[ERROR] Failed to establish Postgres DB connection: %v\n", err)
		}

		err = migrateDB(postgresDb)
		if err != nil {
			log.Fatalf("[ERROR] Failed to migrate Postgres database: %v\n", err)
		}
		db = postgresDb
		log.Printf("[INFO] Initialized Postgres database successfully [%s]\n", d.Driver)
	}

	dbConfig, _ := db.DB()
	dbConfig.SetMaxOpenConns(25)
	dbConfig.SetMaxIdleConns(125)
	dbConfig.SetConnMaxLifetime(time.Hour)

	return db, nil
}

func connectPostgresDB(host, port, user, password, name string) (*gorm.DB, error) {
	newLogger := logger.New(
		log.New(os.Stdout, "\r", log.LstdFlags),
		logger.Config{
			SlowThreshold: time.Second,
			// LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			ParameterizedQueries:      true,
			Colorful:                  true,
		},
	).LogMode(logger.Info)

	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, name)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("[ERROR] Failed to connect to Postgresql database: %v\n", err)
		return nil, err
	}

	log.Println("[INFO] Connected to Postgresql database successfully")
	return db, nil
}
