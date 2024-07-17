package database

import (
	"errors"
	"log"
	"os"
	"strings"
	"time"

	"github.com/cloud-barista/cm-ant/pkg/config"
	"github.com/cloud-barista/cm-ant/pkg/load/domain/model"
	"github.com/cloud-barista/cm-ant/pkg/utils"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	db *gorm.DB
)

func DB() *gorm.DB {
	return db
}

func InitDatabase() error {
	log.Println(">>>> start database initialize")
	ds := config.AppConfig.Database

	if ds.Driver == "sqlite" || ds.Driver == "sqlite3" {
		sqlFilePath := sqliteFilePath(ds.Host)
		sqliteDB, err := connectSqliteDB(sqlFilePath)
		if err != nil {
			log.Fatal(err)
		}

		err = migrateDB(sqliteDB)
		if err != nil {
			log.Fatal(err)
		}
		db = sqliteDB
		log.Printf(">>>> complete initialize database [%s] \n", ds.Driver)
		return nil
	}

	return errors.New("no matching database driver")
}

func migrateDB(defaultDb *gorm.DB) error {
	err := defaultDb.AutoMigrate(
		&model.LoadEnv{},
		&model.LoadExecutionState{},
	)

	if err != nil {
		log.Println("migrateDB() fail to connect to sqlite database")
		return err
	}

	return nil
}

func connectSqliteDB(dbPath string) (*gorm.DB, error) {
	log.Println(">>>> sqlite configuration; meta sqliteDb path is", dbPath)

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
		log.Println("connectSqliteDB() fail to connect to sqlite database")
		return nil, err
	}

	return sqliteDb, nil
}

func sqliteFilePath(host string) string {
	dbFile := host
	rp := utils.RootPath()

	if dbFile != "" {
		dbFile = strings.Replace(dbFile, "${ROOT}", rp, 1)
	} else {
		dbFile = rp + "/meta/ant_meta.db"
	}
	return dbFile
}
