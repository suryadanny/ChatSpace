package dbservice

import (
	"dev/chatspace/models"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/scylladb/gocqlx/table"
	"github.com/scylladb/gocqlx/v2"
	"github.com/scylladb/gocqlx/v2/qb"
)

// EventMetadata is the metadata for the event table
var EventMetadata = table.Metadata{
	Name: "chat_store",
	Columns: []string{"receiver_id", "sender_id", "delivered", "received", "message", "event_id", "is_delivered"},
	PartKey: []string{"event_id","sender_id"},
	SortKey: []string{"rceived"},
}

// EventTable stuct used for binding the event table
var EventTable = table.New(EventMetadata)


//using scylladb/gocqlx/v2 for cassandra operations, thats helps us get rid of the boilerplate code
type EventRepository struct {
	session *gocqlx.Session
}




func NewEventRepository(session *gocqlx.Session) *EventRepository {
	return &EventRepository{
		session: session,
	}
}

// AddEvent adds an event to the cassandra store
func (e *EventRepository) AddEvent(event *models.Event) error {
	insertQuery := e.session.Query(EventTable.Insert())

	if err := insertQuery.BindStruct(event).ExecRelease(); err != nil {	
		log.Println("error while inserting event : ", err)
		return err
	}	

	return nil

}

//updating the event in the cassandra store
func (e *EventRepository) UpdateEvent(event map[string]interface{}, event_id gocql.UUID, sender_id string, received time.Time) error {
	updateQuery := qb.Update("store.chat_store")
	qb_map := qb.M{}

	for key, value := range event {	
		updateQuery.Set(key)
		qb_map[key] = value
	}


	updateQuery.Where(qb.Eq("event_id"))
	qb_map["event_id"] = event_id
	updateQuery.Where(qb.Eq("sender_id"))
	qb_map["sender_id"] = sender_id
	updateQuery.Where(qb.Eq("received"))
	qb_map["received"] = received
	stmt, names := updateQuery.ToCql()

	if err := e.session.Query(stmt, names).BindMap(qb_map).ExecRelease(); err != nil {
		log.Println("error while updating event : ", err)
		return err
	}

	return nil
}