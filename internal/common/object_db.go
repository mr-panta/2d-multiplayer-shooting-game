package common

import (
	"github.com/mr-panta/2d-multiplayer-shooting-game/internal/util"
)

type ObjectDB interface {
	SelectOne(id string) (o Object, exists bool)
	SelectAll() []Object
	Set(p Object)
	Delete(id string)
}

type objectDB struct {
	inMemDB util.InMemDB
}

func NewObjectDB() ObjectDB {
	return &objectDB{inMemDB: util.NewInMemDB()}
}

func (db *objectDB) SelectOne(id string) (p Object, exists bool) {
	if row, exists := db.inMemDB.SelectOne(id); exists {
		return row.(Object), true
	}
	return nil, false
}

func (db *objectDB) SelectAll() (players []Object) {
	for _, row := range db.inMemDB.SelectAll() {
		players = append(players, row.(Object))
	}
	return players
}

func (db *objectDB) Set(p Object) {
	db.inMemDB.Set(p.GetID(), p)
}

func (db *objectDB) Delete(id string) {
	db.inMemDB.Delete(id)
}
