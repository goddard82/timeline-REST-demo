package handlers

import (
	"fmt"
	"github.com/go-pg/pg/v9"
	guuid "github.com/google/uuid"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v9/orm"
)

type Event struct {
	InstanceID  string      `json:"instance_id"`
	Description string      `json:"description"`
	Type        string      `json:"type,required" binding:"required"`
	Subsystem   string      `json:"subsystem" binding:"required"`
	Status      string      `json:"status,required" binding:"required"`
	Payload     interface{} `json:"payload"`
	Timestamp   string      `json:"timestamp"`
	Endtime     string      `json:"endtime,omitempty"`
	Suppressed  bool        `json:"suppressed,omitempty"`
	tableName   struct{}    `pg:"timeline"`
}

// todo add a struct for events _from_ the DB

type IngestionEvent struct {
	Event // not needed yet
	//InstanceID  string      `json:"instance_id"`
	//Description string      `json:"description"`
	//EventType   string      `json:"type"`
	//Subsystem   string      `json:"subsystem"`
	//Status      string      `json:"status"`
	//Payload     interface{} `json:"payload"`
	//Timestamp   time.Time   `json:"timestamp"`
	//Endtime     time.Time   `json:"endtime"`
}

type CreatedEvent struct {
	InstanceID  string      `json:"instance_id"`
	Description string      `json:"description,omitempty"`
	EventType   string      `json:"type"`
	Subsystem   string      `json:"subsystem"`
	Status      string      `json:"status,omitempty"`
	Payload     interface{} `json:"payload,omitempty"`
	Timestamp   string      `json:"timestamp"` // changed time.Time to string, might have to change back
	Endtime     string      `json:"endtime,omitempty"`
}

func CreateTimelineTable(db *pg.DB) error {
	opts := &orm.CreateTableOptions{
		IfNotExists: true,
	}
	createError := db.CreateTable(&Event{}, opts)
	if createError != nil {
		log.Printf("Error while creating event table, Reason: %v\n", createError)
		return createError
	}
	log.Printf("Event table created")
	return nil
}

// INITIALISE DB CONNECTION
var dbConnect *pg.DB

func InitiateDB(db *pg.DB) {
	dbConnect = db
	if db != nil {
		log.Printf("running...", db)
	}
}

// GetAllEvents
// done
func GetAllEvents(c *gin.Context) {
	var events []Event
	upto := c.Query("upto")
	limitstring := c.Query("limit")
	offsetstring := c.Query("offset")
	limit, _ := strconv.Atoi(limitstring)
	offset, _ := strconv.Atoi(offsetstring)
	if upto == "" {
		err := dbConnect.Model(&events).
			Order("timestamp desc").
			Limit(limit).
			Offset(offset).
			Select()
		if err != nil {
			log.Printf("Error while getting all events, Reason: %v\n", err)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Something went wrong",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"events": events,
		})
	}
	if upto == "now" {
		errt := dbConnect.Model(&events).
			Where("timestamp < ?()", upto).
			Order("timestamp desc").
			Limit(limit).
			Offset(offset).
			Select()
		if errt != nil {
			log.Printf("Error while getting all events, Reason: %v\n", errt)
			c.JSON(http.StatusInternalServerError, gin.H{
				"status":  http.StatusInternalServerError,
				"message": "Something went wrong",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"events": events,
		})
		return
	}
}

