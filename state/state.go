package state

import (
	"context"
	"encoding/json"
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

type EventEnvelope struct {
	ID    string          `json:"id"`
	Event json.RawMessage `json:"event"`
}

func (e *EventEnvelope) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`{id: "%s",event: %s}`, e.ID, string(e.Event))), nil
}

const eventBufferSize = 40

func (s *State) StreamEvents(ctx context.Context, lastSeen string) func() (*EventEnvelope, error) {

	obs, cancel := s.db.Observe(eventsPath.ToMatcher().AppendAnyElementMatcher())

	envelopesChan := make(chan *EventEnvelope, eventBufferSize)

	readNextEvents := func(lastSeen string) (envelopes []*EventEnvelope, err error) {
		err = bolted.SugaredRead(s.db, func(tx bolted.SugaredReadTx) error {
			it := tx.Iterator(eventsPath)
			it.Seek(lastSeen)
			if it.GetKey() == lastSeen {
				it.Next()
			}

			for ; !it.IsDone(); it.Next() {
				envelopes = append(envelopes, &EventEnvelope{
					ID:    it.GetKey(),
					Event: it.GetValue(),
				})
			}
			return nil
		})
		if err != nil {
			return nil, err
		}

		return envelopes, nil

	}

	go func() {
		defer cancel()
		defer close(envelopesChan)

		for ctx.Err() == nil {

			select {
			case <-ctx.Done():
				return
			case <-obs:
				for ctx.Err() == nil {

					nextEvents, err := readNextEvents(lastSeen)
					if err != nil {
						s.log.Error(err, "could not read next events")
						return
					}
					if len(nextEvents) == 0 {
						break
					}

					for _, e := range nextEvents {

						envelopesChan <- e
						if ctx.Err() != nil {
							return
						}

						lastSeen = e.ID
					}

				}

			}
			// events :=
		}
	}()

	return func() (*EventEnvelope, error) {
		env, ok := <-envelopesChan
		if !ok {
			return nil, fmt.Errorf("closed")
		}

		return env, nil
	}
}
