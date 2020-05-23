package util

import (
	"sync"
)

type InMemDB interface {
	SelectOne(key string) (value interface{}, exists bool)
	SelectAll() []interface{}
	Set(key string, value interface{})
	Delete(key string)
}

type inMemDB struct {
	rows      []interface{}
	indexMap  map[string]int
	indexLock sync.RWMutex
}

func NewInMemDB() InMemDB {
	return &inMemDB{
		indexMap: make(map[string]int),
	}
}

func (db *inMemDB) SelectOne(key string) (value interface{}, exists bool) {
	db.indexLock.RLock()
	defer db.indexLock.RUnlock()
	if index, exists := db.indexMap[key]; exists {
		return db.rows[index], true
	}
	return nil, false
}

func (db *inMemDB) SelectAll() []interface{} {
	return db.rows
}

func (db *inMemDB) Set(key string, value interface{}) {
	db.indexLock.Lock()
	defer db.indexLock.Unlock()
	if index, exists := db.indexMap[key]; exists {
		db.rows[index] = value
	} else {
		db.indexMap[key] = len(db.rows)
		db.rows = append(db.rows, value)
	}
}

func (db *inMemDB) Delete(key string) {
	db.indexLock.Lock()
	defer db.indexLock.Unlock()
	if index, exists := db.indexMap[key]; exists {
		delete(db.indexMap, key)
		rows := db.rows[:index]
		if index != len(db.rows)-1 {
			rows = append(rows, db.rows[index+1:]...)
		}
		db.rows = rows
		for key, i := range db.indexMap {
			if i > index {
				db.indexMap[key]--
			}
		}
	}
}