// GetRecentEvents
// done
func GetRecentEvents(c *gin.Context) {
	var events []Event
	err := dbConnect.Model(&events).
		Where("timestamp BETWEEN now()::timestamp - (interval '30m') AND now()::timestamp").
		Order("timestamp desc").
		Select()

	if err != nil {
		log.Printf("Error while getting recent events, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
			"query":   err,
		})
		return
	}
	if len(events) == 0 {
		var nowt []string
		c.JSON(http.StatusOK, gin.H{
			"events": nowt,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
	return
}

// GetDateRange
// done
func GetDateRange(c *gin.Context) {
	var events []Event
	start := c.Query("start")
	end := c.Query("end")
	err := dbConnect.Model(&events).
		Where("timestamp between ? and ?", start, end).
		Order("timestamp desc").
		Select()
	if err != nil {
		log.Printf("Error while getting events in daterange, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	//log.Printf(err)
	if len(events) == 0 {
		var nowt []string
		c.JSON(http.StatusOK, gin.H{
			"events": nowt,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
	return
}

// GetDateRangeAdvanced
// done
func GetDateRangeAdvanced(c *gin.Context) {
	var events []Event
	start := c.Query("start") + " 00:00:00.000"
	end := c.Query("end") + " 23:59:59.999"
	err := dbConnect.Model(&events).
		Where("(case when endtime isnull then timestamp between ? and ? else (tsrange(timestamp, endtime) && tsrange(?, ?)) end)", start, end, start, end).
		Order("timestamp asc").
		Select()
	if err != nil {
		log.Printf("Error while getting all events, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	if len(events) == 0 {
		var nowt []string
		c.JSON(http.StatusOK, gin.H{
			"events": nowt,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
	return
}

// (not important rn) rename this directory Handlers (done), and move the functions that actually do the querying to another dir

// GetQuery
// done
func GetQuery(c *gin.Context) {
	var events []Event
	description := c.Query("description")
	status := c.Query("status")
	subsystem := c.Query("subsystem")
	eventType := c.Query("type")
	//timestamp := c.Query("timestamp")
	//endtime := c.Query("endtime")
	payload := c.Query("payload")
	//payloadType := fmt.Sprintf("%T", payload)
	//fmt.Println(payloadType)
	query := dbConnect.Model(&events)
	if description != "" {
		query.Where("description ILIKE ?", description)
	}
	if subsystem != "" {
		query.Where("subsystem = ?", subsystem)
	}
	if status != "" {
		query.Where("status = ?", status)
	}
	if eventType != "" {
		query.Where("event_type = ?", eventType)
	}
	if payload != "" {
		query.Where("? = ANY (SELECT (jsonb_each_text(payload)).value) or ? = ANY (SELECT (jsonb_each_text(payload)).key)", payload, payload)
	}
	query.Order("timestamp desc")
	queryError := query.Select()
	if queryError != nil {
		log.Printf("Error while getting events, Reason: %v\n", queryError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	if len(events) == 0 {
		log.Printf("query: %v\n", query.Select())
		var nowt []string
		c.JSON(http.StatusOK, gin.H{
			"events": nowt,
		})
		return
	}
	// todo log request/URL
	log.Printf("query: %v\n", query.Select())
	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
	return

}

//GetQueryLastX
// done
func GetQueryLastX(c *gin.Context) {
	var events []Event
	lastString := c.Query("last")
	last, _ := strconv.Atoi(lastString)
	err := dbConnect.Model(&events).
		Where("timestamp <= now()::timestamp").
		Order("timestamp desc").
		Limit(last).
		Select()
	if err != nil {
		log.Printf("Error while getting events, Reason: %v\n", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"events": events,
	})
	return
}

// CreateEvent
// need to add additional event type for data platform ingestion service (not urgent, not yet implemented at data side)
// https://github.com/lkwd/timeline_api_go/blob/a81c4e07088679b26ff7d359ff43592a580251f2/internal/handlers/create/handler.go#L99
// the ingestion service event will need to then be sent in a POST request to an external api, using an auth token that'll need to be secure
func CreateEvent(c *gin.Context) {
	var event Event
	//timestamp := c.Query("timestamp")
	err := c.BindJSON(&event)
	if err != nil {
		fmt.Println(err)
		c.AbortWithError(400, err)
	}
	subsystem := event.Subsystem
	description := event.Description
	eventType := event.Type
	instanceid := guuid.New().String()
	payload := event.Payload
	endtime := event.Endtime
	status := event.Status
	timestamp := event.Timestamp
	timespamp := time.Now().Format(time.RFC1123Z)
	if timestamp == "" {
		timestamp = timespamp
	}
	fmt.Println(instanceid, description, subsystem, eventType)
	insertError := dbConnect.Insert(&Event{
		InstanceID:  instanceid,
		Subsystem:   subsystem,
		Description: description,
		Type:        eventType,
		Payload:     payload,
		Status:      status,
		Timestamp:   timestamp,
		Endtime:     endtime,
	})
	if insertError != nil {
		log.Printf("Error while inserting new event into db, Reason: %v\n", insertError)
		c.JSON(http.StatusInternalServerError, gin.H{
			"status":  http.StatusInternalServerError,
			"message": "Something went wrong",
		})
		return
	}
	// todo at this point, send the event (with some extra data) to the ingestion service api
	c.JSON(http.StatusCreated, gin.H{
		"event": &CreatedEvent{
			InstanceID:  instanceid,
			Subsystem:   subsystem,
			Description: description,
			EventType:   eventType,
			Payload:     payload,
			Status:      status,
			Timestamp:   timestamp,
			Endtime:     endtime,
		},
	})
	return
}

//GetEvent
// done
func GetEvent(c *gin.Context) {
	var events []Event
	instanceId := c.Param("eventId")
	err := dbConnect.Model(&events).
		Where("instance_id = ?", instanceId).
		Select()
	if err != nil {
		log.Printf("Error while getting a single event, Reason: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":     http.StatusNotFound,
			"message":    "Event not found",
			"instanceId": instanceId,
		})
		return
	}
	if len(events) == 0 {
		c.Status(http.StatusNoContent)
		return
	}
	c.JSON(http.StatusOK, gin.H{"events": events})
	return
}

//EditEvent - this resolves the status to OK and enters the current timestamp as endtime. Is intended to be used to mark
// ongoing events (alerts) as complete.
// done
func EditEvent(c *gin.Context) {
	instanceId := c.Param("eventId")
	status := "ok"
	endtime := time.Now()
	var e []Event
	_, err := dbConnect.Model(&e).
		Set("endtime = ?", endtime).
		Set("status = ?", status).
		Where("instance_id = ?", instanceId).
		Returning("*").
		Update()

	if err != nil {
		log.Printf("Error, Reason: %v\n", err)
		c.JSON(http.StatusNotFound, gin.H{
			"status":  500,
			"message": err,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": e,
	})
	return
}
