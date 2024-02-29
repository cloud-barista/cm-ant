package domain

import (
	"fmt"
	"log"
)

var DBMap map[string]TempRepository

func InitializeDatabase() {
	DBMap = map[string]TempRepository{
		"agent": NewRepository(),
	}

	log.Println("[CM-ANT] Database initialized")
}

type TempRepository interface {
	FindById(string) (interface{}, error)
	Insert(string, interface{}) error
	DeleteById(string)
}

type LocalRepository struct {
	db map[string]interface{}
}

func NewRepository() TempRepository {
	return &LocalRepository{
		db: map[string]interface{}{},
	}
}

func (lr *LocalRepository) FindById(id string) (interface{}, error) {
	if item, ok := lr.db[id]; ok {
		return item, nil
	}

	return nil, nil
}

func (lr *LocalRepository) Insert(id string, item interface{}) error {
	if _, ok := lr.db[id]; !ok {
		lr.db[id] = item
		fmt.Println("item inserted")
		return nil
	}

	return fmt.Errorf("%s already exist;", id)
}

func (lr *LocalRepository) DeleteById(id string) {
	delete(lr.db, id)
}
