package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"go_db/handlers"
)

func Routes(router *gin.Engine) {
	router.GET("/health", health)                                                               // returns a 200 if the service is running
	router.GET("/timeline/events", basicAuth, handlers.GetAllEvents)                            // returns a list of all events
	router.GET("/timeline/events/recent", basicAuth, handlers.GetRecentEvents)                  // returns all events that started or were entered in last 30 minutes
	router.GET("/timeline/events/daterange", basicAuth, handlers.GetDateRange)                  // takes a querystring
	router.GET("/timeline/events/daterange/advanced", basicAuth, handlers.GetDateRangeAdvanced) // takes a querystring
	router.GET("/timeline/events/query", basicAuth, handlers.GetQuery)                          // takes a querystring
	router.GET("/timeline/events/", basicAuth, handlers.GetQueryLastX)                          // takes a querystring
	router.POST("/timeline/events", basicAuth, handlers.CreateEvent)                            // creates event from submitted json
	router.GET("/timeline/events/:eventId", basicAuth, handlers.GetEvent)                       // gets event by it's UUID from url
	router.PUT("/timeline/events/:eventId", basicAuth, handlers.EditEvent)                      // gets event by it's UUID from url & updates it
	router.NoRoute(notFound)
}

func health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  http.StatusOK,
		"message": "Healthy",
	})
	return
}

func basicAuth(c *gin.Context) {
	// Get the Basic Authentication credentials
	user, password, hasAuth := c.Request.BasicAuth()
	if hasAuth == false || user != "timeline" || password != "p@S$wOrd" {
		c.JSON(http.StatusUnauthorized, gin.H{})
		c.Abort()
		return
	}
}

func notFound(c *gin.Context) {
	c.JSON(http.StatusNotFound, gin.H{
		"status":  http.StatusNotFound,
		"message": "Route Not Found",
	})
	return
}
