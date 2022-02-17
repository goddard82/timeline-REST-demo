package main

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/gin-gonic/gin"
	"go_db/config"
	"go_db/routes"
	"log"
)

func main() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable, // Must be set to enable
		Profile:           "default",
	})) // local

	//sess := session.Must(session.NewSession()) // prod

	ssmClient := ssm.New(sess, &aws.Config{
		Region: aws.String("eu-west-1"),
	})

	// Get DB creds
	creds, err := config.GetCredentials(ssmClient)
	if err != nil {
		panic(err)
	}

	fmt.Println(creds)

	// Connect DB
	config.Connect(creds)

	// Init Router
	router := gin.Default()

	// Route Handlers / Endpoints
	routes.Routes(router)

	log.Fatal(router.Run(":4747"))
}
