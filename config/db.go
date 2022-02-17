package config

import (
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/aws/aws-sdk-go/service/ssm/ssmiface"
	"github.com/go-pg/pg/v9"
	"go_db/handlers"
	"log"
	"os"
)

type DBCreds struct {
	Hostname string `json:"hostname"`
	Database string `json:"database"`
	Username string `json:"username"`
	Password string `json:"password"`
}

func GetCredentials(ssmClient ssmiface.SSMAPI) (DBCreds, error) {
	output, err := ssmClient.GetParameter(&ssm.GetParameterInput{
		Name:           aws.String("/dev/timeline-api/dbcreds"),
		WithDecryption: aws.Bool(true),
	})
	if err != nil {
		log.Print(err)
		return DBCreds{}, err
	}

	var creds DBCreds

	if err := json.Unmarshal([]byte(*output.Parameter.Value), &creds); err != nil {
		log.Print(err)
		return DBCreds{}, err
	}

	return creds, nil
}

func Connect(creds DBCreds) *pg.DB {
	var db = pg.Connect(&pg.Options{
		User:     creds.Username,
		Password: creds.Password,
		Addr:     creds.Hostname + ":5432",
		Database: creds.Database,
	})

	if db == nil {
		log.Printf("Failed to connect")
		os.Exit(100)
	}
	log.Printf("Connected to db")
	handlers.CreateTimelineTable(db)
	handlers.InitiateDB(db)
	return db
}
