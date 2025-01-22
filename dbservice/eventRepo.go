package dbservice

import "github.com/scylladb/gocqlx/v2"

type EventRepository struct {
	session *gocqlx.Session
}


func NewEventRepository(session *gocqlx.Session) *EventRepository {
	return &EventRepository{
		session: session,
	}
}