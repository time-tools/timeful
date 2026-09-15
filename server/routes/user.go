/* The /user group contains all the routes to get all the information about the currently signed in user */
package routes

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"timeful/server/accounts"
	"timeful/server/errs"
	"timeful/server/eventid"
	"timeful/server/logger"
	"timeful/server/middleware"
	"timeful/server/models"
	pgstore "timeful/server/postgres"
	"timeful/server/responses"
	"timeful/server/services/auth"
	"timeful/server/services/calendar"
	"timeful/server/services/contacts"
	"timeful/server/services/microsoftgraph"
	"timeful/server/utils"
)

func InitUser(router *gin.RouterGroup) {
	userRouter := router.Group("/user")
	userRouter.Use(middleware.AuthRequired())

	userRouter.GET("/profile", getProfile)
	userRouter.PATCH("/name", updateName)
	userRouter.PATCH("/calendar-options", updateCalendarOptions)
	userRouter.GET("/events", getEvents)
	userRouter.POST("/events/:eventId/set-folder", setEventFolder)
	userRouter.GET("/calendars", getCalendars)
	userRouter.POST("/add-google-calendar-account", addGoogleCalendarAccount)
	userRouter.POST("/add-apple-calendar-account", addAppleCalendarAccount)
	userRouter.POST("/add-outlook-calendar-account", addOutlookCalendarAccount)
	userRouter.POST("/add-ics-calendar-account", addICSCalendarAccount)
	userRouter.DELETE("/remove-calendar-account", removeCalendarAccount)
	userRouter.POST("/toggle-calendar", toggleCalendar)
	userRouter.POST("/toggle-sub-calendar", toggleSubCalendar)
	userRouter.GET("/searchContacts", searchContacts)
	userRouter.DELETE("", deleteUser)
}

// @Summary Gets the user's profile
// @Tags user
// @Produce json
// @Success 200 {object} models.User "A user profile object"
// @Router /user/profile [get]
func getProfile(c *gin.Context) {
	userInterface, _ := c.Get("authUser")
	user := userInterface.(*models.User)

	// The usage counter is authoritative.
	account := utils.GetAuthAccount(c)
	if account == nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
		return
	}
	repository, err := pgstore.DefaultRepository()
	if err != nil {
		logger.StdErr.Panicln(err)
	}
	// Sign-in activity is recorded in the daily log using the
	// authoritative account identifier and timezone offset.
	if err := repository.RecordDailyUserLogMembership(c.Request.Context(), account.PlatformIdentityID, account.TimezoneOffset); err != nil {
		logger.StdErr.Panicln(err)
	}

	c.JSON(http.StatusOK, user)
}

