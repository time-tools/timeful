package microsoftgraph

import (
	"encoding/json"
	"fmt"

	"timeful/server/logger"
	"timeful/server/models"
	"timeful/server/services"
	"timeful/server/services/providerconfig"
)

type UserInfo struct {
	FirstName string `json:"givenName"`
	LastName  string `json:"surname"`
	Email     string `json:"mail"`
}

func GetUserInfo(user *models.User, calendarAuth *models.OAuth2CalendarAuth) UserInfo {
	response := services.CallApi(
		user,
		calendarAuth,
		"GET",
		fmt.Sprintf("%s/me?$select=givenName,surname,mail", providerconfig.MicrosoftGraphAPIBaseURL()),
		nil,
	)
	defer response.Body.Close()

	userResponse := struct {
		GivenName string `json:"givenName"`
		Surname   string `json:"surname"`
		Mail      string `json:"mail"`
	}{}

	if err := json.NewDecoder(response.Body).Decode(&userResponse); err != nil {
		logger.StdErr.Panicln(err)
	}

	return UserInfo{
		FirstName: userResponse.GivenName,
		LastName:  userResponse.Surname,
		Email:     userResponse.Mail,
	}
}
