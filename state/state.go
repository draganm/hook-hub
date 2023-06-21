package state

import (
	"fmt"

	"github.com/draganm/bolted"
	"github.com/draganm/bolted/dbpath"
	"github.com/go-logr/logr"
	"github.com/gofrs/uuid"
)

var eventsPath = dbpath.ToPath("events")

type State struct {
	db  bolted.Database
	log logr.Logger
}

func New(db bolted.Database, log logr.Logger) (*State, error) {
	err := bolted.SugaredWrite(db, func(tx bolted.SugaredWriteTx) error {
		if !tx.Exists(eventsPath) {
			tx.CreateMap(eventsPath)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("could not initialize state: %w", err)
	}

	return &State{db: db, log: log}, nil

}

func (s *State) StoreEvent(data string) (err error) {
	id, err := uuid.NewV6()
	defer func() {
		if err == nil {
			s.log.Info("new event stored", "id", id.String())
		}
	}()
	if err != nil {
		return fmt.Errorf("could not create uuid: %w", err)
	}
	return bolted.SugaredWrite(s.db, func(tx bolted.SugaredWriteTx) error {
		tx.Put(eventsPath.Append(id.String()), []byte(data))
		return nil
	})
}