// @Summary Updates the user's name
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{firstName=string,lastName=string} true "Object containing the updated name"
// @Success 200
// @Router /user/name [patch]
func updateName(c *gin.Context) {
	payload := struct {
		FirstName string `json:"firstName" binding:"required"`
		LastName  string `json:"lastName" binding:"required"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	account := utils.GetAuthAccount(c)
	if account == nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
		return
	}
	account.FirstName = payload.FirstName
	account.LastName = payload.LastName
	account.HasCustomName = utils.TruePtr()

	// The profile is authoritative; this path does not write it.
	if err := accounts.UpdateProfile(c.Request.Context(), account); err != nil {
		logger.StdErr.Panicln(err)
	}

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Updates the user's calendar options
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{bufferTime=models.BufferTimeOptions,workingHours=models.WorkingHoursOptions} true "Object containing the updated options"
// @Success 200
// @Router /user/calendar-options [patch]
func updateCalendarOptions(c *gin.Context) {
	payload := struct {
		BufferTime   *models.BufferTimeOptions   `json:"bufferTime"`
		WorkingHours *models.WorkingHoursOptions `json:"workingHours"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	authUser := utils.GetAuthUser(c)

	// Set default values for calendar options if nil
	if authUser.CalendarOptions == nil {
		authUser.CalendarOptions = &models.CalendarOptions{
			BufferTime: models.BufferTimeOptions{
				Enabled: false,
				Time:    15,
			},
			WorkingHours: models.WorkingHoursOptions{
				Enabled:   false,
				StartTime: 9,
				EndTime:   17,
			},
		}
	}

	// Update calendar options
	if payload.BufferTime != nil {
		authUser.CalendarOptions.BufferTime = *payload.BufferTime
	}
	if payload.WorkingHours != nil {
		authUser.CalendarOptions.WorkingHours = *payload.WorkingHours
	}

	// Calendar preferences are authoritative.
	authAccount := utils.GetAuthAccount(c)
	if authAccount == nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
		return
	}
	if err := accounts.SaveCalendarPreferences(c.Request.Context(), authAccount.PlatformIdentityID, accounts.CalendarPreferences{
		PrimaryAccountKey: authUser.PrimaryAccountKey,
		TokenOrigin:       authUser.TokenOrigin,
		CalendarOptions:   authUser.CalendarOptions,
	}); err != nil {
		logger.StdErr.Panicln(err)
	}

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Gets all the user's events
// @Description Returns an array containing all the user's events
// @Tags user
// @Produce json
// @Success 200 {object} []models.Event
// @Router /user/events [get]
func getEvents(c *gin.Context) {
	user := utils.GetAuthUser(c)
	userId := user.Id

	repository := defaultRepository(c)
	if repository == nil {
		return
	}
	dashboardEvents, err := repository.ListDashboardEvents(c.Request.Context(), userId.String(), user.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.Error{Error: "failed-to-load-events"})
		return
	}
	result := make([]any, 0, len(dashboardEvents))
	for _, item := range dashboardEvents {
		payload, err := dashboardEvent(item.Event, item.Owned, userId.String(), item.Responded && item.Member)
		if err != nil {
			logger.StdErr.Panicln(err)
		}
		result = append(result, payload)
	}

	c.JSON(http.StatusOK, result)
}

// dashboardEvent renders an event in the dashboard wire
// shape. The canonical public identifier is exposed as both _id and shortId so
// the frontend opens the event without a store prefix and uses it as a stable
// list key. ownerId carries the account identifier only for owned events;
// responded-only events stay anonymous. Group entries carry the derived
// responded state the dashboard sets.
func dashboardEvent(event pgstore.Event, owned bool, platformIdentityID string, responded bool) (map[string]any, error) {
	value, err := eventModel(&event)
	if err != nil {
		return nil, err
	}
	value.ResponsesMap = nil
	value.HasResponded = nil
	encoded, err := value.MarshalAPIJSON()
	if err != nil {
		return nil, err
	}
	result := map[string]any{}
	if err := json.Unmarshal(encoded, &result); err != nil {
		return nil, err
	}
	result["_id"] = event.ShortID
	result["shortId"] = event.ShortID
	ownerID := models.ZeroUUID().String()
	if owned {
		if _, ok := models.ParseUUID(platformIdentityID); ok {
			ownerID = platformIdentityID
		}
	}
	result["ownerId"] = ownerID
	if event.Type == pgstore.EventTypeGroup {
		result["hasResponded"] = responded
	}
	return result, nil
}

// @Summary Sets the folder for the specified event
// @Tags user
// @Accept json
// @Produce json
// @Param eventId path string true "The ID of the event to set the folder for"
// @Param payload body object{folderId=string} true "The ID of the folder to set the event to"
// @Success 200
// @Router /user/events/{eventId}/set-folder [post]
func setEventFolder(c *gin.Context) {
	eventID := c.Param("eventId")
	if !eventid.Canonical(eventID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid event ID"})
		return
	}

	var body = struct {
		FolderId *string `json:"folderId"`
	}{}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session := sessions.Default(c)
	platformIdentityID, ok := session.Get("userId").(string)
	if !ok || platformIdentityID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	var folderId *string
	if body.FolderId != nil {
		if !validFolderID(*body.FolderId) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid folder ID"})
			return
		}
		folderId = body.FolderId
	}

	repository := defaultRepository(c)
	if repository == nil {
		return
	}

	event, err := repository.GetEventByShortID(c.Request.Context(), eventID)
	if errors.Is(err, pgx.ErrNoRows) || (err == nil && event.IsDeleted) {
		c.JSON(http.StatusNotFound, responses.Error{Error: errs.EventNotFound})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, responses.Error{Error: "failed-to-load-event"})
		return
	}
	eventIDValue := event.ID
	member := pgstore.FolderMember{EventID: &eventIDValue}

	err = repository.AssignEventToFolder(c.Request.Context(), platformIdentityID, folderId, member)
	if errors.Is(err, pgx.ErrNoRows) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Folder not found"})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add event to folder"})
		return
	}

	c.Status(http.StatusOK)
}

