package configuration

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"log"
	"strings"
)

type Repo interface {
}

type SQLiteRepo struct {
	db *gorm.DB
}

func initDatabase() error {
	log.Println(">>>> start initDatabase()")
	ds := Get().datasource
	if ds.Driver == "sqlite" || ds.Driver == "sqlite3" {
		repo, err := NewSQLiteRepository()
		if err != nil {
			log.Fatal(err)
		}
		appConfig.DB = repo
	} else if ds.Driver == "mysql" {

	}
	log.Println(">>>> complete initDatabase()")

	return nil
}

func NewSQLiteRepository() (Repo, error) {
	connection := sqliteFilePath()
	log.Println(">>>> sqlite configuration; meta db path is", connection)

	db, err := gorm.Open(sqlite.Open(connection), &gorm.Config{})
	if err != nil {
		log.Println("NewSQLiteRepository() fail to connect to sqlite database")
		return nil, err
	}

	db.AutoMigrate()

	return &SQLiteRepo{
		db: db,
	}, nil
}

func sqliteFilePath() string {
	connection := Get().datasource.Connection
	if connection != "" {
		connection = strings.Replace(connection, "${ROOT}", RootPath(), 1)
	} else {
		connection = RootPath() + "/meta/ant_meta.db"
	}
	return connection
}
