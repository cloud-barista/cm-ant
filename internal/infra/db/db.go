package db

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/internal/config"
	"github.com/cloud-barista/cm-ant/internal/core/cost"
	"github.com/cloud-barista/cm-ant/internal/core/load"
	"github.com/cloud-barista/cm-ant/internal/utils"
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
		&load.LoadTestExecutionInfo{},
		&load.LoadTestExecutionHttpInfo{},
		&load.LoadTestExecutionState{},

		&cost.PriceInfo{},
	)

	if err != nil {
		utils.LogErrorf("Failed to auto migrate database tables: %v\n", err)
		return err
	}

	utils.LogInfo("Database tables auto migration completed successfully")
	return nil
}

func connectSqliteDB(dbPath string) (*gorm.DB, error) {
	utils.LogInfof("SQLite configuration: meta SQLite DB path is %s\n", dbPath)

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
	driver := strings.ToLower(d.Driver)
	// if driver == "sqlite" || driver == "sqlite3" {
	// 	sqlFilePath := sqliteFilePath(d.Host)
	// 	sqliteDB, err := connectSqliteDB(sqlFilePath)
	// 	if err != nil {
	// 		utils.LogErrorf("Failed to establish SQLite DB connection: %v\n", err)
	// 		return nil, err
	// 	}

	// 	db = sqliteDB
	// 	utils.LogInfof("Initialized SQLite database successfully [%s]\n", d.Driver)
	// } else
	if driver == "postgres" {
		postgresDb, err := connectPostgresDB(d.Host, d.Port, d.User, d.Password, d.Name)
		if err != nil {
			utils.LogErrorf("Failed to establish Postgres DB connection: %v\n", err)
			return nil, err
		}

		db = postgresDb
		utils.LogInfof("Initialized Postgres database successfully [%s]\n", d.Driver)
	} else {
		return nil, errors.New("unsuppored database driver")
	}

	err := migrateDB(db)
	if err != nil {
		utils.LogErrorf("Failed to migrate database: %v\n", err)
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
		log.New(os.Stdout, "\r", log.LstdFlags),
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
		utils.LogErrorf("Failed to connect to Postgresql database: %v\n", err)
		return nil, err
	}

	utils.LogInfo("Connected to Postgresql database successfully")
	return db, nil
}