// @Summary Gets the user's calendar events
// @Description Gets the user's calendar events between "timeMin" and "timeMax"
// @Tags user
// @Produce json
// @Param timeMin query string true "Lower bound for event's start time to filter by"
// @Param timeMax query string true "Upper bound for event's end time to filter by"
// @Param accounts query string false "Comma separated list of accounts to fetch calendar events from"
// @Success 200 {object} map[string]calendar.CalendarEventsWithError
// @Router /user/calendars [get]
func getCalendars(c *gin.Context) {
	// Bind query parameters
	payload := struct {
		TimeMin  time.Time `form:"timeMin" binding:"required"`
		TimeMax  time.Time `form:"timeMax" binding:"required"`
		Accounts string    `form:"accounts"`
	}{}
	if err := c.Bind(&payload); err != nil {
		return
	}

	var requestedAccounts []string
	if len(payload.Accounts) == 0 {
		requestedAccounts = make([]string, 0)
	} else {
		requestedAccounts = utils.ParseArrayQueryParam(payload.Accounts)
	}
	accountsSet := utils.ArrayToSet(requestedAccounts)
	user := utils.GetAuthUser(c)

	calendarEvents, editedCalendarAccounts := calendar.GetUsersCalendarEvents(user, accountsSet, payload.TimeMin, payload.TimeMax)

	if editedCalendarAccounts {
		// The provider refresh may add or drop sub-calendars; persist the
		// reconciled set.
		authAccount := utils.GetAuthAccount(c)
		if authAccount != nil {
			for calendarAccountKey, calendarAccount := range user.CalendarAccounts {
				if calendarAccount.SubCalendars == nil {
					continue
				}
				if err := accounts.SyncCalendarSubCalendars(c.Request.Context(), authAccount.PlatformIdentityID, calendarAccountKey, *calendarAccount.SubCalendars); err != nil {
					logger.StdErr.Panicln(err)
				}
			}
		}
	}

	c.JSON(http.StatusOK, calendarEvents)
}

