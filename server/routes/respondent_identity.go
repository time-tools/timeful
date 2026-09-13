package routes

import (
	"timeful/server/models"
	"timeful/server/respondents"
)

func cloneUser(user *models.User) *models.User {
	if user == nil {
		return nil
	}

	clone := *user
	return &clone
}

func sanitizedResponseUser(user *models.User) *models.User {
	sanitized := cloneUser(user)
	if sanitized == nil {
		return nil
	}

	firstName, lastName, _ := respondents.SanitizeUserDisplayName(sanitized)
	sanitized.FirstName = firstName
	sanitized.LastName = lastName

	return sanitized
}

func canonicalGuestName(name string) string {
	validation := respondents.ValidateGuestName(name)
	if validation.Code != respondents.GuestNameValid {
		return ""
	}

	return validation.Name
}

func guestNameValidationErrorMessage(code respondents.GuestNameValidationCode) string {
	switch code {
	case respondents.GuestNameRequired:
		return "Guest name is required"
	case respondents.GuestNameInvalidFormatting:
		return "Guest name contains only unsupported formatting characters"
	case respondents.GuestNameAccountIDLike:
		return "Guest name cannot look like an account ID"
	case respondents.GuestNameTooLong:
		return "Guest name must be 100 characters or fewer"
	default:
		return "Guest name is invalid"
	}
}

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
