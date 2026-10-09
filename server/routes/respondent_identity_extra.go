package routes

import (
	"timeful/server/models"
	"timeful/server/respondents"
)

// populateSignUpResponsePayloadIdentity resolves the response identity key and
// promotes the live account profile when one exists. liveUsers carries the
// batched account read keyed by platform identity UUID; a missing account
// (including a failed batch read) falls back to the stored response identity so
// the wire shape never gains an account shape it did not have before.
func populateSignUpResponsePayloadIdentity(response *models.SignUpResponse, liveUsers map[string]*models.User) (string, bool) {
	if response == nil {
		return "", false
	}

	if !response.UserId.IsZero() {
		lookupKey := response.UserId.String()
		liveUser, ok := liveUsers[lookupKey]
		if ok && liveUser != nil {
			response.User = sanitizedResponseUser(liveUser)
		} else {
			fallbackName := respondents.NormalizeGuestName(response.Name)
			response.User = &models.User{
				Id:        response.UserId,
				FirstName: fallbackName,
				Email:     response.Email,
			}
		}
		return lookupKey, true
	}

	name := canonicalGuestName(response.Name)
	if name == "" {
		return "", false
	}

	response.Name = name
	response.User = &models.User{
		FirstName: name,
		Email:     response.Email,
	}

	return name, true
}
