/* The /analytics group contains all the routes to track analytics */
package routes

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"timeful/server/accounts"
	"timeful/server/logger"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
	"timeful/server/slackbot"
)

// BasicAuth middleware for analytics routes
func AnalyticsBasicAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		analyticsUsername := os.Getenv("ANALYTICS_USERNAME")
		analyticsPassword := os.Getenv("ANALYTICS_PASSWORD")
		user, pass, hasAuth := c.Request.BasicAuth()

		if !hasAuth || user != analyticsUsername || pass != analyticsPassword {
			c.Header("WWW-Authenticate", `Basic realm="Restricted"`)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
			return
		}
		c.Next()
	}
}

func InitAnalytics(router *gin.RouterGroup) {
	analyticsRouter := router.Group("/analytics")

	analyticsRouter.POST("/scanned-poster", scannedPoster)
	analyticsRouter.GET("/monthly-active-event-creators", AnalyticsBasicAuth(), getMonthlyActiveEventCreators)
	analyticsRouter.GET("/monthly-active-event-creators-with-more-than-x-events", AnalyticsBasicAuth(), getMonthlyActiveEventCreatorsWithMoreThanXEvents)
	analyticsRouter.GET("/user/:email", AnalyticsBasicAuth(), getUserByEmail)
}

// @Summary Notifies us when poster QR code has been scanned
// @Tags analytics
// @Accept json
// @Produce json
// @Param payload body object{url=string,location=models.Location} true "Object containing the location that poster was scanned from and the url that was scanned"
// @Success 200
// @Router /analytics/scanned-poster [post]
func scannedPoster(c *gin.Context) {
	payload := struct {
		Url      string           `json:"url" binding:"required"`
		Location *models.Location `json:"location"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	if payload.Location != nil {
		slackbot.SendTextMessage(
			fmt.Sprintf(":face_with_monocle: Poster was scanned :face_with_monocle:\n*Location:* %s, %s, %s\n*URL:* %s",
				payload.Location.City,
				payload.Location.State,
				payload.Location.CountryCode,
				payload.Url,
			),
		)
	} else {
		slackbot.SendTextMessage(
			fmt.Sprintf(":face_with_monocle: Poster was scanned :face_with_monocle:\n*URL:* %s", payload.Url),
		)
	}

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Gets the daily count of monthly active event creators over a date range
// @Tags analytics
// @Accept json
// @Produce json
// @Param startDate query string true "Start date (YYYY-MM-DD) for the range"
// @Param endDate query string true "End date (YYYY-MM-DD) for the range"
// @Param timezoneOffset query integer true "Client's timezone offset in minutes from UTC (e.g., -420 for UTC-7)"
// @Success 200 {array} object{date=string,count=int}
// @Failure 400 {object} object{error=string} "Invalid date format, range, or timezone offset"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /analytics/monthly-active-event-creators [get]
func getMonthlyActiveEventCreators(c *gin.Context) {
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	timezoneOffsetStr := c.Query("timezoneOffset") // Get timezone offset param

	if startDateStr == "" || endDateStr == "" || timezoneOffsetStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate, endDate, and timezoneOffset query parameters are required"})
		return
	}

	// Parse timezone offset
	timezoneOffset, err := strconv.Atoi(timezoneOffsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timezoneOffset format. Must be an integer representing minutes."})
		return
	}
	// Convert offset minutes to a location (Go's time package uses seconds west of UTC)
	// Note: JS getTimezoneOffset() is positive for west, negative for east.
	// Go FixedZone expects seconds east of UTC. So, offset needs to be negated and converted to seconds.
	location := time.FixedZone("UserOffset", -timezoneOffset*60)

	layout := "2006-01-02" // YYYY-MM-DD
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startDate format. Use YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endDate format. Use YYYY-MM-DD"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate cannot be before startDate"})
		return
	}

	// Creator analytics read authoritative event storage. Because
	// migrated and new events live in one store, each event contributes once
	// and no creator is counted twice across stores.
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		logger.StdErr.Panicln(err)
	}

	// One day-spine query returns the whole range in ascending day order. Unlike
	// the retired per-day loop, a query failure cannot yield a short partial
	// array, so the request fails whole instead of silently dropping days.
	results, err := repository.CountDistinctMonthlyActiveEventCreatorsByDay(c.Request.Context(), analyticsDayEnds(startDate, endDate, location))
	if err != nil {
		logger.StdErr.Printf("monthly active event creators failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load monthly active event creators"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// analyticsDayEnds returns the inclusive per-day reporting instants from
// startDate through endDate at 23:59:59 in the client's fixed offset. The
// ascending slice is the day spine for one set-based analytics query.
func analyticsDayEnds(startDate, endDate time.Time, location *time.Location) []time.Time {
	dayEnds := make([]time.Time, 0)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		year, month, day := d.Date()
		dayEnds = append(dayEnds, time.Date(year, month, day, 23, 59, 59, 0, location))
	}
	return dayEnds
}

// @Summary Gets the daily count of monthly active event creators over a date range with more than x events
// @Tags analytics
// @Accept json
// @Produce json
// @Param startDate query string true "Start date (YYYY-MM-DD) for the range"
// @Param endDate query string true "End date (YYYY-MM-DD) for the range"
// @Param timezoneOffset query integer true "Client's timezone offset in minutes from UTC (e.g., -420 for UTC-7)"
// @Success 200 {array} object{date=string,count=int}
// @Failure 400 {object} object{error=string} "Invalid date format, range, or timezone offset"
// @Failure 500 {object} object{error=string} "Internal server error"
// @Router /analytics/monthly-active-event-creators-with-more-than-x-events [get]
func getMonthlyActiveEventCreatorsWithMoreThanXEvents(c *gin.Context) {
	startDateStr := c.Query("startDate")
	endDateStr := c.Query("endDate")
	timezoneOffsetStr := c.Query("timezoneOffset") // Get timezone offset param
	xStr := c.Query("x")

	if startDateStr == "" || endDateStr == "" || timezoneOffsetStr == "" || xStr == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "startDate, endDate, timezoneOffset, and x query parameters are required"})
		return
	}

	x, err := strconv.Atoi(xStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid x format. Must be an integer."})
		return
	}

	// Parse timezone offset
	timezoneOffset, err := strconv.Atoi(timezoneOffsetStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid timezoneOffset format. Must be an integer representing minutes."})
		return
	}
	// Convert offset minutes to a location (Go's time package uses seconds west of UTC)
	// Note: JS getTimezoneOffset() is positive for west, negative for east.
	// Go FixedZone expects seconds east of UTC. So, offset needs to be negated and converted to seconds.
	location := time.FixedZone("UserOffset", -timezoneOffset*60)

	layout := "2006-01-02" // YYYY-MM-DD
	startDate, err := time.Parse(layout, startDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid startDate format. Use YYYY-MM-DD"})
		return
	}
	endDate, err := time.Parse(layout, endDateStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid endDate format. Use YYYY-MM-DD"})
		return
	}

	if endDate.Before(startDate) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "endDate cannot be before startDate"})
		return
	}

	// Creator analytics read authoritative event storage, matching
	// the distinct-creator counting boundary. The whole range is one day-spine
	// query, and a failure fails the request instead of dropping days.
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		logger.StdErr.Panicln(err)
	}

	results, err := repository.CountDistinctMonthlyActiveEventCreatorsWithMoreThanXEventsByDay(
		c.Request.Context(), analyticsDayEnds(startDate, endDate, location), x)
	if err != nil {
		logger.StdErr.Printf("monthly active event creators with more than x events failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load monthly active event creators"})
		return
	}

	c.JSON(http.StatusOK, results)
}

// @Summary Gets the user by email
// @Tags analytics
// @Accept json
// @Produce json
// @Param email path string true "User email"
// @Success 200 {object} models.User
// @Router /analytics/user/{email} [get]
func getUserByEmail(c *gin.Context) {
	email := c.Param("email")
	user := accounts.UserByEmail(email)
	c.JSON(http.StatusOK, user)
}