// @Summary Adds a new calendar account
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{code=string,scope=string} true "Object containing the Google authorization code and scope"
// @Success 200
// @Router /user/add-google-calendar-account [post]
func addGoogleCalendarAccount(c *gin.Context) {
	payload := struct {
		Code  string `json:"code" binding:"required"`
		Scope string `json:"scope" binding:"required"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	// Get tokens
	tokens := auth.GetTokensFromAuthCode(payload.Code, payload.Scope, utils.GetOrigin(c), models.GoogleCalendarType)

	// Get user info from JWT
	claims := utils.ParseJWT(tokens.IdToken)
	email, _ := claims.GetStr("email")
	picture, _ := claims.GetStr("picture")

	// Get access token expire time
	accessTokenExpireDate := utils.GetAccessTokenExpireDate(tokens.ExpiresIn)

	calendarAuth := &models.OAuth2CalendarAuth{
		AccessToken:           tokens.AccessToken,
		AccessTokenExpireDate: models.NewDateTimeFromTime(accessTokenExpireDate),
		RefreshToken:          tokens.RefreshToken,
	}

	addCalendarAccount(c, addCalendarAccountArgs{
		calendarType:       models.GoogleCalendarType,
		oAuth2CalendarAuth: calendarAuth,
		email:              email,
		picture:            picture,
	})

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Adds an apple calendar account
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{email=string,password=string} true "Object containing the email and app password of the apple account"
// @Success 200
// @Router /user/add-apple-calendar-account [post]
func addAppleCalendarAccount(c *gin.Context) {
	payload := struct {
		Email    string `json:"email" binding:"required"`
		Password string `json:"password" binding:"required"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	// The repository encrypts the app password at rest, so the route
	// passes plaintext and the provider consumes plaintext.
	auth := &models.AppleCalendarAuth{
		Email:    payload.Email,
		Password: payload.Password,
	}

	// Check if the provided credentials are valid
	calendarProvider := calendar.AppleCalendar{
		AppleCalendarAuth: *auth,
	}
	_, err := calendarProvider.GetCalendarList()
	if err != nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.InvalidCredentials})
		return
	}

	addCalendarAccount(c, addCalendarAccountArgs{
		calendarType:      models.AppleCalendarType,
		appleCalendarAuth: auth,
		email:             payload.Email,
		picture:           "",
	})

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Adds a new outlook calendar account
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{code=string,scope=string} true "Object containing the Outlook authorization code and scope"
// @Success 200
// @Router /user/add-outlook-calendar-account [post]
func addOutlookCalendarAccount(c *gin.Context) {
	payload := struct {
		Code  string `json:"code" binding:"required"`
		Scope string `json:"scope" binding:"required"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	// Get auth user
	authUser := utils.GetAuthUser(c)

	// Get tokens
	tokens := auth.GetTokensFromAuthCode(payload.Code, payload.Scope, utils.GetOrigin(c), models.OutlookCalendarType)

	// Get access token expire time
	accessTokenExpireDate := utils.GetAccessTokenExpireDate(tokens.ExpiresIn)

	// Construct calendarAuth object
	calendarAuth := &models.OAuth2CalendarAuth{
		AccessToken:           tokens.AccessToken,
		AccessTokenExpireDate: models.NewDateTimeFromTime(accessTokenExpireDate),
		RefreshToken:          tokens.RefreshToken,
		Scope:                 payload.Scope,
	}

	// Get user info
	userInfo := microsoftgraph.GetUserInfo(authUser, calendarAuth)

	addCalendarAccount(c, addCalendarAccountArgs{
		calendarType:       models.OutlookCalendarType,
		oAuth2CalendarAuth: calendarAuth,
		email:              userInfo.Email,
		picture:            "",
	})

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Adds an ICS calendar account
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{feedUrl=string,label=string} true "Object containing the feed URL and label of the ICS calendar"
// @Success 200
// @Router /user/add-ics-calendar-account [post]
func addICSCalendarAccount(c *gin.Context) {
	payload := struct {
		FeedURL string `json:"feedUrl" binding:"required"`
		Label   string `json:"label" binding:"required"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	auth := &models.ICSCalendarAuth{
		FeedURL: payload.FeedURL,
		Label:   payload.Label,
	}

	// Check if the provided feed URL is reachable
	calendarProvider := calendar.ICSCalendar{
		ICSCalendarAuth: *auth,
	}
	_, err := calendarProvider.GetCalendarList()
	if err != nil {
		c.JSON(http.StatusBadRequest, responses.Error{Error: "Invalid ICS feed URL"})
		return
	}

	addCalendarAccount(c, addCalendarAccountArgs{
		calendarType:    models.ICSCalendarType,
		icsCalendarAuth: auth,
		// ICS feeds don't have an email, so we use the label instead
		email:   payload.Label,
		picture: "",
	})

	c.JSON(http.StatusOK, gin.H{})
}

// Implements the shared functionality for adding a calendar account
type addCalendarAccountArgs struct {
	calendarType       models.CalendarType
	oAuth2CalendarAuth *models.OAuth2CalendarAuth
	appleCalendarAuth  *models.AppleCalendarAuth
	icsCalendarAuth    *models.ICSCalendarAuth
	email              string
	picture            string
}

func addCalendarAccount(c *gin.Context, args addCalendarAccountArgs) {
	// Get auth user
	authUser := utils.GetAuthUser(c)

	ident := args.email
	if args.calendarType != models.ICSCalendarType {
		ident = utils.NormalizeEmail(args.email)
	} else {
		ident = strings.TrimSpace(args.email)
	}

	// Create calendar account object
	calendarAccount := models.CalendarAccount{
		CalendarType: args.calendarType,

		Email:   ident,
		Picture: args.picture,
		Enabled: utils.TruePtr(), // Workaround to pass a boolean pointer
	}
	switch args.calendarType {
	case models.GoogleCalendarType:
		calendarAccount.OAuth2CalendarAuth = args.oAuth2CalendarAuth
	case models.OutlookCalendarType:
		calendarAccount.OAuth2CalendarAuth = args.oAuth2CalendarAuth
	case models.AppleCalendarType:
		calendarAccount.AppleCalendarAuth = args.appleCalendarAuth
	case models.ICSCalendarType:
		calendarAccount.ICSCalendarAuth = args.icsCalendarAuth
	}
	canonicalKey := utils.GetCalendarAccountKey(ident, args.calendarType)

	// Set subcalendars map based on whether calendar account already exists
	var oldForSub *models.CalendarAccount
	if acc, ok := authUser.CalendarAccounts[canonicalKey]; ok {
		oldForSub = &acc
	}
	if oldForSub != nil && oldForSub.SubCalendars != nil {
		calendarAccount.SubCalendars = oldForSub.SubCalendars
	} else {
		subCalendars, err := calendar.GetCalendarProvider(calendarAccount).GetCalendarList()
		if err == nil {
			calendarAccount.SubCalendars = &subCalendars
		}
	}

	if authUser.CalendarAccounts == nil {
		authUser.CalendarAccounts = make(map[string]models.CalendarAccount)
	}
	authUser.CalendarAccounts[canonicalKey] = calendarAccount

	// Calendar connections are authoritative; the profile is never
	// written from this path.
	authAccount := utils.GetAuthAccount(c)
	if authAccount == nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
		return
	}
	if err := accounts.SaveCalendarAccount(c.Request.Context(), authAccount.PlatformIdentityID, canonicalKey, calendarAccount); err != nil {
		logger.StdErr.Panicln(err)
	}
}

