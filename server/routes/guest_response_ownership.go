package routes

import (
	"timeful/server/models"
	"timeful/server/respondents"
)

const guestOwnershipModeToken = "token"

func hasValidGuestName(name string) bool {
	return respondents.HasValidGuestName(name)
}

func shouldExposeGuestSignUpResponsePayload(_ string, response *models.SignUpResponse) bool {
	if response == nil {
		return true
	}

	if !response.UserId.IsZero() {
		return true
	}

	return hasValidGuestName(response.Name)
}