// @Summary Removes an existing calendar account
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{email=string,calendarType=models.CalendarType} true "Object containing the email + type of the calendar account to remove"
// @Success 200
// @Router /user/remove-calendar-account [delete]
func removeCalendarAccount(c *gin.Context) {
	payload := struct {
		Email        string              `json:"email" binding:"required"`
		CalendarType models.CalendarType `json:"calendarType" binding:"required"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		return
	}

	calendarAccountKey := utils.GetCalendarAccountKey(payload.Email, payload.CalendarType)

	// Calendar connections are authoritative.
	authAccount := utils.GetAuthAccount(c)
	if authAccount == nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
		return
	}
	if err := accounts.DeleteCalendarAccount(c.Request.Context(), authAccount.PlatformIdentityID, calendarAccountKey); err != nil {
		logger.StdErr.Panicln(err)
	}

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Toggles whether the specified calendar is enabled or disabled for the user
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{calendarAccountKey=string,enabled=bool} true "Email of calendar account and whether to enable it"
// @Success 200
// @Router /user/toggle-calendar [post]
func toggleCalendar(c *gin.Context) {
	payload := struct {
		Email        string              `json:"email" binding:"required"`
		CalendarType models.CalendarType `json:"calendarType" binding:"required"`
		Enabled      *bool               `json:"enabled" binding:"required"`
	}{}
	if err := c.Bind(&payload); err != nil {
		logger.StdErr.Panicln(err)
		return
	}

	// Update enabled status for the specified account
	authUser := utils.GetAuthUser(c)
	calendarAccountKey := utils.GetCalendarAccountKey(payload.Email, payload.CalendarType)
	if _, ok := authUser.CalendarAccounts[calendarAccountKey]; ok {
		authAccount := utils.GetAuthAccount(c)
		if authAccount == nil {
			c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
			return
		}
		if err := accounts.SetCalendarAccountEnabled(c.Request.Context(), authAccount.PlatformIdentityID, calendarAccountKey, *payload.Enabled); err != nil {
			logger.StdErr.Panicln(err)
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Toggles whether the specified sub-calendar is enabled or disabled for the user
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{calendarAccountKey=string,subCalendarId=string,enabled=bool} true "Email of calendar account, the sub calendar id, and whether to enable it"
// @Success 200
// @Router /user/toggle-sub-calendar [post]
func toggleSubCalendar(c *gin.Context) {
	payload := struct {
		Email         string              `json:"email" binding:"required"`
		CalendarType  models.CalendarType `json:"calendarType" binding:"required"`
		SubCalendarId string              `json:"subCalendarId" binding:"required"`
		Enabled       *bool               `json:"enabled" binding:"required"`
	}{}
	if err := c.Bind(&payload); err != nil {
		logger.StdErr.Panicln(err)
		return
	}

	// Update enabled status for the specified sub calendar
	authUser := utils.GetAuthUser(c)
	calendarAccountKey := utils.GetCalendarAccountKey(payload.Email, payload.CalendarType)
	if calendarAccount, ok := authUser.CalendarAccounts[calendarAccountKey]; ok && calendarAccount.SubCalendars != nil {
		if _, ok := (*calendarAccount.SubCalendars)[payload.SubCalendarId]; ok {
			authAccount := utils.GetAuthAccount(c)
			if authAccount == nil {
				c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
				return
			}
			if err := accounts.SetSubCalendarEnabled(c.Request.Context(), authAccount.PlatformIdentityID, calendarAccountKey, payload.SubCalendarId, *payload.Enabled); err != nil {
				logger.StdErr.Panicln(err)
				return
			}
		}
	}

	c.JSON(http.StatusOK, gin.H{})
}

// @Summary Searches the user's contacts based on the given query
// @Tags user
// @Produce json
// @Param query query string true "Query to search for"
// @Success 200 {object} []models.User
// @Router /user/searchContacts [get]
func searchContacts(c *gin.Context) {
	// Bind query parameters
	payload := struct {
		Query string `form:"query"`
	}{}
	if err := c.Bind(&payload); err != nil {
		return
	}

	userInterface, _ := c.Get("authUser")
	user := userInterface.(*models.User)

	contacts, googleError := contacts.SearchContacts(user, payload.Query)
	if googleError != nil {
		c.JSON(googleError.Code, responses.Error{Error: *googleError})
		return
	}

	c.JSON(http.StatusOK, contacts)
}

// @Summary Deletes the currently signed in user
// @Description Requires the account email address as confirmation. Deletion is permanent and immediate: the account profile, platform identity, calendar connections, and historical user logs are removed, and events the account organized survive with ownership released.
// @Tags user
// @Accept json
// @Produce json
// @Param payload body object{email=string} true "The account email address that must match the signed-in account"
// @Success 200
// @Failure 400 {object} responses.Error "The supplied email does not match the account"
// @Router /user [delete]
func deleteUser(c *gin.Context) {
	payload := struct {
		Email string `json:"email"`
	}{}
	if err := c.BindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, responses.Error{Error: errs.AccountEmailMismatch})
		return
	}

	account := utils.GetAuthAccount(c)
	if account == nil {
		c.JSON(http.StatusUnauthorized, responses.Error{Error: errs.UserDoesNotExist})
		return
	}
	if !strings.EqualFold(strings.TrimSpace(payload.Email), strings.TrimSpace(account.Email)) {
		c.JSON(http.StatusBadRequest, responses.Error{Error: errs.AccountEmailMismatch})
		return
	}

	// The deletion transaction removes the account authority,
	// platform identity, calendar connections, responses, folders, and
	// daily-log membership. The session is cleared only after the whole unit
	// succeeds, so a failure leaves the visitor signed in and able to retry.
	if err := accounts.DeleteAccount(c.Request.Context(), account.PlatformIdentityID); err != nil {
		log.Printf("account deletion failed for %s: %v", account.PlatformIdentityID, err)
		c.JSON(http.StatusInternalServerError, responses.Error{Error: "account-deletion-failed"})
		return
	}

	session := sessions.Default(c)
	session.Delete("userId")
	if err := session.Save(); err != nil {
		logger.StdErr.Panicln(err)
	}

	c.JSON(http.StatusOK, gin.H{})
}
